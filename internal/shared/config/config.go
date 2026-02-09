package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	PASETO   PASETOConfig
	OAuth    OAuthConfig
	CORS     CORSConfig
	Crypto   CryptoConfig
	OTel     OTelConfig
}

type ServerConfig struct {
	Host          string
	Port          string
	Env           string
	LogLevel      string
	LogStackLevel string // Log level at which to automatically capture stack traces (empty = disabled)
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type PASETOConfig struct {
	SymmetricKey         string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type OAuthConfig struct {
	Discord DiscordConfig
}

type DiscordConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type CryptoConfig struct {
	// User domain keys
	UserEncryptionKey string
	UserHMACSecret    string
	// Identity domain keys
	IdentityEncryptionKey string
	IdentityHMACSecret    string
}

type OTelLogsConfig struct {
	Enabled      bool
	ExporterType string // "otlp", "console"
	OTLPEndpoint string // Falls back to parent OTLPEndpoint if empty
}

type OTelConfig struct {
	Enabled        bool
	ServiceName    string
	ServiceVersion string
	Environment    string
	ExporterType   string // "otlp", "console"
	OTLPEndpoint   string
	OTLPInsecure   bool
	SampleRate     float64 // 0.0 to 1.0
	Logs           OTelLogsConfig
}

// Load reads configuration from environment variables
// Attempts to load .env file for local development (fails silently if not found)
func Load() (*Config, error) {
	// Load .env file if it exists - ignore error if file doesn't exist
	// This is useful for local development
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host:          getEnv("SERVER_HOST", "0.0.0.0"),
			Port:          getEnv("SERVER_PORT", "8080"),
			Env:           getEnv("SERVER_ENV", "development"),
			LogLevel:      getEnv("LOG_LEVEL", ""),       // empty = environment-based default
			LogStackLevel: getEnv("LOG_STACK_LEVEL", ""), // empty = no automatic stack traces
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "marshal"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		PASETO: PASETOConfig{
			SymmetricKey:         getEnv("PASETO_SYMMETRIC_KEY", ""),
			AccessTokenDuration:  getEnvAsDuration("PASETO_ACCESS_TOKEN_DURATION", 15*time.Minute),
			RefreshTokenDuration: getEnvAsDuration("PASETO_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
		},
		OAuth: OAuthConfig{
			Discord: DiscordConfig{
				ClientID:     getEnv("DISCORD_CLIENT_ID", ""),
				ClientSecret: getEnv("DISCORD_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("DISCORD_REDIRECT_URL", ""),
			},
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		},
		Crypto: CryptoConfig{
			UserEncryptionKey:     getEnv("USER_ENCRYPTION_KEY", ""),
			UserHMACSecret:        getEnv("USER_HMAC_SECRET", ""),
			IdentityEncryptionKey: getEnv("IDENTITY_ENCRYPTION_KEY", ""),
			IdentityHMACSecret:    getEnv("IDENTITY_HMAC_SECRET", ""),
		},
		OTel: OTelConfig{
			Enabled:        getEnvAsBool("OTEL_ENABLED", false),
			ServiceName:    getEnv("OTEL_SERVICE_NAME", "marshal"),
			ServiceVersion: getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
			Environment:    getEnv("OTEL_ENVIRONMENT", getEnv("SERVER_ENV", "development")),
			ExporterType:   getEnv("OTEL_EXPORTER_TYPE", "console"),
			OTLPEndpoint:   getEnv("OTEL_OTLP_ENDPOINT", "localhost:4317"),
			OTLPInsecure:   getEnvAsBool("OTEL_OTLP_INSECURE", true),
			SampleRate:     getEnvAsFloat("OTEL_SAMPLE_RATE", 1.0),
			Logs: OTelLogsConfig{
				Enabled:      getEnvAsBool("OTEL_LOGS_ENABLED", false),
				ExporterType: getEnv("OTEL_LOGS_EXPORTER_TYPE", "otlp"),
				OTLPEndpoint: getEnv("OTEL_LOGS_OTLP_ENDPOINT", ""),
			},
		},
	}

	// Validate PASETO configuration
	if err := cfg.PASETO.Validate(); err != nil {
		return nil, fmt.Errorf("PASETO config validation failed: %w", err)
	}

	// Validate crypto configuration
	if err := cfg.Crypto.Validate(); err != nil {
		return nil, fmt.Errorf("crypto config validation failed: %w", err)
	}

	return cfg, nil
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// Address returns the server address
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Address returns the Redis address
func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Validate checks if PASETO configuration is valid
func (c *PASETOConfig) Validate() error {
	if c.SymmetricKey == "" {
		return fmt.Errorf("PASETO_SYMMETRIC_KEY is required (generate with: openssl rand -hex 32)")
	}

	// Verify it's valid hex and correct length
	keyBytes, err := hex.DecodeString(c.SymmetricKey)
	if err != nil {
		return fmt.Errorf("PASETO_SYMMETRIC_KEY must be hex-encoded: %w", err)
	}

	if len(keyBytes) != 32 {
		return fmt.Errorf("PASETO_SYMMETRIC_KEY must be 32 bytes (64 hex chars), got %d bytes", len(keyBytes))
	}

	return nil
}

// Validate checks if crypto configuration is valid
func (c *CryptoConfig) Validate() error {
	// Validate user domain keys
	if c.UserEncryptionKey == "" {
		return fmt.Errorf("USER_ENCRYPTION_KEY is required (generate with: openssl rand -hex 32)")
	}

	if c.UserHMACSecret == "" {
		return fmt.Errorf("USER_HMAC_SECRET is required (generate with: openssl rand -hex 32)")
	}

	// Validate identity domain keys
	if c.IdentityEncryptionKey == "" {
		return fmt.Errorf("IDENTITY_ENCRYPTION_KEY is required (generate with: openssl rand -hex 32)")
	}

	if c.IdentityHMACSecret == "" {
		return fmt.Errorf("IDENTITY_HMAC_SECRET is required (generate with: openssl rand -hex 32)")
	}

	return nil
}

// GetLogsEndpoint returns the effective endpoint for log export
// Falls back to OTLPEndpoint if Logs.OTLPEndpoint is not set
func (c *OTelConfig) GetLogsEndpoint() string {
	if c.Logs.OTLPEndpoint != "" {
		return c.Logs.OTLPEndpoint
	}
	return c.OTLPEndpoint
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var result []string
	current := ""
	for _, char := range valueStr {
		if char == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}

	return result
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}
