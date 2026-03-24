//go:build test
// +build test

package encryption

import (
	"errors"
	"reflect"
	"testing"
)

func TestEncryptDecrypt_Bytes(t *testing.T) {
	tests := []struct {
		name             string
		iterations       int
		input            []byte
		shouldErrEncrypt bool
		shouldErrDecrypt bool
	}{
		{
			name:             "normal use case",
			iterations:       MinIterations,
			input:            []byte("test input"),
			shouldErrEncrypt: false,
			shouldErrDecrypt: false,
		},
		// additional test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed, err := New[[]byte]([]byte("1234567890abcdef"), tt.iterations)
			if err != nil {
				t.Fatalf("Error during New: %v", err)
			}

			ciphertext, err := ed.Encrypt(tt.input)
			if (err != nil) != tt.shouldErrEncrypt {
				t.Fatalf("Encrypt() error = %v, wantErr %v", err, tt.shouldErrEncrypt)
			}

			decrypted, err := ed.Decrypt(ciphertext)
			if (err != nil) != tt.shouldErrDecrypt {
				t.Fatalf("Decrypt() error = %v, wantErr %v", err, tt.shouldErrDecrypt)
			}

			if !tt.shouldErrDecrypt && !reflect.DeepEqual(decrypted, tt.input) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.input)
			}
		})
	}
}

func TestEncryptDecrypt_String(t *testing.T) {
	tests := []struct {
		name             string
		iterations       int
		input            string
		shouldErrEncrypt bool
		shouldErrDecrypt bool
	}{
		{
			name:             "normal use case",
			iterations:       MinIterations,
			input:            "Test input",
			shouldErrEncrypt: false,
			shouldErrDecrypt: false,
		},
		// additional test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed, err := New[string]([]byte("1234567890abcdef"), tt.iterations)
			if err != nil {
				t.Fatalf("Error during New: %v", err)
			}

			ciphertext, err := ed.Encrypt(tt.input)
			if (err != nil) != tt.shouldErrEncrypt {
				t.Fatalf("Encrypt() error = %v, wantErr %v", err, tt.shouldErrEncrypt)
			}

			decrypted, err := ed.Decrypt(ciphertext)
			if (err != nil) != tt.shouldErrDecrypt {
				t.Fatalf("Decrypt() error = %v, wantErr %v", err, tt.shouldErrDecrypt)
			}

			if !tt.shouldErrDecrypt && !reflect.DeepEqual(decrypted, tt.input) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.input)
			}
		})
	}
}

func TestEncryptDecrypt_Struct(t *testing.T) {
	type Foo struct {
		Value string
		Item  []byte
		Thing int
	}
	tests := []struct {
		name             string
		iterations       int
		input            Foo
		shouldErrEncrypt bool
		shouldErrDecrypt bool
	}{
		{
			name:       "normal use case",
			iterations: MinIterations,
			input: Foo{
				Value: "dflkasjflkajsdf",
				Item:  nil,
				Thing: 1231414,
			},
			shouldErrEncrypt: false,
			shouldErrDecrypt: false,
		},
		// additional test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed, err := New[Foo]([]byte("1234567890abcdef"), tt.iterations)
			if err != nil {
				t.Fatalf("Error during New: %v", err)
			}

			ciphertext, err := ed.Encrypt(tt.input)
			if (err != nil) != tt.shouldErrEncrypt {
				t.Fatalf("Encrypt() error = %v, wantErr %v", err, tt.shouldErrEncrypt)
			}

			decrypted, err := ed.Decrypt(ciphertext)
			if (err != nil) != tt.shouldErrDecrypt {
				t.Fatalf("Decrypt() error = %v, wantErr %v", err, tt.shouldErrDecrypt)
			}

			if !tt.shouldErrDecrypt && !reflect.DeepEqual(decrypted, tt.input) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.input)
			}
		})
	}
}

