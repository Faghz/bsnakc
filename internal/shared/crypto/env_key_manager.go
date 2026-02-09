package crypto

import (
	"encoding/hex"
	"fmt"

	"github.com/elzestia/go-boilerplate/internal/shared/config"
)

// EnvKeyManager implements KeyManager by reading keys from environment variables.
// Keys are expected as hex-encoded strings (e.g., from `openssl rand -hex 32`).
// Each instance is scoped to a specific domain (e.g., "user", "identity").
type EnvKeyManager struct {
	encryptionKey []byte
	hmacKey       []byte
	domain        string
}

// NewEnvKeyManager creates a domain-scoped KeyManager that reads from application config.
// domain should be "user" or "identity" to load the appropriate keys.
// Validates that all keys are exactly 32 bytes (AES-256/HMAC-SHA256 requirement).
func NewEnvKeyManager(cfg *config.Config, domain string) (*EnvKeyManager, error) {
	km := &EnvKeyManager{domain: domain}

	var encKeyStr, hmacSecretStr string

	// Load domain-specific keys
	switch domain {
	case "user":
		encKeyStr = cfg.Crypto.UserEncryptionKey
		hmacSecretStr = cfg.Crypto.UserHMACSecret
	case "identity":
		encKeyStr = cfg.Crypto.IdentityEncryptionKey
		hmacSecretStr = cfg.Crypto.IdentityHMACSecret
	default:
		return nil, fmt.Errorf("unsupported domain: %s (must be 'user' or 'identity')", domain)
	}

	// Load encryption key (required)
	if encKeyStr == "" {
		return nil, fmt.Errorf("%s encryption key is required", domain)
	}
	encKey, err := hex.DecodeString(encKeyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid %s encryption key hex encoding: %w", domain, err)
	}
	if len(encKey) != 32 {
		return nil, fmt.Errorf("%s encryption key must be 32 bytes (64 hex chars), got %d bytes", domain, len(encKey))
	}
	km.encryptionKey = encKey

	// Load HMAC secret (required)
	if hmacSecretStr == "" {
		return nil, fmt.Errorf("%s HMAC secret is required", domain)
	}
	hmacKey, err := hex.DecodeString(hmacSecretStr)
	if err != nil {
		return nil, fmt.Errorf("invalid %s HMAC secret hex encoding: %w", domain, err)
	}
	if len(hmacKey) != 32 {
		return nil, fmt.Errorf("%s HMAC secret must be 32 bytes (64 hex chars), got %d bytes", domain, len(hmacKey))
	}
	km.hmacKey = hmacKey

	return km, nil
}

// GetEncryptionKey returns the encryption key for this domain.
func (km *EnvKeyManager) GetEncryptionKey() ([]byte, error) {
	return km.encryptionKey, nil
}

// GetHMACKey returns the HMAC secret key for this domain.
func (km *EnvKeyManager) GetHMACKey() ([]byte, error) {
	return km.hmacKey, nil
}
