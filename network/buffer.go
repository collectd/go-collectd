package network

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"math"
	"sync"
	"time"

	"collectd.org/api"
)

const (
	dsTypeGauge  = 1
	dsTypeDerive = 2
)

const (
	typeHost           = 0x0000
	typeTime           = 0x0001
	typeTimeHR         = 0x0008
	typePlugin         = 0x0002
	typePluginInstance = 0x0003
	typeType           = 0x0004
	typeTypeInstance   = 0x0005
	typeValues         = 0x0006
	typeInterval       = 0x0007
	typeIntervalHR     = 0x0009
	typeSignSHA256     = 0x0200
	typeEncryptAES256  = 0x0210
)

const DefaultBufferSize = 1452

var errNotEnoughSpace = errors.New("not enough space")

// Buffer contains the binary representation of multiple ValueLists and state
// optimally write the next ValueList.
type Buffer struct {
	lock               *sync.Mutex
	buffer             *bytes.Buffer
	output             io.Writer
	state              api.ValueList
	size               int
	username, password string
	encrypt            bool
}

// NewBuffer initializes a new Buffer.
func NewBuffer(w io.Writer) *Buffer {
	return &Buffer{
		lock:   new(sync.Mutex),
		buffer: new(bytes.Buffer),
		output: w,
		size:   DefaultBufferSize,
	}
}

// NewBufferSigned initializes a new Buffer which is cryptographically signed.
func NewBufferSigned(w io.Writer, username, password string) *Buffer {
	encoded := bytes.NewBufferString(username)
	sigSize := 36 + encoded.Len()

	return &Buffer{
		lock:     new(sync.Mutex),
		buffer:   new(bytes.Buffer),
		output:   w,
		size:     DefaultBufferSize - sigSize,
		username: username,
		password: password,
		encrypt:  false,
	}
}

// NewBufferEncrypted initializes a new Buffer which is encrypted.
func NewBufferEncrypted(w io.Writer, username, password string) *Buffer {
	encoded := bytes.NewBufferString(username)
	sigSize := 42 + encoded.Len()

	return &Buffer{
		lock:     new(sync.Mutex),
		buffer:   new(bytes.Buffer),
		output:   w,
		size:     DefaultBufferSize - sigSize,
		username: username,
		password: password,
		encrypt:  true,
	}
}

// Free returns the number of bytes still available in the buffer.
func (b *Buffer) Free() int {
	used := b.buffer.Len()
	if b.size < used {
		return 0
	}
	return b.size - used
}

// Flush writes all data currently in the buffer to the associated io.Writer.
func (b *Buffer) Flush() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.flush()
}

// WriteValueList adds a ValueList to the network buffer.
func (b *Buffer) WriteValueList(vl api.ValueList) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	l := b.buffer.Len()

	if err := b.writeValueList(vl); err != nil {
		// Buffer is empty; we can't flush and retry.
		if l == 0 {
			return err
		}
	} else {
		return nil
	}

	// flush
	b.buffer.Truncate(l)
	if err := b.flush(); err != nil {
		return err
	}

	// retry
	return b.writeValueList(vl)
}

func (b *Buffer) writeValueList(vl api.ValueList) error {
	if err := b.writeIdentifier(vl.Identifier); err != nil {
		return err
	}

	if err := b.writeTime(vl.Time); err != nil {
		return err
	}

	if err := b.writeInterval(vl.Interval); err != nil {
		return err
	}

	if err := b.writeValues(vl.Values); err != nil {
		return err
	}

	return nil
}

func (b *Buffer) writeIdentifier(id api.Identifier) error {
	if id.Host != b.state.Host {
		if err := b.writeString(typeHost, id.Host); err != nil {
			return err
		}
		b.state.Host = id.Host
	}
	if id.Plugin != b.state.Plugin {
		if err := b.writeString(typePlugin, id.Plugin); err != nil {
			return err
		}
		b.state.Plugin = id.Plugin
	}
	if id.PluginInstance != b.state.PluginInstance {
		if err := b.writeString(typePluginInstance, id.PluginInstance); err != nil {
			return err
		}
		b.state.PluginInstance = id.PluginInstance
	}
	if id.Type != b.state.Type {
		if err := b.writeString(typeType, id.Type); err != nil {
			return err
		}
		b.state.Type = id.Type
	}
	if id.TypeInstance != b.state.TypeInstance {
		if err := b.writeString(typeTypeInstance, id.TypeInstance); err != nil {
			return err
		}
		b.state.TypeInstance = id.TypeInstance
	}

	return nil
}

func (b *Buffer) writeTime(t time.Time) error {
	if b.state.Time == t {
		return nil
	}
	b.state.Time = t

	return b.writeInt(typeTimeHR, api.Cdtime(t))
}

func (b *Buffer) writeInterval(d time.Duration) error {
	if b.state.Interval == d {
		return nil
	}
	b.state.Interval = d

	return b.writeInt(typeIntervalHR, api.CdtimeDuration(d))
}