func TestEncryptDecrypt_StructSlice(t *testing.T) {
	type Foo struct {
		Value string
		Item  []byte
		Thing int
	}
	tests := []struct {
		name             string
		iterations       int
		input            []*Foo
		shouldErrEncrypt bool
		shouldErrDecrypt bool
	}{
		{
			name:       "normal use case",
			iterations: MinIterations,
			input: []*Foo{
				{
					Value: "item 1",
					Item:  []byte{1, 33, 44},
					Thing: 111,
				},
				{

					Value: "item 2",
					Item:  nil,
					Thing: 0,
				},
			},
			shouldErrEncrypt: false,
			shouldErrDecrypt: false,
		},
		// additional test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed, err := New[[]*Foo]([]byte("1234567890abcdef"), tt.iterations)
			if err != nil {
				t.Fatalf("Error during New: %v", err)
			}

			ciphertext, err := ed.Encrypt(tt.input)
			if (err != nil) != tt.shouldErrEncrypt {
				t.Fatalf("Encrypt() error = %v, wantErr %v", err, tt.shouldErrEncrypt)
			}

			decrypted, err := ed.Decrypt(ciphertext)
			if (err != nil) != tt.shouldErrDecrypt {
				t.Fatalf("Decrypt() error = %v, wantErr %v", err, tt.shouldErrDecrypt)
			}

			if !tt.shouldErrDecrypt && !reflect.DeepEqual(decrypted, tt.input) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.input)
			}
		})
	}
}

func TestEncryptDecrypt_Struct_WithPassword(t *testing.T) {
	type Foo struct {
		Value string
		Item  []byte
		Thing int
	}
	tests := []struct {
		name             string
		iterations       int
		input            Foo
		shouldErrEncrypt bool
		shouldErrDecrypt bool
	}{
		{
			name:       "normal use case",
			iterations: MinIterations,
			input: Foo{
				Value: "dflkasjflkajsdf",
				Item:  nil,
				Thing: 1231414,
			},
			shouldErrEncrypt: false,
			shouldErrDecrypt: false,
		},
		// additional test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed, err := NewWithPasswordNonce[Foo]([]byte("password"), []byte("my-nonce-123"), []byte("1234567890abcdef"), tt.iterations)
			if err != nil {
				t.Fatalf("Error during New: %v", err)
			}

			ciphertext, err := ed.Encrypt(tt.input)
			if (err != nil) != tt.shouldErrEncrypt {
				t.Fatalf("Encrypt() error = %v, wantErr %v", err, tt.shouldErrEncrypt)
			}

			decrypted, err := ed.Decrypt(ciphertext)
			if (err != nil) != tt.shouldErrDecrypt {
				t.Fatalf("Decrypt() error = %v, wantErr %v", err, tt.shouldErrDecrypt)
			}

			if !tt.shouldErrDecrypt && !reflect.DeepEqual(decrypted, tt.input) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.input)
			}
		})
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

func TestNewWithPasswordNonce_RejectsInvalidNonce(t *testing.T) {
	tests := []struct {
		name  string
		nonce []byte
	}{
		{"nil nonce", nil},
		{"empty nonce", []byte{}},
		{"too short", make([]byte, NonceLength-1)},
		{"too long", make([]byte, NonceLength+1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWithPasswordNonce[string]([]byte("password"), tt.nonce, []byte("1234567890abcdef"), MinIterations)
			if err == nil {
				t.Fatal("expected error for invalid nonce, got nil")
			}
			if !errors.Is(err, ErrInvalidNonce) {
				t.Fatalf("expected ErrInvalidNonce, got: %v", err)
			}
		})
	}
}

func TestNewWithPasswordNonce_RejectsShortSalt(t *testing.T) {
	_, err := NewWithPasswordNonce[string]([]byte("password"), []byte("my-nonce-123"), []byte("short"), MinIterations)
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

func TestNewWithPasswordNonce_RejectsLowIterations(t *testing.T) {
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
			_, err := NewWithPasswordNonce[string]([]byte("password"), []byte("my-nonce-123"), []byte("1234567890abcdef"), tt.iterations)
			if err == nil {
				t.Fatal("expected error for low iterations, got nil")
			}
			if !errors.Is(err, ErrIterationsTooLow) {
				t.Fatalf("expected ErrIterationsTooLow, got: %v", err)
			}
		})
	}
}
