package ports

// UserInfo represents a simplified view of a User from another context
// This is used in the Anti-Corruption Layer to prevent domain entity leakage
type UserInfo struct {
	ID        int64
	UID       string
	Email     string
	Name      string
	Username  string
	AvatarURL string
}
