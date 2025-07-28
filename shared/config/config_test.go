package config

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetConfig resets the singleton for testing
func resetConfig() {
	instance = nil
	once = sync.Once{}
}

func TestConfig_Load(t *testing.T) {
	t.Run("loads default values when no env vars set", func(t *testing.T) {
		// Clean up and reset
		os.Clearenv()
		resetConfig()

		cfg := Load()

		assert.NotNil(t, cfg)
		assert.Equal(t, "localhost:8080", cfg.HTTPPort)
		assert.Equal(t, "localhost:50051", cfg.GRPCPort)
		assert.Equal(t, "default-secret-key-change-in-production", cfg.JWTSecret)
		assert.Equal(t, "localhost", cfg.RedisHost)
		assert.Equal(t, "6379", cfg.RedisPort)
		assert.Equal(t, "", cfg.DatabaseURL)
	})

	t.Run("loads environment variables when set", func(t *testing.T) {
		// Clean up and reset
		os.Clearenv()
		resetConfig()

		// Set environment variables
		os.Setenv("HTTP_PORT", "192.168.1.100:9090")
		os.Setenv("GRPC_PORT", "192.168.1.100:50052")
		os.Setenv("MY_JWT_SECRET", "test-secret-key")
		os.Setenv("REDIS_HOST", "redis.example.com")
		os.Setenv("REDIS_PORT", "6380")
		os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb")

		cfg := Load()

		assert.Equal(t, "192.168.1.100:9090", cfg.HTTPPort)
		assert.Equal(t, "192.168.1.100:50052", cfg.GRPCPort)
		assert.Equal(t, "test-secret-key", cfg.JWTSecret)
		assert.Equal(t, "redis.example.com", cfg.RedisHost)
		assert.Equal(t, "6380", cfg.RedisPort)
		assert.Equal(t, "postgres://test:test@localhost:5432/testdb", cfg.DatabaseURL)

		// Clean up
		os.Clearenv()
	})

	t.Run("singleton pattern works correctly", func(t *testing.T) {
		// Reset instance
		resetConfig()

		cfg1 := Load()
		cfg2 := Load()
		cfg3 := Get()

		// All should return the same instance
		assert.Same(t, cfg1, cfg2)
		assert.Same(t, cfg1, cfg3)
	})
}

func TestConfig_Get(t *testing.T) {
	// Reset instance
	resetConfig()

	// First call should load configuration
	cfg := Get()
	assert.NotNil(t, cfg)

	// Second call should return same instance
	cfg2 := Get()
	assert.Same(t, cfg, cfg2)
}

func TestConfig_IsProduction(t *testing.T) {
	t.Run("returns false when ENVIRONMENT is not set", func(t *testing.T) {
		os.Clearenv()
		resetConfig()
		cfg := Load()
		assert.False(t, cfg.IsProduction())
	})

	t.Run("returns false when ENVIRONMENT is development", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENVIRONMENT", "development")
		resetConfig()
		cfg := Load()
		assert.False(t, cfg.IsProduction())
	})

	t.Run("returns true when ENVIRONMENT is production", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENVIRONMENT", "production")
		resetConfig()
		cfg := Load()
		assert.True(t, cfg.IsProduction())
	})

	// Clean up
	os.Clearenv()
}

func TestConfig_GetRedisAddress(t *testing.T) {
	t.Run("returns default redis address", func(t *testing.T) {
		os.Clearenv()
		resetConfig()
		cfg := Load()
		assert.Equal(t, "localhost:6379", cfg.GetRedisAddress())
	})

	t.Run("returns custom redis address", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("REDIS_HOST", "redis.example.com")
		os.Setenv("REDIS_PORT", "6380")
		resetConfig()
		cfg := Load()
		assert.Equal(t, "redis.example.com:6380", cfg.GetRedisAddress())
	})

	// Clean up
	os.Clearenv()
}

func TestGetEnvWithDefault(t *testing.T) {
	t.Run("returns environment variable when set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		result := getEnvWithDefault("TEST_VAR", "default_value")
		assert.Equal(t, "test_value", result)
		os.Unsetenv("TEST_VAR")
	})

	t.Run("returns default when environment variable not set", func(t *testing.T) {
		os.Unsetenv("TEST_VAR")
		result := getEnvWithDefault("TEST_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})

	t.Run("returns default when environment variable is empty", func(t *testing.T) {
		os.Setenv("TEST_VAR", "")
		result := getEnvWithDefault("TEST_VAR", "default_value")
		assert.Equal(t, "default_value", result)
		os.Unsetenv("TEST_VAR")
	})
}

// Helper function to clean up after tests
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Clean up
	os.Clearenv()
	resetConfig()

	os.Exit(code)
}
