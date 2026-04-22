//go:build test
// +build test

package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"reflect"
	"testing"

	"golang.org/x/crypto/pbkdf2"

	"github.com/zendesk/go-generics/serialize"
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

func TestNew_WithAllowLegacyParameters_StillRejectsInsaneParameters(t *testing.T) {
	// WithAllowLegacyParameters relaxes the floor — it does NOT disable the
	// basic sanity checks that protect against meaningless PBKDF2 inputs.
	// pbkdf2.Key does not error on non-positive iterations or empty salts;
	// it silently produces a trivially-derived key, so callers must be stopped
	// at construction time regardless of the opt-out.
	cases := []struct {
		name       string
		salt       []byte
		iterations int
		wantErr    error
	}{
		{"zero iterations", []byte("saltyvalu2"), 0, ErrIterationsTooLow},
		{"negative iterations", []byte("saltyvalu2"), -1, ErrIterationsTooLow},
		{"empty salt", []byte{}, 4096, ErrSaltTooShort},
		{"nil salt", nil, 4096, ErrSaltTooShort},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := New[string](tc.salt, tc.iterations, WithAllowLegacyParameters()); err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			} else if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got: %v", tc.wantErr, err)
			}

			if _, err := NewWithPassword[string]([]byte("password"), tc.salt, tc.iterations, WithAllowLegacyParameters()); err == nil {
				t.Fatalf("NewWithPassword: expected error for %s, got nil", tc.name)
			} else if !errors.Is(err, tc.wantErr) {
				t.Fatalf("NewWithPassword: expected %v, got: %v", tc.wantErr, err)
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

// encryptLegacyV0 faithfully reproduces the pre-#18 (2025-era) Encrypt wire format:
// a single fixed nonce is reused for every call, and Seal is called with nil as
// the dst so the output is just [ciphertext|tag] — no nonce prefix.
// Key derivation matches pre-#18 NewWithPasswordNonce: pbkdf2 with the same
// parameters that the current NewWithPassword uses.
func encryptLegacyV0[T any](t *testing.T, password, salt, nonce []byte, iterations int, value T) []byte {
	t.Helper()

	aesKey := pbkdf2.Key(password, salt, iterations, 32, sha256.New)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("cipher.NewGCM: %v", err)
	}

	// Exactly matches the pre-#18 Encrypt path in encryption/encryptor_decryptor.go:
	//   w := wrapper[T]{T: value}
	//   bytes, _ := serialize.NewSerializer[wrapper[T]]().FromDynamicType(w).ToBytes()
	//   ciphertext := e.aesgcm.Seal(nil, e.nonce, bytes, nil)
	w := wrapper[T]{T: value}
	plaintext, err := serialize.NewSerializer[wrapper[T]]().FromDynamicType(w).ToBytes()
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	return aesgcm.Seal(nil, nonce, plaintext, nil)
}

// TestLegacyV0_Decryptable_ByNewWithLegacyNonce is the core cross-version proof:
// ciphertext produced by the pre-#18 wire format (fixed nonce, no prefix) must
// roundtrip through a current EncryptorDecryptor when WithLegacyNonce is
// supplied, and the default (no-option) constructor must fail to decrypt it.
func TestLegacyV0_Decryptable_ByNewWithLegacyNonce(t *testing.T) {
	password := []byte("production-password")
	// Intentionally a 10-byte salt to mirror ssv2-api's CryptoSalt "73616C7479" —
	// the exact reason WithAllowLegacyParameters exists.
	salt := []byte("saltyvalu2")
	// Intentionally 4096 iterations, the legacy default that pre-dates MinIterations.
	iterations := 4096
	nonce := []byte("legacy-none1") // 12 bytes — matches NonceLength

	input := "secret written by old go-generics"
	legacyCiphertext := encryptLegacyV0[string](t, password, salt, nonce, iterations, input)

	// 1. A current-default encryptor must NOT be able to decrypt legacy bytes:
	//    it will slice off the first 12 bytes as a (wrong) nonce and fail auth.
	currentDefault, err := NewWithPassword[string](password, []byte("1234567890abcdef"), MinIterations)
	if err != nil {
		t.Fatalf("NewWithPassword (default) setup: %v", err)
	}
	if _, err := currentDefault.Decrypt(legacyCiphertext); err == nil {
		t.Fatal("current default decrypter unexpectedly decrypted legacy ciphertext — wire format detection is broken")
	}

	// 2. A current encryptor opted in to legacy parameters AND legacy nonce MUST
	//    be able to decrypt the legacy bytes.
	legacyReader, err := NewWithPassword[string](
		password, salt, iterations,
		WithAllowLegacyParameters(),
		WithLegacyNonce(nonce),
	)
	if err != nil {
		t.Fatalf("NewWithPassword with legacy opts: %v", err)
	}
	decrypted, err := legacyReader.Decrypt(legacyCiphertext)
	if err != nil {
		t.Fatalf("legacyReader.Decrypt: %v", err)
	}
	if decrypted != input {
		t.Errorf("Decrypt() = %q, want %q", decrypted, input)
	}
}

