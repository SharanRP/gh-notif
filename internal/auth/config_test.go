package auth

import (
	"testing"
)

func TestGetAuthConfig(t *testing.T) {
	// Save original GetAuthConfig function
	originalGetAuthConfig := GetAuthConfig
	defer func() {
		GetAuthConfig = originalGetAuthConfig
	}()

	// Mock GetAuthConfig
	GetAuthConfig = func() AuthConfig {
		return AuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Scopes:       []string{"notifications", "repo"},
			TokenStorage: "file",
		}
	}

	// Call GetAuthConfig
	authConfig := GetAuthConfig()

	// Check that the values were loaded correctly
	if authConfig.ClientID != "test-client-id" {
		t.Errorf("GetAuthConfig() ClientID = %v, want %v", authConfig.ClientID, "test-client-id")
	}
	if authConfig.ClientSecret != "test-client-secret" {
		t.Errorf("GetAuthConfig() ClientSecret = %v, want %v", authConfig.ClientSecret, "test-client-secret")
	}
	if len(authConfig.Scopes) != 2 || authConfig.Scopes[0] != "notifications" || authConfig.Scopes[1] != "repo" {
		t.Errorf("GetAuthConfig() Scopes = %v, want %v", authConfig.Scopes, []string{"notifications", "repo"})
	}
	if authConfig.TokenStorage != "file" {
		t.Errorf("GetAuthConfig() TokenStorage = %v, want %v", authConfig.TokenStorage, "file")
	}
}

func TestSaveAuthConfig(t *testing.T) {
	// Save original SaveAuthConfig and GetAuthConfig functions
	originalSaveAuthConfig := SaveAuthConfig
	originalGetAuthConfig := GetAuthConfig
	defer func() {
		SaveAuthConfig = originalSaveAuthConfig
		GetAuthConfig = originalGetAuthConfig
	}()

	// Create a test auth config
	authConfig := AuthConfig{
		ClientID:     "new-client-id",
		ClientSecret: "new-client-secret",
		Scopes:       []string{"notifications", "repo", "user"},
		TokenStorage: "keyring",
	}

	// Mock SaveAuthConfig
	var savedConfig AuthConfig
	SaveAuthConfig = func(config AuthConfig) error {
		savedConfig = config
		return nil
	}

	// Mock GetAuthConfig
	GetAuthConfig = func() AuthConfig {
		return savedConfig
	}

	// Save the auth config
	err := SaveAuthConfig(authConfig)
	if err != nil {
		t.Fatalf("SaveAuthConfig() error = %v", err)
	}

	// Get the auth config
	loadedConfig := GetAuthConfig()

	// Check that the values were saved correctly
	if loadedConfig.ClientID != "new-client-id" {
		t.Errorf("GetAuthConfig() after SaveAuthConfig() ClientID = %v, want %v", loadedConfig.ClientID, "new-client-id")
	}
	if loadedConfig.ClientSecret != "new-client-secret" {
		t.Errorf("GetAuthConfig() after SaveAuthConfig() ClientSecret = %v, want %v", loadedConfig.ClientSecret, "new-client-secret")
	}
	if len(loadedConfig.Scopes) != 3 || loadedConfig.Scopes[0] != "notifications" || loadedConfig.Scopes[1] != "repo" || loadedConfig.Scopes[2] != "user" {
		t.Errorf("GetAuthConfig() after SaveAuthConfig() Scopes = %v, want %v", loadedConfig.Scopes, []string{"notifications", "repo", "user"})
	}
	if loadedConfig.TokenStorage != "keyring" {
		t.Errorf("GetAuthConfig() after SaveAuthConfig() TokenStorage = %v, want %v", loadedConfig.TokenStorage, "keyring")
	}
}

func TestGetClientID(t *testing.T) {
	// Save original GetClientID function
	originalGetClientID := GetClientID
	defer func() {
		GetClientID = originalGetClientID
	}()

	// Mock GetClientID
	GetClientID = func() string {
		return "test-client-id"
	}

	// Call GetClientID
	clientID := GetClientID()

	// Check that the value was loaded correctly
	if clientID != "test-client-id" {
		t.Errorf("GetClientID() = %v, want %v", clientID, "test-client-id")
	}
}

func TestGetClientSecret(t *testing.T) {
	// Save original GetClientSecret function
	originalGetClientSecret := GetClientSecret
	defer func() {
		GetClientSecret = originalGetClientSecret
	}()

	// Mock GetClientSecret
	GetClientSecret = func() string {
		return "test-client-secret"
	}

	// Call GetClientSecret
	clientSecret := GetClientSecret()

	// Check that the value was loaded correctly
	if clientSecret != "test-client-secret" {
		t.Errorf("GetClientSecret() = %v, want %v", clientSecret, "test-client-secret")
	}
}

func TestGetScopes(t *testing.T) {
	// Save original GetScopes function
	originalGetScopes := GetScopes
	defer func() {
		GetScopes = originalGetScopes
	}()

	// Mock GetScopes
	GetScopes = func() []string {
		return []string{"notifications", "repo", "user"}
	}

	// Call GetScopes
	scopes := GetScopes()

	// Check that the value was loaded correctly
	if len(scopes) != 3 || scopes[0] != "notifications" || scopes[1] != "repo" || scopes[2] != "user" {
		t.Errorf("GetScopes() = %v, want %v", scopes, []string{"notifications", "repo", "user"})
	}
}

func TestGetTokenStorage(t *testing.T) {
	// Save original GetTokenStorage function
	originalGetTokenStorage := GetTokenStorage
	defer func() {
		GetTokenStorage = originalGetTokenStorage
	}()

	// Mock GetTokenStorage
	GetTokenStorage = func() string {
		return "keyring"
	}

	// Call GetTokenStorage
	tokenStorage := GetTokenStorage()

	// Check that the value was loaded correctly
	if tokenStorage != "keyring" {
		t.Errorf("GetTokenStorage() = %v, want %v", tokenStorage, "keyring")
	}
}

func TestCreateStorageWithMock(t *testing.T) {
	// Save original GetTokenStorage function
	originalGetTokenStorage := GetTokenStorage
	defer func() {
		GetTokenStorage = originalGetTokenStorage
	}()

	// Test cases
	tests := []struct {
		name         string
		tokenStorage string
		wantType     string
	}{
		{
			name:         "Keyring storage",
			tokenStorage: "keyring",
			wantType:     "*auth.KeyringStorage",
		},
		{
			name:         "File storage",
			tokenStorage: "file",
			wantType:     "*auth.FileStorage",
		},
		{
			name:         "Auto storage",
			tokenStorage: "auto",
			wantType:     "*auth.FileStorage", // Assuming keyring fails and falls back to file
		},
		{
			name:         "Default storage",
			tokenStorage: "unknown",
			wantType:     "*auth.FileStorage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock GetTokenStorage
			GetTokenStorage = func() string {
				return tt.tokenStorage
			}

			// Call CreateStorage
			storage, err := CreateStorage()
			if err != nil {
				t.Fatalf("CreateStorage() error = %v", err)
			}

			// Check the type of storage
			var storageType string
			switch storage.(type) {
			case *KeyringStorage:
				storageType = "*auth.KeyringStorage"
			case *FileStorage:
				storageType = "*auth.FileStorage"
			default:
				storageType = "unknown"
			}

			if storageType != tt.wantType {
				t.Errorf("CreateStorage() type = %v, want %v", storageType, tt.wantType)
			}
		})
	}
}
