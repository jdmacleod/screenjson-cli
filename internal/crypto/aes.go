// Package crypto provides AES-256 encryption/decryption for screenplay content.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// MinKeyLength is the minimum encryption key length.
const MinKeyLength = 10

// Encoding types supported for encrypted output.
const (
	EncodingHex    = "hex"
	EncodingBase64 = "base64"
)

// Encryptor handles AES-256 encryption.
type Encryptor struct {
	key      []byte
	encoding string
}

// NewEncryptor creates a new AES-256 encryptor from a passphrase.
func NewEncryptor(passphrase string) (*Encryptor, error) {
	if len(passphrase) < MinKeyLength {
		return nil, fmt.Errorf("encryption key must be at least %d characters", MinKeyLength)
	}

	// Derive 256-bit key from passphrase using SHA-256
	hash := sha256.Sum256([]byte(passphrase))
	
	return &Encryptor{
		key:      hash[:],
		encoding: EncodingHex,
	}, nil
}

// WithEncoding sets the output encoding.
func (e *Encryptor) WithEncoding(encoding string) *Encryptor {
	e.encoding = encoding
	return e
}

// Encrypt encrypts plaintext using AES-256-CTR.
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate random IV
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	// Encode
	switch e.encoding {
	case EncodingBase64:
		return base64.StdEncoding.EncodeToString(ciphertext), nil
	default:
		return hex.EncodeToString(ciphertext), nil
	}
}

// Decryptor handles AES-256 decryption.
type Decryptor struct {
	key      []byte
	encoding string
}

// NewDecryptor creates a new AES-256 decryptor from a passphrase.
func NewDecryptor(passphrase string) (*Decryptor, error) {
	if len(passphrase) < MinKeyLength {
		return nil, fmt.Errorf("decryption key must be at least %d characters", MinKeyLength)
	}

	hash := sha256.Sum256([]byte(passphrase))
	
	return &Decryptor{
		key:      hash[:],
		encoding: EncodingHex,
	}, nil
}

// WithEncoding sets the input encoding.
func (d *Decryptor) WithEncoding(encoding string) *Decryptor {
	d.encoding = encoding
	return d
}

// Decrypt decrypts ciphertext using AES-256-CTR.
func (d *Decryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode
	var data []byte
	var err error
	switch d.encoding {
	case EncodingBase64:
		data, err = base64.StdEncoding.DecodeString(ciphertext)
	default:
		data, err = hex.DecodeString(ciphertext)
	}
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	if len(data) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(d.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Extract IV and ciphertext
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	// Decrypt
	plaintext := make([]byte, len(data))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, data)

	return string(plaintext), nil
}

// TestKey tests if a key can decrypt a sample ciphertext.
func TestKey(passphrase, ciphertext, encoding string) bool {
	dec, err := NewDecryptor(passphrase)
	if err != nil {
		return false
	}
	dec.WithEncoding(encoding)
	
	_, err = dec.Decrypt(ciphertext)
	return err == nil
}
