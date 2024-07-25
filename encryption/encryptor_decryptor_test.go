package encryption

import (
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
			iterations:       4096,
			input:            []byte("test input"),
			shouldErrEncrypt: false,
			shouldErrDecrypt: false,
		},
		// additional test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed, err := New[[]byte]([]byte("salt"), tt.iterations)
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
			iterations:       4096,
			input:            "Test input",
			shouldErrEncrypt: false,
			shouldErrDecrypt: false,
		},
		// additional test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed, err := New[string]([]byte("salt"), tt.iterations)
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
			iterations: 4096,
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
			ed, err := New[Foo]([]byte("salt"), tt.iterations)
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
			iterations: 4096,
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
			ed, err := New[[]*Foo]([]byte("salt"), tt.iterations)
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
