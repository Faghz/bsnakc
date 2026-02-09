package crypto

// KeyManager abstracts key storage and retrieval for encryption and HMAC operations.
// This interface allows switching between different key storage backends
// (e.g., environment variables, external key management services) without changing crypto logic.
type KeyManager interface {
	// GetEncryptionKey returns the AES-256 encryption key.
	// Key must be exactly 32 bytes for AES-256.
	GetEncryptionKey() ([]byte, error)

	// GetHMACKey returns the HMAC-SHA256 secret key used for deterministic lookup hashes.
	// This key must be kept secret and should be at least 32 bytes.
	GetHMACKey() ([]byte, error)
}
