package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params configures the Argon2id KDF.
type Argon2Params struct {
	Time    uint32 // number of iterations
	Memory  uint32 // memory in KiB
	Threads uint8  // parallelism
	KeyLen  uint32 // derived key length in bytes
}

// DefaultArgon2Params returns parameters per security.md (time=2, memory=64 MiB,
// threads=number of CPUs, key length 32 bytes).
func DefaultArgon2Params() Argon2Params {
	threads := runtime.NumCPU()
	if threads < 1 {
		threads = 1
	}
	if threads > 255 {
		threads = 255
	}
	return Argon2Params{
		Time:    2,
		Memory:  64 * 1024,
		Threads: uint8(threads),
		KeyLen:  32,
	}
}

// DeriveKey derives a key from a password and salt using Argon2id. Used to
// derive the KEK from the master password.
func DeriveKey(password string, salt []byte, p Argon2Params) []byte {
	return argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
}

// HashPassword hashes a user password with Argon2id and returns a PHC-style
// encoded string that embeds the parameters and salt.
func HashPassword(password string, p Argon2Params) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.Memory, p.Time, p.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
	return encoded, nil
}

// VerifyPassword reports whether the password matches the encoded Argon2id hash.
func VerifyPassword(password, encoded string) (bool, error) {
	p, salt, hash, err := decodeHash(encoded)
	if err != nil {
		return false, err
	}
	computed := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	if subtle.ConstantTimeCompare(hash, computed) == 1 {
		return true, nil
	}
	return false, nil
}

// decodeHash parses a PHC-style Argon2id encoded hash.
func decodeHash(encoded string) (Argon2Params, []byte, []byte, error) {
	var p Argon2Params
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return p, nil, nil, errors.New("invalid argon2 hash format")
	}
	if parts[1] != "argon2id" {
		return p, nil, nil, errors.New("unsupported argon2 variant")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return p, nil, nil, fmt.Errorf("parse version: %w", err)
	}
	if version != argon2.Version {
		return p, nil, nil, errors.New("incompatible argon2 version")
	}

	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Threads); err != nil {
		return p, nil, nil, fmt.Errorf("parse params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return p, nil, nil, fmt.Errorf("decode salt: %w", err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return p, nil, nil, fmt.Errorf("decode hash: %w", err)
	}
	p.KeyLen = uint32(len(hash))
	return p, salt, hash, nil
}
