package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// KeyPrefix is the expected prefix for all Saber API keys.
	KeyPrefix = "sk_live_"
	// KeyLength is the total length of a valid API key (prefix + 43 chars).
	KeyLength = 51
)

// Credentials is persisted to ~/.saber/credentials.json.
type Credentials struct {
	APIKey    string    `json:"apiKey"`
	CreatedAt time.Time `json:"createdAt"`
}

// APIKey returns the active API key using the lookup order:
// 1. SABER_API_KEY env var
// 2. ~/.saber/credentials.json
// Returns ("", nil) if no key is configured.
func APIKey() (string, error) {
	if k := os.Getenv("SABER_API_KEY"); k != "" {
		return k, nil
	}
	creds, err := LoadCredentials()
	if err != nil {
		// File not found is not an error — just not authenticated.
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return creds.APIKey, nil
}

// RequireAPIKey returns the API key or exits with code 2 if not configured.
func RequireAPIKey() (string, error) {
	key, err := APIKey()
	if err != nil {
		return "", err
	}
	if key == "" {
		return "", &ErrNotAuthenticated{}
	}
	return key, nil
}

// ErrNotAuthenticated is returned when no API key is found.
type ErrNotAuthenticated struct{}

func (e *ErrNotAuthenticated) Error() string {
	return "not authenticated — run: saber auth login"
}

// ValidateKeyFormat checks the format of an API key without hitting the network.
func ValidateKeyFormat(key string) error {
	if !strings.HasPrefix(key, KeyPrefix) {
		return fmt.Errorf("API key must start with %q", KeyPrefix)
	}
	if len(key) != KeyLength {
		return fmt.Errorf("API key must be %d characters (got %d)", KeyLength, len(key))
	}
	return nil
}

// LoadCredentials reads ~/.saber/credentials.json.
func LoadCredentials() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("malformed credentials file: %w", err)
	}
	return &creds, nil
}

// SaveCredentials writes credentials to ~/.saber/credentials.json (mode 0600).
func SaveCredentials(key string) error {
	dir, err := saberDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	creds := Credentials{
		APIKey:    key,
		CreatedAt: time.Now().UTC(),
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "credentials.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}
	return nil
}

// DeleteCredentials removes ~/.saber/credentials.json. Idempotent.
func DeleteCredentials() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func saberDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".saber"), nil
}

func credentialsPath() (string, error) {
	dir, err := saberDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}
