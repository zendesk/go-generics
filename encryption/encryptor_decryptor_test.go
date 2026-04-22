//go:build test
// +build test

package encryption

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestEncryptDecrypt_Bytes(t *testing.T) {
	ed, err := New[[]byte]([]byte("1234567890abcdef"), MinIterations)
	if err != nil {
		t.Fatalf("Error during New: %v", err)
	}

	input := []byte("test input")
	ciphertext, err := ed.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	decrypted, err := ed.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if !reflect.DeepEqual(decrypted, input) {
		t.Errorf("Decrypt() = %v, want %v", decrypted, input)
	}
}

func TestEncryptDecrypt_String(t *testing.T) {
	ed, err := New[string]([]byte("1234567890abcdef"), MinIterations)
	if err != nil {
		t.Fatalf("Error during New: %v", err)
	}

	input := "Test input"
	ciphertext, err := ed.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	decrypted, err := ed.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if decrypted != input {
		t.Errorf("Decrypt() = %v, want %v", decrypted, input)
	}
}

func TestEncryptDecrypt_Struct(t *testing.T) {
	type Foo struct {
		Value string
		Item  []byte
		Thing int
	}

	ed, err := New[Foo]([]byte("1234567890abcdef"), MinIterations)
	if err != nil {
		t.Fatalf("Error during New: %v", err)
	}

	input := Foo{Value: "dflkasjflkajsdf", Item: nil, Thing: 1231414}
	ciphertext, err := ed.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	decrypted, err := ed.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if !reflect.DeepEqual(decrypted, input) {
		t.Errorf("Decrypt() = %v, want %v", decrypted, input)
	}
}

func TestEncryptDecrypt_StructSlice(t *testing.T) {
	type Foo struct {
		Value string
		Item  []byte
		Thing int
	}

	ed, err := New[[]*Foo]([]byte("1234567890abcdef"), MinIterations)
	if err != nil {
		t.Fatalf("Error during New: %v", err)
	}

	input := []*Foo{
		{Value: "item 1", Item: []byte{1, 33, 44}, Thing: 111},
		{Value: "item 2", Item: nil, Thing: 0},
	}
	ciphertext, err := ed.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	decrypted, err := ed.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if !reflect.DeepEqual(decrypted, input) {
		t.Errorf("Decrypt() = %v, want %v", decrypted, input)
	}
}

func TestEncryptDecrypt_WithPassword(t *testing.T) {
	type Foo struct {
		Value string
		Item  []byte
		Thing int
	}

	ed, err := NewWithPassword[Foo]([]byte("password"), []byte("1234567890abcdef"), MinIterations)
	if err != nil {
		t.Fatalf("Error during NewWithPassword: %v", err)
	}

	input := Foo{Value: "dflkasjflkajsdf", Item: nil, Thing: 1231414}
	ciphertext, err := ed.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	decrypted, err := ed.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if !reflect.DeepEqual(decrypted, input) {
		t.Errorf("Decrypt() = %v, want %v", decrypted, input)
	}
}

func TestEncryptDecrypt_WithPasswordNonce_Compat(t *testing.T) {
	// Verify the deprecated NewWithPasswordNonce shim still works
	ed, err := NewWithPasswordNonce[string]([]byte("password"), []byte("ignored-nonce"), []byte("1234567890abcdef"), MinIterations)
	if err != nil {
		t.Fatalf("Error during NewWithPasswordNonce: %v", err)
	}

	input := "compat test"
	ciphertext, err := ed.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	decrypted, err := ed.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if decrypted != input {
		t.Errorf("Decrypt() = %v, want %v", decrypted, input)
	}
}

func TestEncrypt_ProducesUniqueNonces(t *testing.T) {
	ed, err := New[string]([]byte("1234567890abcdef"), MinIterations)
	if err != nil {
		t.Fatalf("Error during New: %v", err)
	}

	ct1, err := ed.Encrypt("same value")
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}
	ct2, err := ed.Encrypt("same value")
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Ciphertexts must differ (different nonces) even for identical plaintext
	if bytes.Equal(ct1, ct2) {
		t.Error("two encryptions of the same value produced identical ciphertext — nonce reuse detected")
	}

	// The prepended nonces (first 12 bytes) must differ
	nonce1 := ct1[:NonceLength]
	nonce2 := ct2[:NonceLength]
	if bytes.Equal(nonce1, nonce2) {
		t.Error("two encryptions produced the same nonce")
	}

	// Both must still decrypt correctly
	dec1, err := ed.Decrypt(ct1)
	if err != nil {
		t.Fatalf("Decrypt(ct1) error: %v", err)
	}
	dec2, err := ed.Decrypt(ct2)
	if err != nil {
		t.Fatalf("Decrypt(ct2) error: %v", err)
	}
	if dec1 != "same value" || dec2 != "same value" {
		t.Errorf("decrypted values don't match: got %q and %q", dec1, dec2)
	}
}

