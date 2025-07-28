# Centralized Configuration Management

This package provides a centralized, thread-safe configuration management system for the HubInvestments application. It eliminates code duplication when accessing environment variables and provides a single source of truth for all configuration.

## Problem Solved

Previously, environment variable loading was scattered throughout the codebase:

```go
// OLD APPROACH - Duplicated in multiple files
func someFunction() {
    err := godotenv.Load("config.env")
    if err != nil {
        log.Printf("Warning: Could not load config.env file: %v", err)
    }
    
    jwtSecret := os.Getenv("MY_JWT_SECRET")
    httpPort := os.Getenv("HTTP_PORT")
    // ... more duplication
}
```

**Issues with the old approach:**
- 📝 **Code duplication** - Same loading logic in multiple files
- 🔄 **Inconsistent error handling** - Different files handled missing config differently  
- ⚡ **Performance** - Config file loaded multiple times
- 🧪 **Testing complexity** - Hard to mock configuration in tests
- 🔧 **Maintenance** - Adding new config required changes in multiple places

## Solution

The centralized config system provides:

1. **Single Point of Configuration** - All config loading happens in one place
2. **Thread-Safe Singleton** - Configuration loaded once and shared safely
3. **Consistent Defaults** - Reliable fallback values for all environments
4. **Easy Testing** - Simple to mock and test configuration
5. **Type Safety** - Strongly typed configuration fields

## Architecture

```
┌─────────────────────┐
│   Application       │
│                     │
│  main.go           │◄──┐
│  token_service.go  │   │
│  other_services.go │   │
└─────────────────────┘   │
                          │
        ┌─────────────────▼───────────────────┐
        │     shared/config Package          │
        │                                    │
        │  ┌──────────────┐  ┌─────────────┐ │
        │  │ config.env   │  │ Environment │ │
        │  │ (file)       │  │ Variables   │ │
        │  └──────────────┘  └─────────────┘ │
        │           │              │        │
        │           └──────┬───────┘        │
        │                  │                │
        │      ┌───────────▼──────────┐     │
        │      │ Centralized Config   │     │
        │      │ (singleton)          │     │
        │      └──────────────────────┘     │
        └────────────────────────────────────┘
```

## Usage

### Basic Usage

```go
package mypackage

import "HubInvestments/shared/config"

func MyFunction() {
    // Get configuration (loads automatically on first call)
    cfg := config.Get()
    
    // Use configuration values
    httpPort := cfg.HTTPPort
    jwtSecret := cfg.JWTSecret
    redisAddr := cfg.GetRedisAddress()
    
    // Check environment
    if cfg.IsProduction() {
        // Production-specific logic
    }
}
```

### Available Configuration Fields

```go
type Config struct {
    HTTPPort    string  // HTTP server port (default: "localhost:8080")
    GRPCPort    string  // gRPC server port (default: "localhost:50051") 
    JWTSecret   string  // JWT signing secret (default: warns about insecure default)
    RedisHost   string  // Redis server host (default: "localhost")
    RedisPort   string  // Redis server port (default: "6379")
    DatabaseURL string  // Database connection string (default: "")
}
```

### Helper Methods

```go
cfg := config.Get()

// Get complete Redis address
redisAddr := cfg.GetRedisAddress() // Returns "localhost:6379"

// Check if running in production
if cfg.IsProduction() {
    // Enable production features
}
```

### Loading Configuration

Configuration is automatically loaded from these sources in order:

1. **config.env file** (if exists)
2. **Environment variables** 
3. **Default values** (fallback)

```go
// Load explicitly (optional - happens automatically)
cfg := config.Load()

// Get existing instance (loads if not already loaded)
cfg := config.Get()
```

## Configuration Files

### config.env (Active Configuration)
```bash
# Your active configuration
HTTP_PORT=192.168.0.3:8080
GRPC_PORT=192.168.0.6:50051
MY_JWT_SECRET=your-secret-key
REDIS_HOST=localhost
REDIS_PORT=6379
```

