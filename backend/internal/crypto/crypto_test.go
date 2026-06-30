package crypto

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/cipherkeep/backend/internal/domain"
)

func mustKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i)
	}
	return key
}

func TestSealOpenRoundTrip(t *testing.T) {
	key := mustKey(t)
	cases := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte("")},
		{"short", []byte("hello")},
		{"secret", []byte("super-secret-database-url://user:pass@host/db")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ct, nonce, err := Seal(key, tc.plaintext)
			if err != nil {
				t.Fatalf("Seal: %v", err)
			}
			if len(nonce) != NonceSize {
				t.Fatalf("nonce size: got %d want %d", len(nonce), NonceSize)
			}
			pt, err := Open(key, ct, nonce)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			if !bytes.Equal(pt, tc.plaintext) {
				t.Fatalf("round trip mismatch: got %q want %q", pt, tc.plaintext)
			}
		})
	}
}

func TestOpenTamperDetection(t *testing.T) {
	key := mustKey(t)
	ct, nonce, err := Seal(key, []byte("tamper-me"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	// Flip a bit in the ciphertext.
	tampered := make([]byte, len(ct))
	copy(tampered, ct)
	tampered[0] ^= 0xFF
	if _, err := Open(key, tampered, nonce); !errors.Is(err, ErrDecrypt) {
		t.Fatalf("expected ErrDecrypt on tampered ciphertext, got %v", err)
	}
}

func TestOpenWrongKey(t *testing.T) {
	key := mustKey(t)
	ct, nonce, err := Seal(key, []byte("data"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	wrong := mustKey(t)
	wrong[0] ^= 0x01
	if _, err := Open(wrong, ct, nonce); !errors.Is(err, ErrDecrypt) {
		t.Fatalf("expected ErrDecrypt with wrong key, got %v", err)
	}
}

func TestSealUniqueNonces(t *testing.T) {
	key := mustKey(t)
	_, n1, _ := Seal(key, []byte("x"))
	_, n2, _ := Seal(key, []byte("x"))
	if bytes.Equal(n1, n2) {
		t.Fatal("nonces must be unique per encryption")
	}
}

func TestArgon2PasswordHashVerify(t *testing.T) {
	p := DefaultArgon2Params()
	hash, err := HashPassword("correct horse battery staple", p)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	ok, err := VerifyPassword("correct horse battery staple", hash)
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}
	bad, err := VerifyPassword("wrong password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword(bad): %v", err)
	}
	if bad {
		t.Fatal("expected wrong password to fail verification")
	}
}

// fakeEncKeyRepo is an in-memory EncryptionKeyRepository for envelope tests.
type fakeEncKeyRepo struct {
	row *domain.EncryptionKey
}

func (f *fakeEncKeyRepo) GetActive(_ context.Context, _ domain.Querier) (*domain.EncryptionKey, error) {
	if f.row == nil {
		return nil, domain.ErrNotFound
	}
	return f.row, nil
}

func (f *fakeEncKeyRepo) Create(_ context.Context, _ domain.Querier, k *domain.EncryptionKey) error {
	cp := *k
	f.row = &cp
	return nil
}

func TestEnvelopeBootstrapAndReload(t *testing.T) {
	ctx := context.Background()
	repo := &fakeEncKeyRepo{}
	const master = "master-password-123"

	// First run: bootstrap.
	svc, err := LoadEnvelopeService(ctx, nil, repo, master)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if repo.row == nil {
		t.Fatal("bootstrap must persist an encryption key row")
	}

	ct, nonce, err := svc.Encrypt([]byte("DATABASE_URL=postgres://x"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Second run: reload + unwrap with correct master password.
	svc2, err := LoadEnvelopeService(ctx, nil, repo, master)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	pt, err := svc2.Decrypt(ct, nonce)
	if err != nil {
		t.Fatalf("Decrypt after reload: %v", err)
	}
	if string(pt) != "DATABASE_URL=postgres://x" {
		t.Fatalf("reloaded DEK decrypts wrong value: %q", pt)
	}
}

func TestEnvelopeWrongMasterPassword(t *testing.T) {
	ctx := context.Background()
	repo := &fakeEncKeyRepo{}
	if _, err := LoadEnvelopeService(ctx, nil, repo, "right-password"); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	_, err := LoadEnvelopeService(ctx, nil, repo, "wrong-password")
	if !errors.Is(err, ErrWrongMasterPassword) {
		t.Fatalf("expected ErrWrongMasterPassword, got %v", err)
	}
}
