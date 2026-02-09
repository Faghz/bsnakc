package ports

// CryptoService defines encryption and hashing operations for PII protection.
// This port interface allows for dependency inversion and enables testing with mocks.
//
// Implementations should provide:
//   - AES-256-GCM encryption for sensitive data at rest
//   - HMAC-SHA256 for deterministic lookup hashes
//   - Domain-specific key isolation (separate keys per domain)
type CryptoService interface {
	// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
	// Each encryption produces a different ciphertext for the same input (nonce varies).
	// Returns base64-encoded ciphertext or error if encryption fails.
	Encrypt(plaintext string) (string, error)

	// Decrypt decrypts AES-256-GCM ciphertext back to plaintext.
	// Returns decrypted string or error if decryption/authentication fails.
	Decrypt(ciphertext string) (string, error)

	// HMAC generates a deterministic HMAC-SHA256 hash for lookup purposes.
	// Same input always produces the same hash (used for database queries).
	// Returns hex-encoded HMAC hash.
	HMAC(value string) (string, error)
}
