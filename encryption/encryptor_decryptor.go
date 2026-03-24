package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"

	"github.com/zendesk/go-generics/serialize"
)

const MinSaltLength = 16
const NonceLength = 12

var ErrSaltTooShort = errors.New("salt too short")
var ErrInvalidNonce = errors.New("invalid nonce length")

type EncryptorDecryptor[T any] struct {
	aesgcm cipher.AEAD
	nonce  []byte
}

func validateSalt(salt []byte) error {
	if len(salt) < MinSaltLength {
		return fmt.Errorf("%w: must be at least %d bytes, got %d", ErrSaltTooShort, MinSaltLength, len(salt))
	}
	return nil
}

// New creates a new EncryptorDecryptor instance with a random password and nonce.
func New[T any](salt []byte, iterations int) (*EncryptorDecryptor[T], error) {
	if err := validateSalt(salt); err != nil {
		return nil, err
	}

	password := make([]byte, 2048)
	if _, err := io.ReadFull(rand.Reader, password); err != nil {
		return nil, fmt.Errorf("error generating password: %w", err)
	}

	aesKey := pbkdf2.Key(password, salt, iterations, 32, sha256.New)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error creating new GCM: %w", err)
	}

	nonce := make([]byte, NonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}

	e := &EncryptorDecryptor[T]{
		aesgcm: aesgcm,
		nonce:  nonce,
	}

	return e, nil
}

// NewWithPasswordNonce creates a new EncryptorDecryptor instance with a specified password and nonce. Use this if
// you intend to persist whatever is encrypted.
func NewWithPasswordNonce[T any](password, nonce, salt []byte, iterations int) (*EncryptorDecryptor[T], error) {
	if len(nonce) != NonceLength {
		return nil, fmt.Errorf("%w: must be exactly %d bytes, got %d", ErrInvalidNonce, NonceLength, len(nonce))
	}
	if err := validateSalt(salt); err != nil {
		return nil, err
	}

	aesKey := pbkdf2.Key(password, salt, iterations, 32, sha256.New)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error creating new GCM: %w", err)
	}

	e := &EncryptorDecryptor[T]{
		aesgcm: aesgcm,
		nonce:  nonce,
	}

	return e, nil
}

func (e *EncryptorDecryptor[T]) Encrypt(value T) ([]byte, error) {
	w := wrapper[T]{T: value}

	bytes, err := serialize.NewSerializer[wrapper[T]]().FromDynamicType(w).ToBytes()
	if err != nil {
		return []byte{}, err
	}
	ciphertext := e.aesgcm.Seal(nil, e.nonce, bytes, nil)

	return ciphertext, nil
}

func (e *EncryptorDecryptor[T]) Decrypt(ciphertext []byte) (T, error) {
	var t T
	decryptedBytes, err := e.aesgcm.Open(nil, e.nonce, ciphertext, nil)
	if err != nil {
		return t, fmt.Errorf("error decrypting ciphertext: %w", err)
	}

	w := wrapper[T]{T: t}

	result, err := serialize.NewSerializer[wrapper[T]]().FromBytes(decryptedBytes).ToDynamicType(serialize.Reflect, w)

	// Convert to wrapper[T]
	converted, ok := result.(wrapper[T])
	if !ok {
		return t, fmt.Errorf("error converting result to type T")
	}

	return converted.T, err
}

type wrapper[T any] struct {
	T T
}
