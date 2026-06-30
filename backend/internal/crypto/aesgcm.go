package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

// NonceSize is the GCM nonce size in bytes.
const NonceSize = 12

// KeySize is the AES-256 key size in bytes.
const KeySize = 32

// ErrDecrypt indicates GCM authentication/decryption failure (tampering or wrong key).
var ErrDecrypt = errors.New("decryption failed: invalid key or corrupted data")

// Seal encrypts plaintext with AES-256-GCM using a fresh random nonce.
// It returns the ciphertext (with appended auth tag) and the nonce.
func Seal(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Open decrypts ciphertext with AES-256-GCM. A failure (auth tag mismatch or
// wrong key) returns ErrDecrypt and never partial data.
func Open(key, ciphertext, nonce []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce size: got %d want %d", len(nonce), NonceSize)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecrypt
	}
	return plaintext, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: got %d want %d", len(key), KeySize)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return gcm, nil
}

// Zero best-effort wipes a sensitive byte slice.
func Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
