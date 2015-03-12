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

func createCipher(password string, iv []byte) (cipher.Stream, error) {
	passwordHash := sha256.Sum256(bytes.NewBufferString(password).Bytes())

	blockCipher, err := aes.NewCipher(passwordHash[:])
	if err != nil {
		return nil, err
	}

	streamCipher := cipher.NewOFB(blockCipher, iv)
	return streamCipher, nil
}

func encryptAES256(plaintext []byte, username, password string) ([]byte, error) {
	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		log.Printf("rand.Read: %v", err)
		return nil, err
	}

	streamCipher, err := createCipher(password, iv)
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

func decryptAES256(ciphertext []byte, userToPassword map[string]string) ([]byte, error) {
	buf := bytes.NewBuffer(ciphertext)
	userLen := int(binary.BigEndian.Uint16(buf.Next(2)))
	if 42+userLen >= buf.Len() {
		return nil, fmt.Errorf("invalid username length %d", userLen)
	}
	user := bytes.NewBuffer(buf.Next(userLen)).String()

	password, ok := userToPassword[user]
	if !ok {
		return nil, fmt.Errorf("no such user: %q", user)
	}

	iv := make([]byte, 16)
	if n, err := buf.Read(iv); n != 16 || err != nil {
		return nil, fmt.Errorf("reading IV failed: %v", err)
	}

	streamCipher, err := createCipher(password, iv)
	if err != nil {
		return nil, err
	}

	r := &cipher.StreamReader{S: streamCipher, R: buf}

	plaintext := make([]byte, buf.Len())
	if n, err := r.Read(plaintext); n != len(plaintext) || err != nil {
		return nil, fmt.Errorf("decryption failure: got (%d, %v), want (%d, nil)", n, err, len(plaintext))
	}

	checksumWant := plaintext[:20]
	plaintext = plaintext[20:]
	checksumGot := sha1.Sum(plaintext)

	if !bytes.Equal(checksumGot[:], checksumWant[:]) {
		return nil, errors.New("checksum mismatch")
	}

	return plaintext, nil
}