func (b *Buffer) writeValues(values []api.Value) error {
	size := 6 + 9*len(values)
	if size > b.Free() {
		return errNotEnoughSpace
	}

	binary.Write(b.buffer, binary.BigEndian, uint16(typeValues))
	binary.Write(b.buffer, binary.BigEndian, uint16(size))
	binary.Write(b.buffer, binary.BigEndian, uint16(len(values)))

	for _, v := range values {
		switch v.(type) {
		case api.Gauge:
			binary.Write(b.buffer, binary.BigEndian, uint8(dsTypeGauge))
		case api.Derive:
			binary.Write(b.buffer, binary.BigEndian, uint8(dsTypeDerive))
		default:
			panic("unexpected type")
		}
	}

	for _, v := range values {
		switch v := v.(type) {
		case api.Gauge:
			if math.IsNaN(float64(v)) {
				b.buffer.Write([]byte{0, 0, 0, 0, 0, 0, 0xf8, 0x7f})
			} else {
				// sic: floats are encoded in little endian.
				binary.Write(b.buffer, binary.LittleEndian, float64(v))
			}
		case api.Derive:
			binary.Write(b.buffer, binary.BigEndian, int64(v))
		default:
			panic("unexpected type")
		}
	}

	return nil
}

func (b *Buffer) writeString(typ uint16, s string) error {
	encoded := bytes.NewBufferString(s)
	encoded.Write([]byte{0})

	// Because s is a Unicode string, encoded.Len() may be larger than
	// len(s).
	size := 4 + encoded.Len()
	if size > b.Free() {
		return errNotEnoughSpace
	}

	binary.Write(b.buffer, binary.BigEndian, typ)
	binary.Write(b.buffer, binary.BigEndian, uint16(size))
	b.buffer.Write(encoded.Bytes())

	return nil
}

func (b *Buffer) writeInt(typ uint16, n uint64) error {
	size := 12
	if size > b.Free() {
		return errNotEnoughSpace
	}

	binary.Write(b.buffer, binary.BigEndian, typ)
	binary.Write(b.buffer, binary.BigEndian, uint16(size))
	binary.Write(b.buffer, binary.BigEndian, n)

	return nil
}

func (b *Buffer) flush() error {
	if b.buffer.Len() == 0 {
		return nil
	}

	buf := make([]byte, b.buffer.Len())
	if _, err := b.buffer.Read(buf); err != nil {
		return err
	}

	if b.username != "" && b.password != "" {
		if b.encrypt {
			var err error
			if buf, err = encrypt(buf, b.username, b.password); err != nil {
				return err
			}
		} else {
			buf = sign(buf, b.username, b.password)
		}
	}

	if _, err := b.output.Write(buf); err != nil {
		return err
	}

	// zero state
	b.state = api.ValueList{}
	return nil
}

func sign(payload []byte, username, password string) []byte {
	mac := hmac.New(sha256.New, bytes.NewBufferString(password).Bytes())

	usernameBuffer := bytes.NewBufferString(username)

	size := uint16(36 + usernameBuffer.Len())

	mac.Write(usernameBuffer.Bytes())
	mac.Write(payload)

	out := new(bytes.Buffer)
	binary.Write(out, binary.BigEndian, uint16(typeSignSHA256))
	binary.Write(out, binary.BigEndian, size)
	out.Write(mac.Sum(nil))
	out.Write(usernameBuffer.Bytes())
	out.Write(payload)

	return out.Bytes()
}

func createCipher(password string) (cipher.Stream, []byte, error) {
	passwordHash := sha256.Sum256(bytes.NewBufferString(password).Bytes())

	blockCipher, err := aes.NewCipher(passwordHash[:])
	if err != nil {
		return nil, nil, err
	}

	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		log.Printf("rand.Read: %v", err)
		return nil, nil, err
	}

	streamCipher := cipher.NewOFB(blockCipher, iv[:])
	return streamCipher, iv, nil
}

func encrypt(plaintext []byte, username, password string) ([]byte, error) {
	streamCipher, iv, err := createCipher(password)
	if err != nil {
		return nil, err
	}

	usernameBuffer := bytes.NewBufferString(username)

	size := uint16(42 + usernameBuffer.Len() + len(plaintext))

	checksum := sha1.Sum(plaintext)

	out := new(bytes.Buffer)
	binary.Write(out, binary.BigEndian, uint16(typeEncryptAES256))
	binary.Write(out, binary.BigEndian, size)
	binary.Write(out, binary.BigEndian, uint16(usernameBuffer.Len()))
	out.Write(usernameBuffer.Bytes())
	out.Write(iv)

	w := &cipher.StreamWriter{S: streamCipher, W: out}
	w.Write(checksum[:])
	w.Write(plaintext)

	return out.Bytes(), nil
}
