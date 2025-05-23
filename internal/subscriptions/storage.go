package subscriptions

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// Storage interface for subscription persistence
type Storage interface {
	AddSubscription(subscription RepositorySubscription) error
	RemoveSubscription(repository string) error
	GetSubscription(repository string) (*RepositorySubscription, error)
	UpdateSubscription(subscription RepositorySubscription) error
	ListSubscriptions() ([]RepositorySubscription, error)
	ExportSubscriptions() (*SubscriptionList, error)
	ImportSubscriptions(list SubscriptionList) error
	BackupSubscriptions(path string) error
	RestoreSubscriptions(path string) error
}

// FileStorage implements Storage using encrypted files
type FileStorage struct {
	filePath string
	password string
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(filePath, password string) (*FileStorage, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FileStorage{
		filePath: filePath,
		password: password,
	}, nil
}

// AddSubscription adds a new subscription
func (fs *FileStorage) AddSubscription(subscription RepositorySubscription) error {
	list, err := fs.loadSubscriptions()
	if err != nil {
		return err
	}

	// Check if subscription already exists
	for i, existing := range list.Subscriptions {
		if existing.Repository == subscription.Repository {
			// Update existing subscription
			list.Subscriptions[i] = subscription
			list.UpdatedAt = time.Now()
			return fs.saveSubscriptions(list)
		}
	}

	// Add new subscription
	list.Subscriptions = append(list.Subscriptions, subscription)
	list.UpdatedAt = time.Now()
	return fs.saveSubscriptions(list)
}

// RemoveSubscription removes a subscription
func (fs *FileStorage) RemoveSubscription(repository string) error {
	list, err := fs.loadSubscriptions()
	if err != nil {
		return err
	}

	// Find and remove subscription
	for i, existing := range list.Subscriptions {
		if existing.Repository == repository {
			list.Subscriptions = append(list.Subscriptions[:i], list.Subscriptions[i+1:]...)
			list.UpdatedAt = time.Now()
			return fs.saveSubscriptions(list)
		}
	}

	return fmt.Errorf("subscription not found: %s", repository)
}

// GetSubscription retrieves a specific subscription
func (fs *FileStorage) GetSubscription(repository string) (*RepositorySubscription, error) {
	list, err := fs.loadSubscriptions()
	if err != nil {
		return nil, err
	}

	for _, subscription := range list.Subscriptions {
		if subscription.Repository == repository {
			return &subscription, nil
		}
	}

	return nil, fmt.Errorf("subscription not found: %s", repository)
}

// UpdateSubscription updates an existing subscription
func (fs *FileStorage) UpdateSubscription(subscription RepositorySubscription) error {
	list, err := fs.loadSubscriptions()
	if err != nil {
		return err
	}

	// Find and update subscription
	for i, existing := range list.Subscriptions {
		if existing.Repository == subscription.Repository {
			list.Subscriptions[i] = subscription
			list.UpdatedAt = time.Now()
			return fs.saveSubscriptions(list)
		}
	}

	return fmt.Errorf("subscription not found: %s", subscription.Repository)
}

// ListSubscriptions returns all subscriptions
func (fs *FileStorage) ListSubscriptions() ([]RepositorySubscription, error) {
	list, err := fs.loadSubscriptions()
	if err != nil {
		return nil, err
	}

	return list.Subscriptions, nil
}

// ExportSubscriptions exports all subscriptions
func (fs *FileStorage) ExportSubscriptions() (*SubscriptionList, error) {
	return fs.loadSubscriptions()
}

// ImportSubscriptions imports subscriptions from a list
func (fs *FileStorage) ImportSubscriptions(list SubscriptionList) error {
	return fs.saveSubscriptions(&list)
}

// BackupSubscriptions creates a backup of subscriptions
func (fs *FileStorage) BackupSubscriptions(path string) error {
	list, err := fs.loadSubscriptions()
	if err != nil {
		return err
	}

	// Create backup directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Save unencrypted backup
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal subscriptions: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

// RestoreSubscriptions restores subscriptions from a backup
func (fs *FileStorage) RestoreSubscriptions(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	var list SubscriptionList
	if err := json.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("failed to unmarshal subscriptions: %w", err)
	}

	return fs.saveSubscriptions(&list)
}

// loadSubscriptions loads subscriptions from encrypted file
func (fs *FileStorage) loadSubscriptions() (*SubscriptionList, error) {
	// Check if file exists
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		// Return empty list if file doesn't exist
		return &SubscriptionList{
			Version:       "1.0",
			Subscriptions: []RepositorySubscription{},
			UpdatedAt:     time.Now(),
		}, nil
	}

	// Read encrypted file
	encryptedData, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscriptions file: %w", err)
	}

	// Decrypt data
	data, err := fs.decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt subscriptions: %w", err)
	}

	// Unmarshal JSON
	var list SubscriptionList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscriptions: %w", err)
	}

	return &list, nil
}

// saveSubscriptions saves subscriptions to encrypted file
func (fs *FileStorage) saveSubscriptions(list *SubscriptionList) error {
	// Marshal to JSON
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal subscriptions: %w", err)
	}

	// Encrypt data
	encryptedData, err := fs.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt subscriptions: %w", err)
	}

	// Write to file
	return os.WriteFile(fs.filePath, encryptedData, 0600)
}

// encrypt encrypts data using AES-GCM
func (fs *FileStorage) encrypt(data []byte) ([]byte, error) {
	// Derive key from password
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key := pbkdf2.Key([]byte(fs.password), salt, 100000, 32, sha256.New)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Combine salt + nonce + ciphertext
	result := make([]byte, len(salt)+len(nonce)+len(ciphertext))
	copy(result, salt)
	copy(result[len(salt):], nonce)
	copy(result[len(salt)+len(nonce):], ciphertext)

	return result, nil
}

// decrypt decrypts data using AES-GCM
func (fs *FileStorage) decrypt(data []byte) ([]byte, error) {
	if len(data) < 32+12 { // salt + nonce minimum
		return nil, fmt.Errorf("invalid encrypted data")
	}

	// Extract salt, nonce, and ciphertext
	salt := data[:32]
	nonce := data[32:44]
	ciphertext := data[44:]

	// Derive key from password
	key := pbkdf2.Key([]byte(fs.password), salt, 100000, 32, sha256.New)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GetDefaultStoragePath returns the default storage path
func GetDefaultStoragePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".gh-notif-subscriptions.enc"), nil
}
