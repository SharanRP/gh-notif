package auth

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/oauth2"
)

const (
	// Service name for keyring
	serviceName = "gh-notif"
	// Username for keyring
	username = "github-user"
	// File name for encrypted token
	encryptedTokenFile = ".gh-notif-token.enc"
	// Key file name
	keyFile = ".gh-notif-key"
)

var (
	// ErrNoToken is returned when no token is found
	ErrNoToken = errors.New("no token found")
	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")
)

// Storage interface for token storage
type Storage interface {
	SaveToken(token *oauth2.Token) error
	LoadToken() (*oauth2.Token, error)
	DeleteToken() error
}

// KeyringStorage implements Storage using the system keyring
type KeyringStorage struct{}

// SaveToken saves the token to the system keyring
func (s *KeyringStorage) SaveToken(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	err = keyring.Set(serviceName, username, string(data))
	if err != nil {
		return fmt.Errorf("failed to save token to keyring: %w", err)
	}

	return nil
}

// LoadToken loads the token from the system keyring
func (s *KeyringStorage) LoadToken() (*oauth2.Token, error) {
	data, err := keyring.Get(serviceName, username)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, ErrNoToken
		}
		return nil, fmt.Errorf("failed to load token from keyring: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

// DeleteToken deletes the token from the system keyring
func (s *KeyringStorage) DeleteToken() error {
	err := keyring.Delete(serviceName, username)
	if err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("failed to delete token from keyring: %w", err)
	}
	return nil
}

// FileStorage implements Storage using encrypted files
type FileStorage struct {
	keyPath  string
	filePath string
}

// NewFileStorage creates a new FileStorage
func NewFileStorage() (*FileStorage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &FileStorage{
		keyPath:  filepath.Join(home, keyFile),
		filePath: filepath.Join(home, encryptedTokenFile),
	}, nil
}

// SaveToken saves the token to an encrypted file
func (s *FileStorage) SaveToken(token *oauth2.Token) error {
	// Marshal token to JSON
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Get or create encryption key
	key, err := s.getOrCreateKey()
	if err != nil {
		return err
	}

	// Encrypt the token
	encrypted, err := encrypt(data, key)
	if err != nil {
		return err
	}

	// Save the encrypted token
	err = os.WriteFile(s.filePath, encrypted, 0600)
	if err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// LoadToken loads the token from an encrypted file
func (s *FileStorage) LoadToken() (*oauth2.Token, error) {
	// Check if the file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return nil, ErrNoToken
	}

	// Read the encrypted token
	encrypted, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	// Get the encryption key
	key, err := s.getKey()
	if err != nil {
		return nil, err
	}

	// Decrypt the token
	data, err := decrypt(encrypted, key)
	if err != nil {
		return nil, err
	}

	// Unmarshal the token
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

// DeleteToken deletes the token file
func (s *FileStorage) DeleteToken() error {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(s.filePath)
}

// getOrCreateKey gets the encryption key or creates a new one
func (s *FileStorage) getOrCreateKey() ([32]byte, error) {
	var key [32]byte

	// Check if the key file exists
	if _, err := os.Stat(s.keyPath); os.IsNotExist(err) {
		// Generate a new key
		if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
			return key, fmt.Errorf("failed to generate key: %w", err)
		}

		// Save the key
		if err := os.WriteFile(s.keyPath, key[:], 0600); err != nil {
			return key, fmt.Errorf("failed to write key file: %w", err)
		}

		return key, nil
	}

	// Read the key
	keyData, err := os.ReadFile(s.keyPath)
	if err != nil {
		return key, fmt.Errorf("failed to read key file: %w", err)
	}

	// Check key length
	if len(keyData) != 32 {
		return key, fmt.Errorf("invalid key length: %d", len(keyData))
	}

	copy(key[:], keyData)
	return key, nil
}

// getKey gets the encryption key
func (s *FileStorage) getKey() ([32]byte, error) {
	var key [32]byte

	// Check if the key file exists
	if _, err := os.Stat(s.keyPath); os.IsNotExist(err) {
		return key, fmt.Errorf("key file not found")
	}

	// Read the key
	keyData, err := os.ReadFile(s.keyPath)
	if err != nil {
		return key, fmt.Errorf("failed to read key file: %w", err)
	}

	// Check key length
	if len(keyData) != 32 {
		return key, fmt.Errorf("invalid key length: %d", len(keyData))
	}

	copy(key[:], keyData)
	return key, nil
}

// encrypt encrypts data using NaCl secretbox
func encrypt(data []byte, key [32]byte) ([]byte, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encrypted := secretbox.Seal(nonce[:], data, &nonce, &key)
	return encrypted, nil
}

// decrypt decrypts data using NaCl secretbox
func decrypt(encrypted []byte, key [32]byte) ([]byte, error) {
	if len(encrypted) < 24 {
		return nil, fmt.Errorf("encrypted data too short")
	}

	var nonce [24]byte
	copy(nonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, &key)
	if !ok {
		return nil, fmt.Errorf("decryption failed")
	}

	return decrypted, nil
}
