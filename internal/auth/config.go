package auth

import (
	"github.com/user/gh-notif/internal/config"
)

// AuthConfig holds the authentication configuration
type AuthConfig struct {
	// ClientID is the GitHub OAuth client ID
	ClientID string

	// ClientSecret is the GitHub OAuth client secret
	ClientSecret string

	// Scopes are the OAuth scopes to request
	Scopes []string

	// TokenStorage defines how to store the OAuth token
	// Options: "keyring", "file", "auto"
	TokenStorage string
}

// GetAuthConfigFunc is the function type for GetAuthConfig
type GetAuthConfigFunc func() AuthConfig

// GetAuthConfig returns the authentication configuration from the global config
var GetAuthConfig GetAuthConfigFunc = func() AuthConfig {
	// Create a new config manager
	cm := config.NewConfigManager()
	if err := cm.Load(); err != nil {
		// Return default config if we can't load the config
		return AuthConfig{
			Scopes:       []string{"notifications", "repo"},
			TokenStorage: "auto",
		}
	}

	// Get the config
	cfg := cm.GetConfig()

	// Return the auth config
	return AuthConfig{
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
		Scopes:       cfg.Auth.Scopes,
		TokenStorage: cfg.Auth.TokenStorage,
	}
}

// SaveAuthConfigFunc is the function type for SaveAuthConfig
type SaveAuthConfigFunc func(AuthConfig) error

// SaveAuthConfig saves the authentication configuration to the global config
var SaveAuthConfig SaveAuthConfigFunc = func(authConfig AuthConfig) error {
	// Create a new config manager
	cm := config.NewConfigManager()
	if err := cm.Load(); err != nil {
		return err
	}

	// Get the config
	cfg := cm.GetConfig()

	// Update the auth config
	cfg.Auth.ClientID = authConfig.ClientID
	cfg.Auth.ClientSecret = authConfig.ClientSecret
	cfg.Auth.Scopes = authConfig.Scopes
	cfg.Auth.TokenStorage = authConfig.TokenStorage

	// Save the config
	return cm.Save()
}

// GetClientIDFunc is the function type for GetClientID
type GetClientIDFunc func() string

// GetClientID returns the GitHub OAuth client ID
var GetClientID GetClientIDFunc = func() string {
	return GetAuthConfig().ClientID
}

// GetClientSecretFunc is the function type for GetClientSecret
type GetClientSecretFunc func() string

// GetClientSecret returns the GitHub OAuth client secret
var GetClientSecret GetClientSecretFunc = func() string {
	return GetAuthConfig().ClientSecret
}

// GetScopesFunc is the function type for GetScopes
type GetScopesFunc func() []string

// GetScopes returns the OAuth scopes to request
var GetScopes GetScopesFunc = func() []string {
	return GetAuthConfig().Scopes
}

// GetTokenStorageFunc is the function type for GetTokenStorage
type GetTokenStorageFunc func() string

// GetTokenStorage returns the token storage method
var GetTokenStorage GetTokenStorageFunc = func() string {
	return GetAuthConfig().TokenStorage
}

// CreateStorage creates a storage implementation based on the configuration
func CreateStorage() (Storage, error) {
	tokenStorage := GetTokenStorage()

	switch tokenStorage {
	case "keyring":
		return &KeyringStorage{}, nil
	case "file":
		return NewFileStorage()
	case "auto":
		// Try keyring first
		keyringStorage := &KeyringStorage{}
		if _, err := keyringStorage.LoadToken(); err == nil {
			return keyringStorage, nil
		}

		// Fall back to file storage
		return NewFileStorage()
	default:
		// Default to file storage
		return NewFileStorage()
	}
}