### config.example.env (Template)
```bash
# Copy this to config.env and modify
HTTP_PORT=192.168.0.3:8080
MY_JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
# ... more examples
```

## Environment Variables

All configuration can be overridden with environment variables:

```bash
# Override specific values
export HTTP_PORT="0.0.0.0:8080"
export MY_JWT_SECRET="production-secret"
export ENVIRONMENT="production"

# Run application
go run main.go
```

## Testing

The configuration system is designed to be test-friendly:

```go
func TestMyFunction(t *testing.T) {
    // Set test environment variables
    os.Setenv("HTTP_PORT", "localhost:9999")
    os.Setenv("MY_JWT_SECRET", "test-secret")
    
    // Reset config for testing (if needed)
    // Note: Only do this in tests
    
    cfg := config.Get()
    assert.Equal(t, "localhost:9999", cfg.HTTPPort)
}
```

## Examples

### 1. HTTP Server Setup
```go
func startServer() {
    cfg := config.Get()
    
    log.Printf("Starting HTTP server on %s", cfg.HTTPPort)
    err := http.ListenAndServe(cfg.HTTPPort, nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 2. JWT Token Service
```go
func (s *TokenService) CreateToken(username, userID string) (string, error) {
    cfg := config.Get()
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": username,
        "userId":   userID,
        "exp":      time.Now().Add(time.Hour).Unix(),
    })
    
    return token.SignedString([]byte(cfg.JWTSecret))
}
```

### 3. Redis Connection
```go
func connectToRedis() *redis.Client {
    cfg := config.Get()
    
    return redis.NewClient(&redis.Options{
        Addr: cfg.GetRedisAddress(), // localhost:6379
    })
}
```

### 4. Environment-Specific Logic
```go
func setupLogging() {
    cfg := config.Get()
    
    if cfg.IsProduction() {
        // JSON structured logging for production
        log.SetFormatter(&logrus.JSONFormatter{})
    } else {
        // Human-readable logs for development
        log.SetFormatter(&logrus.TextFormatter{})
    }
}
```

## Best Practices

### ✅ Do
- Use `config.Get()` to access configuration
- Set environment-specific values in `config.env`
- Use the `IsProduction()` helper for environment checks
- Provide sensible defaults for all configuration

### ❌ Don't
- Load `godotenv` directly in your services
- Access `os.Getenv()` directly for application config
- Hardcode configuration values
- Forget to add new config fields to the Config struct

## Migration Guide

### From Old Environment Loading
```go
// OLD ❌
func oldFunction() {
    err := godotenv.Load("config.env")
    if err != nil {
        log.Printf("Warning: Could not load config.env file: %v", err)
    }
    
    jwtSecret := os.Getenv("MY_JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "default-secret"
    }
}

// NEW ✅
func newFunction() {
    cfg := config.Get()
    jwtSecret := cfg.JWTSecret
}
```

### Adding New Configuration
1. Add field to `Config` struct in `config.go`
2. Add to `Load()` function with `getEnvWithDefault()`
3. Add to `config.example.env` with documentation
4. Update tests if needed

```go
// 1. Add to Config struct
type Config struct {
    // ... existing fields
    NewSetting string
}

// 2. Add to Load() function
instance = &Config{
    // ... existing fields
    NewSetting: getEnvWithDefault("NEW_SETTING", "default-value"),
}
```

## Thread Safety

The configuration system is thread-safe using `sync.Once`:

- ✅ **Safe** to call `config.Get()` from multiple goroutines
- ✅ **Guaranteed** to load configuration only once
- ✅ **No race conditions** when accessing configuration

## Summary

You now have a centralized configuration system that:
- ✅ **Eliminates code duplication** across environment variable loading
- ✅ **Provides thread-safe access** to configuration
- ✅ **Supports multiple sources** (files, env vars, defaults)
- ✅ **Is easy to test** and mock
- ✅ **Follows singleton pattern** for efficiency
- ✅ **Includes helpful utilities** for common operations

This makes your codebase more maintainable, testable, and consistent! 🎉 