func TestDecrypt_RejectsTruncatedCiphertext(t *testing.T) {
	ed, err := New[string]([]byte("1234567890abcdef"), MinIterations)
	if err != nil {
		t.Fatalf("Error during New: %v", err)
	}

	_, err = ed.Decrypt([]byte("short"))
	if err == nil {
		t.Fatal("expected error for truncated ciphertext, got nil")
	}
	if !errors.Is(err, ErrCiphertextTooShort) {
		t.Fatalf("expected ErrCiphertextTooShort, got: %v", err)
	}
}

func TestNew_RejectsShortSalt(t *testing.T) {
	_, err := New[string]([]byte("short"), MinIterations)
	if err == nil {
		t.Fatal("expected error for short salt, got nil")
	}
	if !errors.Is(err, ErrSaltTooShort) {
		t.Fatalf("expected ErrSaltTooShort, got: %v", err)
	}
}

func TestNewWithPassword_RejectsShortSalt(t *testing.T) {
	_, err := NewWithPassword[string]([]byte("password"), []byte("short"), MinIterations)
	if err == nil {
		t.Fatal("expected error for short salt, got nil")
	}
	if !errors.Is(err, ErrSaltTooShort) {
		t.Fatalf("expected ErrSaltTooShort, got: %v", err)
	}
}

func TestNew_RejectsLowIterations(t *testing.T) {
	tests := []struct {
		name       string
		iterations int
	}{
		{"zero", 0},
		{"negative", -1},
		{"just below minimum", MinIterations - 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New[string]([]byte("1234567890abcdef"), tt.iterations)
			if err == nil {
				t.Fatal("expected error for low iterations, got nil")
			}
			if !errors.Is(err, ErrIterationsTooLow) {
				t.Fatalf("expected ErrIterationsTooLow, got: %v", err)
			}
		})
	}
}

func TestNewWithPassword_RejectsLowIterations(t *testing.T) {
	tests := []struct {
		name       string
		iterations int
	}{
		{"zero", 0},
		{"negative", -1},
		{"just below minimum", MinIterations - 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWithPassword[string]([]byte("password"), []byte("1234567890abcdef"), tt.iterations)
			if err == nil {
				t.Fatal("expected error for low iterations, got nil")
			}
			if !errors.Is(err, ErrIterationsTooLow) {
				t.Fatalf("expected ErrIterationsTooLow, got: %v", err)
			}
		})
	}
}

func TestNew_WithAllowLegacyParameters_BypassesValidation(t *testing.T) {
	// Short salt (10 bytes) and iterations below MinIterations must be accepted
	// when the opt-out option is provided, to support legacy persisted data.
	shortSalt := []byte("saltyvalu2") // 10 bytes, below MinSaltLength=16
	lowIterations := 4096             // below MinIterations=100_000

	ed, err := New[string](shortSalt, lowIterations, WithAllowLegacyParameters())
	if err != nil {
		t.Fatalf("New() with WithAllowLegacyParameters returned error: %v", err)
	}

	input := "legacy roundtrip"
	ciphertext, err := ed.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}
	decrypted, err := ed.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}
	if decrypted != input {
		t.Errorf("Decrypt() = %q, want %q", decrypted, input)
	}
}

func TestNewWithPassword_WithAllowLegacyParameters_BypassesValidation(t *testing.T) {
	shortSalt := []byte("saltyvalu2")
	lowIterations := 4096

	ed, err := NewWithPassword[string]([]byte("password"), shortSalt, lowIterations, WithAllowLegacyParameters())
	if err != nil {
		t.Fatalf("NewWithPassword() with WithAllowLegacyParameters returned error: %v", err)
	}

	input := "legacy roundtrip"
	ciphertext, err := ed.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}
	decrypted, err := ed.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}
	if decrypted != input {
		t.Errorf("Decrypt() = %q, want %q", decrypted, input)
	}
}

func TestNewWithPasswordNonce_WithAllowLegacyParameters_BypassesValidation(t *testing.T) {
	// The deprecated shim must also forward the option.
	shortSalt := []byte("saltyvalu2")
	lowIterations := 4096

	ed, err := NewWithPasswordNonce[string]([]byte("password"), []byte("ignored-nonce"), shortSalt, lowIterations, WithAllowLegacyParameters())
	if err != nil {
		t.Fatalf("NewWithPasswordNonce() with WithAllowLegacyParameters returned error: %v", err)
	}

	input := "legacy roundtrip"
	ciphertext, err := ed.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}
	decrypted, err := ed.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}
	if decrypted != input {
		t.Errorf("Decrypt() = %q, want %q", decrypted, input)
	}
}
