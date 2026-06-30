package crypto

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/cipherkeep/backend/internal/domain"
)

// ErrWrongMasterPassword indicates the DEK could not be unwrapped, meaning the
// supplied master password is incorrect. The process must fail fast.
var ErrWrongMasterPassword = errors.New("wrong master password: cannot unwrap DEK")

// EnvelopeService performs envelope encryption: it holds the unwrapped DEK in
// memory and seals/opens secret values with it.
type EnvelopeService struct {
	dek []byte
}

// Encryptor is the subset of EnvelopeService used by services to protect secrets.
type Encryptor interface {
	Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error)
	Decrypt(ciphertext, nonce []byte) ([]byte, error)
}

// LoadEnvelopeService initializes the envelope service. On first run (no active
// encryption_keys row) it bootstraps a salt + DEK, wraps the DEK with the KEK
// derived from masterPassword, and persists it. On subsequent runs it loads the
// row and unwraps the DEK, failing fast if the master password is wrong.
func LoadEnvelopeService(
	ctx context.Context,
	q domain.Querier,
	repo domain.EncryptionKeyRepository,
	masterPassword string,
) (*EnvelopeService, error) {
	params := DefaultArgon2Params()

	existing, err := repo.GetActive(ctx, q)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("load encryption key: %w", err)
	}

	if existing == nil || errors.Is(err, domain.ErrNotFound) {
		return bootstrap(ctx, q, repo, masterPassword, params)
	}

	// Subsequent run: derive KEK, unwrap DEK.
	kek := DeriveKey(masterPassword, existing.KDFSalt, params)
	defer Zero(kek)

	dek, openErr := Open(kek, existing.WrappedDEK, existing.Nonce)
	if openErr != nil {
		return nil, ErrWrongMasterPassword
	}
	if len(dek) != KeySize {
		return nil, fmt.Errorf("unwrapped DEK has invalid size: %d", len(dek))
	}
	return &EnvelopeService{dek: dek}, nil
}

// bootstrap generates and persists a new wrapped DEK.
func bootstrap(
	ctx context.Context,
	q domain.Querier,
	repo domain.EncryptionKeyRepository,
	masterPassword string,
	params Argon2Params,
) (*EnvelopeService, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate kdf salt: %w", err)
	}

	kek := DeriveKey(masterPassword, salt, params)
	defer Zero(kek)

	dek := make([]byte, KeySize)
	if _, err := rand.Read(dek); err != nil {
		return nil, fmt.Errorf("generate dek: %w", err)
	}

	wrapped, nonce, err := Seal(kek, dek)
	if err != nil {
		return nil, fmt.Errorf("wrap dek: %w", err)
	}

	key := &domain.EncryptionKey{
		WrappedDEK: wrapped,
		Nonce:      nonce,
		KDFSalt:    salt,
		IsActive:   true,
	}
	if err := repo.Create(ctx, q, key); err != nil {
		return nil, fmt.Errorf("persist encryption key: %w", err)
	}

	return &EnvelopeService{dek: dek}, nil
}

// Encrypt seals plaintext with the DEK.
func (s *EnvelopeService) Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
	return Seal(s.dek, plaintext)
}

// Decrypt opens ciphertext with the DEK.
func (s *EnvelopeService) Decrypt(ciphertext, nonce []byte) ([]byte, error) {
	return Open(s.dek, ciphertext, nonce)
}
