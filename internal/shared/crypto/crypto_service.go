package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// CryptoService provides encryption, decryption, and HMAC operations for PII protection.
//
// Key responsibilities:
// - Encrypt sensitive data (email, name, provider IDs) using AES-256-GCM
// - Decrypt encrypted data
// - Generate deterministic HMAC-SHA256 hashes for database lookups
// - Normalize input before HMAC to ensure case-insensitive, whitespace-agnostic matching
//
// Security properties:
// - Authenticated encryption via GCM mode (confidentiality + integrity)
// - Random nonce per encryption (prevents pattern analysis)
// - HMAC uses secret key (not reversible like plain hashes)
type CryptoService struct {
	keyManager KeyManager
}

// NewCryptoService creates a new cryptographic service with the provided key manager.
func NewCryptoService(keyManager KeyManager) *CryptoService {
	return &CryptoService{
		keyManager: keyManager,
	}
}

// Encrypt encrypts plaintext using AES-256-GCM.
// Returns base64-encoded string in format: "<nonce>:<ciphertext>"
//
// The nonce is randomly generated per encryption (required for GCM security).
// The GCM authentication tag is appended to the ciphertext automatically.
//
// Example output: "a3f9e8c7b6d5a4e3:9x8w7v6u5t4s3r2q..."
func (cs *CryptoService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil // Allow empty - caller handles null semantics
	}

	key, err := cs.keyManager.GetEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce (96 bits / 12 bytes for GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and append authentication tag
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Format: <base64_nonce>:<base64_ciphertext>
	encoded := fmt.Sprintf("%s:%s",
		base64.StdEncoding.EncodeToString(nonce),
		base64.StdEncoding.EncodeToString(ciphertext),
	)

	return encoded, nil
}

// Decrypt decrypts a ciphertext encrypted by Encrypt().
// Returns an error if the ciphertext is malformed or authentication fails.
func (cs *CryptoService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil // Allow empty - caller handles null semantics
	}

	// Parse format: <nonce>:<ciphertext>
	parts := strings.SplitN(ciphertext, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid ciphertext format (expected <nonce>:<ciphertext>)")
	}

	// Get the encryption key
	key, err := cs.keyManager.GetEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get decryption key: %w", err)
	}

	// Decode nonce and ciphertext from base64
	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}

	encryptedData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Create cipher and GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt and verify authentication tag
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed (wrong key or corrupted data): %w", err)
	}

	return string(plaintext), nil
}

// HMAC generates a deterministic HMAC-SHA256 hash for database lookup.
// The input is normalized (lowercased and trimmed) before hashing to ensure
// case-insensitive and whitespace-agnostic matching.
//
// Example: "  User@Example.COM  " → same HMAC as "user@example.com"
//
// Returns hex-encoded hash (64 characters).
// Use this for email_lookup_hash, provider_id_lookup_hash fields.
func (cs *CryptoService) HMAC(value string) (string, error) {
	if value == "" {
		return "", nil // Allow empty - caller handles null semantics
	}

	// Normalize: trim whitespace and lowercase
	normalized := strings.ToLower(strings.TrimSpace(value))

	hmacKey, err := cs.keyManager.GetHMACKey()
	if err != nil {
		return "", fmt.Errorf("failed to get HMAC key: %w", err)
	}

	h := hmac.New(sha256.New, hmacKey)
	h.Write([]byte(normalized))
	hash := h.Sum(nil)

	return fmt.Sprintf("%x", hash), nil
}
