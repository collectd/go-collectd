package network

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"io"
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
)

const DefaultBufferSize = 1452

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

// Flush writes all data currently in the buffer to the associated io.Writer.
func (b *Buffer) Flush() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.flush(b.buffer.Len())
}

// WriteValueList adds a ValueList to the network buffer.
func (b *Buffer) WriteValueList(vl api.ValueList) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	origLen := b.buffer.Len()

	b.writeIdentifier(vl.Identifier)
	b.writeTime(vl.Time)
	b.writeInterval(vl.Interval)
	b.writeValues(vl.Values)

	if b.buffer.Len() >= b.size {
		if err := b.flush(origLen); err != nil {
			return err
		}
	}

	return nil
}

func (b *Buffer) writeIdentifier(id api.Identifier) {
	if id.Host != b.state.Host {
		b.writeString(typeHost, id.Host)
		b.state.Host = id.Host
	}
	if id.Plugin != b.state.Plugin {
		b.writeString(typePlugin, id.Plugin)
		b.state.Plugin = id.Plugin
	}
	if id.PluginInstance != b.state.PluginInstance {
		b.writeString(typePluginInstance, id.PluginInstance)
		b.state.PluginInstance = id.PluginInstance
	}
	if id.Type != b.state.Type {
		b.writeString(typeType, id.Type)
		b.state.Type = id.Type
	}
	if id.TypeInstance != b.state.TypeInstance {
		b.writeString(typeTypeInstance, id.TypeInstance)
		b.state.TypeInstance = id.TypeInstance
	}
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
	size := uint16(6 + 9*len(values))

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
	size := uint16(4 + encoded.Len())

	binary.Write(b.buffer, binary.BigEndian, typ)
	binary.Write(b.buffer, binary.BigEndian, size)
	b.buffer.Write(encoded.Bytes())

	return nil
}

func (b *Buffer) writeInt(typ uint16, n uint64) error {
	binary.Write(b.buffer, binary.BigEndian, typ)
	binary.Write(b.buffer, binary.BigEndian, uint16(12))
	binary.Write(b.buffer, binary.BigEndian, n)

	return nil
}

func (b *Buffer) flush(n int) error {
	if n == 0 {
		return nil
	}

	buf := make([]byte, n)
	if b.username != "" && b.password != "" {
		if b.encrypt {
			// TODO
		} else {
			buf = sign(buf, b.username, b.password)
		}
	}

	if _, err := b.buffer.Read(buf); err != nil {
		return err
	}

	if _, err := b.output.Write(buf); err != nil {
		return err
	}

	return nil
}

func sign(payload []byte, username, password string) []byte {
	key := bytes.NewBufferString(password)
	mac := hmac.New(sha256.New, key.Bytes())

	signedData := bytes.NewBufferString(username)
	signedData.Write(payload)

	size := uint16(36 + signedData.Len())

	signedData.WriteTo(mac)

	out := new(bytes.Buffer)
	binary.Write(out, binary.BigEndian, uint16(typeSignSHA256))
	binary.Write(out, binary.BigEndian, size)
	out.Write(mac.Sum(nil))
	out.WriteString(username)
	out.Write(payload)

	return out.Bytes()
}
