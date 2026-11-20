package crypto

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGCMEncryptor_KeyValidation(t *testing.T) {
	t.Run("Valid 32-byte raw key", func(t *testing.T) {
		rawKey := "12345678901234567890123456789012"
		enc, err := NewGCMEncryptor(rawKey)
		require.NoError(t, err)
		assert.NotNil(t, enc)
	})

	t.Run("Valid 64-char hex key", func(t *testing.T) {
		hexKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
		enc, err := NewGCMEncryptor(hexKey)
		require.NoError(t, err)
		assert.NotNil(t, enc)
	})

	t.Run("Invalid short key", func(t *testing.T) {
		shortKey := "short_key_123"
		enc, err := NewGCMEncryptor(shortKey)
		assert.Error(t, err)
		assert.Nil(t, enc)
		assert.ErrorIs(t, err, ErrInvalidKeySize)
	})

	t.Run("Invalid hex length", func(t *testing.T) {
		invalidHex := "0123456789abcdef"
		enc, err := NewGCMEncryptor(invalidHex)
		assert.Error(t, err)
		assert.Nil(t, enc)
	})
}

func TestGCMEncryptor_EncryptDecrypt_RoundTrip(t *testing.T) {
	key := "01234567890123456789012345678901"
	enc, err := NewGCMEncryptor(key)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "Simple JSON credentials",
			plaintext: `{"api_key":"sup_live_12345","merchant_code":"MC9988"}`,
		},
		{
			name:      "Empty payload",
			plaintext: "",
		},
		{
			name:      "Large complex payload with special characters",
			plaintext: `{"token":"pbg_token_!@#$%^&*()_+{}[]:;<>?,./~","client_id":"cli_12345","additional":{"url":"https://api.pagbank.com/v1"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := enc.Encrypt([]byte(tc.plaintext))
			require.NoError(t, err)
			assert.NotEmpty(t, ciphertext)

			// Decrypt and assert round-trip equality
			decrypted, err := enc.Decrypt(ciphertext)
			require.NoError(t, err)
			assert.Equal(t, tc.plaintext, string(decrypted))
		})
	}
}

func TestGCMEncryptor_NonceUniqueness(t *testing.T) {
	key := "01234567890123456789012345678901"
	enc, err := NewGCMEncryptor(key)
	require.NoError(t, err)

	plaintext := []byte("identical_secret_payload")

	// Encrypt identical text twice
	c1, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	c2, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	// Due to distinct 12-byte random nonces, ciphertexts must NEVER be equal
	assert.NotEqual(t, c1, c2, "Two encryptions of identical plaintext must yield distinct ciphertexts")

	// Both must decrypt back to original plaintext
	d1, err := enc.Decrypt(c1)
	require.NoError(t, err)
	assert.Equal(t, plaintext, d1)

	d2, err := enc.Decrypt(c2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, d2)
}

func TestGCMEncryptor_TamperResistance(t *testing.T) {
	key := "01234567890123456789012345678901"
	enc, err := NewGCMEncryptor(key)
	require.NoError(t, err)

	plaintext := []byte("critical_financial_secret")
	ciphertextBase64, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	raw, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	require.NoError(t, err)

	// Tamper with the last byte (part of authentication tag)
	raw[len(raw)-1] ^= 0xFF
	tamperedCiphertext := base64.StdEncoding.EncodeToString(raw)

	// Decrypt must fail due to GCM authentication tag mismatch
	_, err = enc.Decrypt(tamperedCiphertext)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDecryptionFailed)
}

func TestGCMEncryptor_InvalidCiphertextInputs(t *testing.T) {
	key := "01234567890123456789012345678901"
	enc, err := NewGCMEncryptor(key)
	require.NoError(t, err)

	t.Run("Invalid base64 string", func(t *testing.T) {
		_, err := enc.Decrypt("not_a_valid_base64!@#$")
		assert.Error(t, err)
	})

	t.Run("Ciphertext too short", func(t *testing.T) {
		shortRaw := make([]byte, 15) // less than 12 nonce + 16 tag = 28 bytes
		shortB64 := base64.StdEncoding.EncodeToString(shortRaw)
		_, err := enc.Decrypt(shortB64)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrCiphertextTooShort)
	})
}
