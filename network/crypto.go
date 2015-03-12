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
	"fmt"
	"log"
)

func signSHA256(payload []byte, username, password string) []byte {
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

func verifySHA256(part, payload []byte, userToPassword map[string]string) (bool, error) {
	if len(part) <= 32 {
		return false, fmt.Errorf("part too small (%d bytes)", len(part))
	}

	hash := part[:32]
	user := bytes.NewBuffer(part[32:]).String()

	password, ok := userToPassword[user]
	if !ok {
		return false, fmt.Errorf("no such user: %q", user)
	}

	mac := hmac.New(sha256.New, bytes.NewBufferString(password).Bytes())

	mac.Write(part[32:])
	mac.Write(payload)

	return bytes.Equal(hash, mac.Sum(nil)), nil
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
