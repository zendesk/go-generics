package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	defaultIterations = 4096 // number of PBKDF2 iterations, this can be increased for more security
	salt              = "73616C7479"
)

type EncryptorDecryptor struct {
	aesgcm cipher.AEAD
	nonce  []byte
}

func New() (*EncryptorDecryptor, error) {
	return NewWithIterations(defaultIterations)
}

func NewWithIterations(iterations int) (*EncryptorDecryptor, error) {
	password := make([]byte, 2048)
	if _, err := io.ReadFull(rand.Reader, password); err != nil {
		return nil, fmt.Errorf("error generating password: %w", err)
	}

	aesKey := pbkdf2.Key(password, []byte(salt), iterations, 32, sha256.New)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error creating new GCM: %w", err)
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}

	e := &EncryptorDecryptor{
		aesgcm: aesgcm,
		nonce:  nonce,
	}

	return e, nil
}

func NewWithPasswordNonceIterations(password []byte, nonce []byte, iterations int) (*EncryptorDecryptor, error) {
	aesKey := pbkdf2.Key(password, []byte(salt), iterations, 32, sha256.New)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error creating new GCM: %w", err)
	}

	e := &EncryptorDecryptor{
		aesgcm: aesgcm,
		nonce:  nonce,
	}

	return e, nil
}

func (e *EncryptorDecryptor) Encrypt(value []byte) ([]byte, error) {
	ciphertext := e.aesgcm.Seal(nil, e.nonce, value, nil)

	return ciphertext, nil
}

func (e *EncryptorDecryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	plaintext, err := e.aesgcm.Open(nil, e.nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting ciphertext: %w", err)
	}

	return plaintext, nil
}
