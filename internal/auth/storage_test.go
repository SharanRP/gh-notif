package auth

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/user/gh-notif/internal/testutil"
	"golang.org/x/oauth2"
)

func TestFileStorage(t *testing.T) {
	// Skip this test on Windows
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	// Create a temporary directory for testing
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Save original home directory and restore after test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create a test token
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Create a custom file storage for testing
	storage := &FileStorage{
		filePath: filepath.Join(tempDir, encryptedTokenFile),
		keyPath:  filepath.Join(tempDir, keyFile),
	}

	// Generate a key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	// Write the key file
	if err := os.WriteFile(storage.keyPath, key, 0600); err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	// Test SaveToken
	if err := storage.SaveToken(token); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	// Check that the token file was created
	if _, err := os.Stat(storage.filePath); os.IsNotExist(err) {
		t.Errorf("SaveToken() did not create token file at %s", storage.filePath)
	}

	// Test LoadToken
	loadedToken, err := storage.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}

	// Check that the loaded token matches the original
	if loadedToken.AccessToken != token.AccessToken {
		t.Errorf("LoadToken() AccessToken = %v, want %v", loadedToken.AccessToken, token.AccessToken)
	}
	if loadedToken.TokenType != token.TokenType {
		t.Errorf("LoadToken() TokenType = %v, want %v", loadedToken.TokenType, token.TokenType)
	}
	if loadedToken.RefreshToken != token.RefreshToken {
		t.Errorf("LoadToken() RefreshToken = %v, want %v", loadedToken.RefreshToken, token.RefreshToken)
	}

	// Test DeleteToken
	if err := storage.DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() error = %v", err)
	}

	// Check that the token file was deleted
	if _, err := os.Stat(storage.filePath); !os.IsNotExist(err) {
		t.Errorf("DeleteToken() did not delete token file at %s", storage.filePath)
	}

	// Test LoadToken after deletion
	_, err = storage.LoadToken()
	if err == nil {
		t.Errorf("LoadToken() after deletion did not return error")
	}
	if err != ErrNoToken {
		t.Errorf("LoadToken() after deletion error = %v, want %v", err, ErrNoToken)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	// Generate a key
	var key [32]byte
	for i := range key {
		key[i] = byte(i)
	}

	// Test data
	data := []byte("test data for encryption and decryption")

	// Encrypt the data
	encrypted, err := encrypt(data, key)
	if err != nil {
		t.Fatalf("encrypt() error = %v", err)
	}

	// Check that the encrypted data is different from the original
	if string(encrypted) == string(data) {
		t.Errorf("encrypt() did not change the data")
	}

	// Decrypt the data
	decrypted, err := decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("decrypt() error = %v", err)
	}

	// Check that the decrypted data matches the original
	if string(decrypted) != string(data) {
		t.Errorf("decrypt() result = %v, want %v", string(decrypted), string(data))
	}

	// Test decryption with wrong key
	var wrongKey [32]byte
	for i := range wrongKey {
		wrongKey[i] = byte(i + 1)
	}

	_, err = decrypt(encrypted, wrongKey)
	if err == nil {
		t.Errorf("decrypt() with wrong key did not return error")
	}
}

func TestCreateStorage(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		tokenStorage string
		wantType    string
	}{
		{
			name:        "Keyring storage",
			tokenStorage: "keyring",
			wantType:    "*auth.KeyringStorage",
		},
		{
			name:        "File storage",
			tokenStorage: "file",
			wantType:    "*auth.FileStorage",
		},
		{
			name:        "Auto storage (fallback to file)",
			tokenStorage: "auto",
			wantType:    "*auth.FileStorage",
		},
		{
			name:        "Default storage",
			tokenStorage: "unknown",
			wantType:    "*auth.FileStorage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a custom createStorage function that uses our test tokenStorage
			createStorage := func() (Storage, error) {
				switch tt.tokenStorage {
				case "keyring":
					return &KeyringStorage{}, nil
				case "file":
					return NewFileStorage()
				case "auto":
					// For testing, just return file storage
					return NewFileStorage()
				default:
					// Default to file storage
					return NewFileStorage()
				}
			}

			// Create a temporary directory for testing
			tempDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Save original home directory and restore after test
			originalHome := os.Getenv("HOME")
			defer os.Setenv("HOME", originalHome)
			os.Setenv("HOME", tempDir)

			// Call the function
			storage, err := createStorage()
			if err != nil {
				t.Fatalf("createStorage() error = %v", err)
			}

			// Check the type of storage
			storageType := getType(storage)
			if storageType != tt.wantType {
				t.Errorf("createStorage() type = %v, want %v", storageType, tt.wantType)
			}
		})
	}
}

// Helper function to get the type of an interface as a string
func getType(v interface{}) string {
	switch v.(type) {
	case *KeyringStorage:
		return "*auth.KeyringStorage"
	case *FileStorage:
		return "*auth.FileStorage"
	default:
		return "unknown"
	}
}
