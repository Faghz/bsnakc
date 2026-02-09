package domain

// PasswordHasher defines the interface for password hashing operations
// This is an output port that should be implemented by infrastructure adapters
type PasswordHasher interface {
	// Hash generates a hash from a plain text password
	Hash(password string) (string, error)

	// Verify compares a plain text password with a hashed password
	Verify(password, hash string) (bool, error)
}
