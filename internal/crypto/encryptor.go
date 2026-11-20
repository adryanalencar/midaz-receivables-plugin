package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKeySize     = errors.New("crypto: encryption key must be exactly 32 bytes (256-bit)")
	ErrCiphertextTooShort = errors.New("crypto: ciphertext payload too short to contain valid nonce and tag")
	ErrDecryptionFailed   = errors.New("crypto: decryption failed, payload corrupted or authentication tag mismatch")
)

// Encryptor defines the cryptographic contract for credentials encryption at rest.
type Encryptor interface {
	Encrypt(plaintext []byte) (string, error)
	Decrypt(ciphertextBase64 string) ([]byte, error)
}

// GCMEncryptor implements AES-256-GCM authenticated encryption with associated data.
type GCMEncryptor struct {
	key []byte
}

// NewGCMEncryptor initializes an Encryptor using a 32-byte key (either raw 32-char string or 64-char hex string).
func NewGCMEncryptor(keyInput string) (*GCMEncryptor, error) {
	var key []byte

	// Check if key is provided as 64-character hexadecimal
	if len(keyInput) == 64 {
		decoded, err := hex.DecodeString(keyInput)
		if err == nil && len(decoded) == 32 {
			key = decoded
		}
	}

	// If not hex, check if raw string is 32 bytes
	if key == nil {
		if len(keyInput) == 32 {
			key = []byte(keyInput)
		} else {
			return nil, fmt.Errorf("%w: received %d bytes", ErrInvalidKeySize, len(keyInput))
		}
	}

	return &GCMEncryptor{key: key}, nil
}

// Encrypt encrypts arbitrary plaintext using AES-256-GCM with a newly generated 12-byte random nonce.
// Returns a Base64-encoded string formatted as: Nonce (12 bytes) || Ciphertext || GCM Tag (16 bytes).
func (g *GCMEncryptor) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(g.key)
	if err != nil {
		return "", fmt.Errorf("crypto: failed to create AES cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: failed to wrap GCM AEAD: %w", err)
	}

	// Generate 12-byte cryptographically secure random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: failed to generate random nonce: %w", err)
	}

	// Seal appends the ciphertext and 16-byte authentication tag to nonce
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)

	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt unpacks the Base64 ciphertext, extracts the 12-byte nonce, and decrypts/verifies with AES-256-GCM.
func (g *GCMEncryptor) Decrypt(ciphertextBase64 string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("crypto: invalid base64 encoding: %w", err)
	}

	block, err := aes.NewCipher(g.key)
	if err != nil {
		return nil, fmt.Errorf("crypto: failed to create AES cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: failed to wrap GCM AEAD: %w", err)
	}

	nonceSize := gcm.NonceSize()
	// Minimum length: NonceSize (12 bytes) + Overhead/Tag (16 bytes)
	if len(raw) < nonceSize+gcm.Overhead() {
		return nil, ErrCiphertextTooShort
	}

	nonce := raw[:nonceSize]
	ciphertextWithTag := raw[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}
