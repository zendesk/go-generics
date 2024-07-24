package encryption

import (
	"reflect"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
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
			ed, err := NewWithIterations(tt.iterations)
			if err != nil {
				t.Fatalf("Error during New: %v", err)
			}

			ciphertext, err := ed.Encrypt(tt.input)
			if (err != nil) != tt.shouldErrEncrypt {
				t.Fatalf("Encrypt() error = %v, wantErr %v", err, tt.shouldErrEncrypt)
			}

			plaintext, err := ed.Decrypt(ciphertext)
			if (err != nil) != tt.shouldErrDecrypt {
				t.Fatalf("Decrypt() error = %v, wantErr %v", err, tt.shouldErrDecrypt)
			}

			if !tt.shouldErrDecrypt && !reflect.DeepEqual(plaintext, tt.input) {
				t.Errorf("Decrypt() = %v, want %v", plaintext, tt.input)
			}
		})
	}
}
