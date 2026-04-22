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
const MinIterations = 100_000

var ErrSaltTooShort = errors.New("salt too short")

// ErrInvalidNonce is returned when a caller supplies a fixed legacy nonce via
// WithLegacyNonce whose length is not NonceLength. The default (no legacy nonce)
// path generates nonces internally and never returns this error.
var ErrInvalidNonce = errors.New("invalid nonce length")
var ErrIterationsTooLow = errors.New("iterations too low")
var ErrCiphertextTooShort = errors.New("ciphertext too short")

type EncryptorDecryptor[T any] struct {
	aesgcm cipher.AEAD
	// legacyNonce, when non-nil, selects the pre-#18 wire format: a fixed
	// caller-supplied nonce is reused for every Encrypt/Decrypt call and the
	// ciphertext is NOT self-describing (no nonce prefix). See WithLegacyNonce.
	legacyNonce []byte
}

func validateSalt(salt []byte) error {
	if len(salt) < MinSaltLength {
		return fmt.Errorf("%w: must be at least %d bytes, got %d", ErrSaltTooShort, MinSaltLength, len(salt))
	}
	return nil
}

func validateIterations(iterations int) error {
	if iterations < MinIterations {
		return fmt.Errorf("%w: must be at least %d, got %d", ErrIterationsTooLow, MinIterations, iterations)
	}
	return nil
}

// validateParameterSanity enforces the absolute minimum constraints that must
// hold regardless of WithAllowLegacyParameters. pbkdf2.Key does not validate
// its inputs and produces meaningless / trivially-derived keys when iterations
// is non-positive or salt is empty, so these checks are *not* opt-out-able.
func validateParameterSanity(salt []byte, iterations int) error {
	if iterations <= 0 {
		return fmt.Errorf("%w: must be greater than 0, got %d", ErrIterationsTooLow, iterations)
	}
	if len(salt) == 0 {
		return fmt.Errorf("%w: must not be empty", ErrSaltTooShort)
	}
	return nil
}

// New creates a new EncryptorDecryptor instance with a random password.
// A fresh random nonce is generated for each Encrypt call and prepended to the ciphertext.
//
// Pass WithAllowLegacyParameters to bypass salt-length and iteration-count
// validation; see that option's documentation for when it is appropriate.
// Pass WithLegacyNonce only to read ciphertext written by pre-#18 go-generics;
// see that option's documentation for the security implications.
func New[T any](salt []byte, iterations int, opts ...Option) (*EncryptorDecryptor[T], error) {
	cfg := resolveOptions(opts...)
	if err := validateParameterSanity(salt, iterations); err != nil {
		return nil, err
	}
	if !cfg.allowLegacyParameters {
		if err := validateIterations(iterations); err != nil {
			return nil, err
		}
		if err := validateSalt(salt); err != nil {
			return nil, err
		}
	}
	if cfg.legacyNonce != nil && len(cfg.legacyNonce) != NonceLength {
		return nil, fmt.Errorf("%w: must be exactly %d bytes, got %d", ErrInvalidNonce, NonceLength, len(cfg.legacyNonce))
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

	return &EncryptorDecryptor[T]{aesgcm: aesgcm, legacyNonce: cfg.legacyNonce}, nil
}

// NewWithPassword creates a new EncryptorDecryptor instance with a specified password. Use this if
// you intend to persist whatever is encrypted. A fresh random nonce is generated for each Encrypt
// call and prepended to the ciphertext.
//
// Pass WithAllowLegacyParameters to bypass salt-length and iteration-count
// validation; see that option's documentation for when it is appropriate.
// Pass WithLegacyNonce only to read ciphertext written by pre-#18 go-generics;
// see that option's documentation for the security implications.
func NewWithPassword[T any](password, salt []byte, iterations int, opts ...Option) (*EncryptorDecryptor[T], error) {
	cfg := resolveOptions(opts...)
	if err := validateParameterSanity(salt, iterations); err != nil {
		return nil, err
	}
	if !cfg.allowLegacyParameters {
		if err := validateIterations(iterations); err != nil {
			return nil, err
		}
		if err := validateSalt(salt); err != nil {
			return nil, err
		}
	}
	if cfg.legacyNonce != nil && len(cfg.legacyNonce) != NonceLength {
		return nil, fmt.Errorf("%w: must be exactly %d bytes, got %d", ErrInvalidNonce, NonceLength, len(cfg.legacyNonce))
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

	return &EncryptorDecryptor[T]{aesgcm: aesgcm, legacyNonce: cfg.legacyNonce}, nil
}

// Deprecated: NewWithPasswordNonce is replaced by NewWithPassword. Nonces are now generated
// per-Encrypt call and prepended to the ciphertext. The nonce parameter is accepted only for
// API compatibility, is completely ignored, and is not validated (any value, including nil, is
// accepted and unused).
//
// To decrypt ciphertext written by pre-#18 go-generics (fixed-nonce, no prefix wire format),
// callers must pass WithLegacyNonce(theirNonce) explicitly.
func NewWithPasswordNonce[T any](password, _ /*nonce*/, salt []byte, iterations int, opts ...Option) (*EncryptorDecryptor[T], error) {
	return NewWithPassword[T](password, salt, iterations, opts...)
}

func (e *EncryptorDecryptor[T]) Encrypt(value T) ([]byte, error) {
	w := wrapper[T]{T: value}

	bytes, err := serialize.NewSerializer[wrapper[T]]().FromDynamicType(w).ToBytes()
	if err != nil {
		return nil, err
	}

	if e.legacyNonce != nil {
		// Legacy wire format: fixed nonce, not prepended. Produces [ciphertext|tag].
		return e.aesgcm.Seal(nil, e.legacyNonce, bytes, nil), nil
	}

	nonce := make([]byte, e.aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}

	// Seal appends ciphertext+tag to the first argument (nonce), producing [nonce|ciphertext|tag]
	return e.aesgcm.Seal(nonce, nonce, bytes, nil), nil
}

func (e *EncryptorDecryptor[T]) Decrypt(ciphertext []byte) (T, error) {
	var t T

	var nonce, ct []byte
	if e.legacyNonce != nil {
		// Legacy wire format: ciphertext is [ciphertext|tag]; nonce is the fixed one.
		minLen := e.aesgcm.Overhead()
		if len(ciphertext) < minLen {
			return t, fmt.Errorf("%w: expected at least %d bytes, got %d", ErrCiphertextTooShort, minLen, len(ciphertext))
		}
		nonce, ct = e.legacyNonce, ciphertext
	} else {
		nonceSize := e.aesgcm.NonceSize()
		minLen := nonceSize + e.aesgcm.Overhead()
		if len(ciphertext) < minLen {
			return t, fmt.Errorf("%w: expected at least %d bytes, got %d", ErrCiphertextTooShort, minLen, len(ciphertext))
		}
		nonce, ct = ciphertext[:nonceSize], ciphertext[nonceSize:]
	}

	decryptedBytes, err := e.aesgcm.Open(nil, nonce, ct, nil)
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