// TestLegacyV0_Encrypt_ReadableByLegacyAlgorithm confirms the legacy-nonce path
// produces the same wire format the old version did. We encrypt with the new
// code + WithLegacyNonce, then decrypt by hand using the documented pre-#18
// algorithm (Open with fixed nonce against [ciphertext|tag]). If this roundtrips
// we've proven Encrypt is byte-compatible with the old format.
func TestLegacyV0_Encrypt_ReadableByLegacyAlgorithm(t *testing.T) {
	password := []byte("production-password")
	salt := []byte("saltyvalu2")
	iterations := 4096
	nonce := []byte("legacy-none1")

	input := "written by new encryptor in legacy mode"

	ed, err := NewWithPassword[string](
		password, salt, iterations,
		WithAllowLegacyParameters(),
		WithLegacyNonce(nonce),
	)
	if err != nil {
		t.Fatalf("NewWithPassword with legacy opts: %v", err)
	}
	ciphertext, err := ed.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Hand-rolled pre-#18 Decrypt: no nonce prefix, fixed nonce, plain Open.
	aesKey := pbkdf2.Key(password, salt, iterations, 32, sha256.New)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("cipher.NewGCM: %v", err)
	}
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		t.Fatalf("legacy-style Open: %v", err)
	}

	w := wrapper[string]{}
	result, err := serialize.NewSerializer[wrapper[string]]().FromBytes(plaintext).ToDynamicType(serialize.Reflect, w)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	got, ok := result.(wrapper[string])
	if !ok {
		t.Fatalf("result is %T, not wrapper[string]", result)
	}
	if got.T != input {
		t.Errorf("roundtrip = %q, want %q", got.T, input)
	}
}

// TestLegacyV0_Roundtrip_WithinCurrent is the pragmatic "ssv2-api" check: the
// current encryptor, opted in to legacy mode, roundtrips its own output. This
// is what the service will actually do in production (write + read legacy rows).
func TestLegacyV0_Roundtrip_WithinCurrent(t *testing.T) {
	salt := []byte("saltyvalu2")
	iterations := 4096
	nonce := []byte("legacy-none1")

	ed, err := NewWithPassword[string](
		[]byte("production-password"), salt, iterations,
		WithAllowLegacyParameters(),
		WithLegacyNonce(nonce),
	)
	if err != nil {
		t.Fatalf("NewWithPassword: %v", err)
	}

	for _, input := range []string{"", "short", "a much longer secret value that is also encrypted"} {
		ct, err := ed.Encrypt(input)
		if err != nil {
			t.Fatalf("Encrypt(%q): %v", input, err)
		}
		// Fixed-nonce Seal does not prefix a nonce: no [nonce|...] framing.
		// We cannot test the exact length cheaply here (serializer adds overhead),
		// but we can assert we did NOT accidentally double up by checking that
		// our own Decrypt on the output succeeds.
		got, err := ed.Decrypt(ct)
		if err != nil {
			t.Fatalf("Decrypt(%q): %v", input, err)
		}
		if got != input {
			t.Errorf("roundtrip: got %q, want %q", got, input)
		}
	}
}

// TestLegacyNonce_RejectsWrongLength protects callers from silently producing
// garbage ciphertext when they pass a non-12-byte nonce.
func TestLegacyNonce_RejectsWrongLength(t *testing.T) {
	tests := []struct {
		name  string
		nonce []byte
	}{
		{"empty", []byte{}},
		{"too short", []byte("short")},
		{"too long", []byte("way too long to be a GCM nonce")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWithPassword[string](
				[]byte("password"), []byte("1234567890abcdef"), MinIterations,
				WithLegacyNonce(tt.nonce),
			)
			if err == nil {
				t.Fatal("expected ErrInvalidNonce, got nil")
			}
			if !errors.Is(err, ErrInvalidNonce) {
				t.Fatalf("expected ErrInvalidNonce, got: %v", err)
			}
		})
	}
}

// TestCurrent_vs_Legacy_WireFormats_Differ is a guard: if someone ever makes
// Encrypt emit the legacy format by default, this test will fail. It confirms
// that in the default (no WithLegacyNonce) path, ciphertext starts with the
// randomly generated nonce — i.e. the wire formats are genuinely distinct.
func TestCurrent_vs_Legacy_WireFormats_Differ(t *testing.T) {
	password := []byte("password")
	salt := []byte("1234567890abcdef")
	nonce := []byte("legacy-none1")

	current, err := NewWithPassword[string](password, salt, MinIterations)
	if err != nil {
		t.Fatalf("default: %v", err)
	}
	legacy, err := NewWithPassword[string](
		password, salt, MinIterations,
		WithLegacyNonce(nonce),
	)
	if err != nil {
		t.Fatalf("legacy: %v", err)
	}

	currentCT, err := current.Encrypt("hello")
	if err != nil {
		t.Fatalf("current.Encrypt: %v", err)
	}
	legacyCT, err := legacy.Encrypt("hello")
	if err != nil {
		t.Fatalf("legacy.Encrypt: %v", err)
	}

	// Current must be exactly NonceLength bytes longer than legacy for the same
	// plaintext, because it prepends a fresh nonce and uses the same AEAD tag.
	if len(currentCT)-len(legacyCT) != NonceLength {
		t.Errorf("expected current ciphertext to be %d bytes longer than legacy, got current=%d legacy=%d",
			NonceLength, len(currentCT), len(legacyCT))
	}

	// Current must NOT decrypt legacy bytes, and vice-versa.
	if _, err := current.Decrypt(legacyCT); err == nil {
		t.Error("current.Decrypt unexpectedly accepted legacy ciphertext")
	}
	if _, err := legacy.Decrypt(currentCT); err == nil {
		t.Error("legacy.Decrypt unexpectedly accepted current ciphertext")
	}

	// Legacy ciphertexts for identical plaintext must be byte-identical (fixed
	// nonce reuse) — this is the documented, insecure property of legacy mode,
	// and proves the nonce is actually fixed.
	legacyCT2, err := legacy.Encrypt("hello")
	if err != nil {
		t.Fatalf("legacy.Encrypt (2): %v", err)
	}
	if !bytes.Equal(legacyCT, legacyCT2) {
		t.Error("legacy encryptor produced different ciphertexts for the same plaintext — fixed nonce is not actually fixed")
	}
}
