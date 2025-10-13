# Phase 10.1 - JWT Token Compatibility Analysis
## Hub User Service Migration - JWT Token Specification

**Date**: 2025-10-13  
**Status**: Step 1.4 - COMPLETED ✅  
**Deliverable**: Complete JWT token specification with compatibility strategy

---

## 📋 Table of Contents
1. [Executive Summary](#executive-summary)
2. [JWT Token Specification](#jwt-token-specification)
3. [Token Structure Analysis](#token-structure-analysis)
4. [Secret Management](#secret-management)
5. [Token Lifecycle](#token-lifecycle)
6. [Compatibility Requirements](#compatibility-requirements)
7. [Migration Strategy](#migration-strategy)
8. [Security Considerations](#security-considerations)
9. [Testing Strategy](#testing-strategy)

---

## Executive Summary

### Key Findings
- ✅ **Algorithm**: HS256 (HMAC with SHA-256)
- ✅ **Library**: `github.com/golang-jwt/jwt v3.2.2+incompatible`
- ✅ **Claims**: `username`, `userId`, `exp` (3 claims total)
- ✅ **Expiration**: 10 minutes (600 seconds)
- ✅ **Secret Source**: Environment variable `MY_JWT_SECRET`
- ✅ **Token Format**: `Bearer <token>` in Authorization header

### Compatibility Decision
**✅ COMPATIBLE**: Microservice MUST use identical JWT configuration

**Critical Requirements**:
1. ✅ Same signing algorithm (HS256)
2. ✅ Same secret key (shared via environment variable)
3. ✅ Same claims structure (`username`, `userId`, `exp`)
4. ✅ Same expiration time (10 minutes)
5. ✅ Same JWT library (or compatible)

---

## JWT Token Specification

### Token Format

#### **Header**
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

**Details**:
- **Algorithm**: HS256 (HMAC-SHA256)
- **Type**: JWT (JSON Web Token)
- **No additional headers**: Standard JWT header only

---

#### **Payload (Claims)**
```json
{
  "username": "user@example.com",
  "userId": "12345",
  "exp": 1728849600
}
```

**Claims Specification**:

| Claim | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `username` | string | ✅ Yes | User's email address | "user@example.com" |
| `userId` | string | ✅ Yes | User's unique ID (from database) | "12345" |
| `exp` | int64 | ✅ Yes | Expiration timestamp (Unix time) | 1728849600 |

**Notes**:
- ✅ **No `iat` (issued at)**: Not included in current implementation
- ✅ **No `nbf` (not before)**: Not included
- ✅ **No `iss` (issuer)**: Not included
- ✅ **No `aud` (audience)**: Not included
- ✅ **No `jti` (JWT ID)**: Not included
- ⚠️ **Username is email**: Field name is "username" but value is email address

---

#### **Signature**
```
HMACSHA256(
  base64UrlEncode(header) + "." +
  base64UrlEncode(payload),
  secret
)
```

**Signature Details**:
- **Algorithm**: HMAC-SHA256
- **Secret**: Loaded from `config.JWTSecret`
- **Source**: Environment variable `MY_JWT_SECRET`

---

### Complete Token Example

**Token String**:
```
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mjg4NDk2MDAsInVzZXJJZCI6IjEyMzQ1IiwidXNlcm5hbWUiOiJ1c2VyQGV4YW1wbGUuY29tIn0.signature_here
```

**Structure**:
- `Bearer` prefix (space-separated)
- Header (base64url encoded)
- `.` separator
- Payload (base64url encoded)
- `.` separator
- Signature (base64url encoded)

---

## Token Structure Analysis

### Token Creation (token_service.go)

**Source Code** (Lines 26-42):
```go
func (s *TokenService) CreateAndSignToken(userName string, userId string) (string, error) {
    cfg := config.Get()

    token := jwt.NewWithClaims(jwt.SigningMethodHS256,
        jwt.MapClaims{
            "username": userName,
            "userId":   userId,
            "exp":      time.Now().Add(time.Minute * 10).Unix(), //token expiration time = 10 min
        })

    tokenString, err := token.SignedString([]byte(cfg.JWTSecret))

    if err != nil {
        return "", err
    }

    return tokenString, nil
}
```

**Analysis**:
1. ✅ Uses `jwt.SigningMethodHS256` (HMAC-SHA256)
2. ✅ Creates `jwt.MapClaims` with 3 fields
3. ✅ Expiration: `time.Now().Add(time.Minute * 10).Unix()`
4. ✅ Signs with `cfg.JWTSecret` as byte array
5. ✅ Returns raw token string (NO "Bearer" prefix added here)

---

### Token Validation (token_service.go)

**Source Code** (Lines 45-86):
```go
func (s *TokenService) ValidateToken(tokenString string) (map[string]interface{}, error) {
    token, err := s.parseToken(tokenString)

    if err != nil {
        return nil, err
    }

    claims, err := validateToken(token)

    if err != nil {
        return nil, err
    }

    bla := TokenClaims(claims)

    return bla, nil
}

func (s *TokenService) parseToken(token string) (*jwt.Token, error) {
    token = token[len("Bearer "):]  // ⚠️ STRIPS "Bearer " PREFIX
    cfg := config.Get()

    jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
        return []byte(cfg.JWTSecret), nil
    })

    return jwtToken, err
}

func validateToken(token *jwt.Token) (jwt.MapClaims, error) {
    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)

    if !ok {
        return nil, errors.New("invalid claims")
    }

    return claims, nil
}
```

**Analysis**:
1. ⚠️ **Assumes "Bearer " prefix**: Line 64 strips it unconditionally
2. ✅ Uses same `cfg.JWTSecret` for validation
3. ✅ Validates token signature automatically via `jwt.Parse()`
4. ✅ Checks `token.Valid` (includes expiration check)
5. ✅ Returns claims as `map[string]interface{}`

**Critical Issue**:
```go
token = token[len("Bearer "):]  // Line 64
```
**This will panic if token doesn't start with "Bearer "**

---

### Token Usage in AuthService (auth_service.go)

**Source Code** (Lines 23-40):
```go
func (s *AuthService) VerifyToken(tokenString string, w http.ResponseWriter) (string, error) {
    if tokenString == "" {
        w.WriteHeader(http.StatusUnauthorized)
        w.Write([]byte("Token não fornecido"))
        return "", errors.New("token not provided")
    }

    tokenInfo, err := s.tokenService.ValidateToken(tokenString)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        w.Write([]byte("Token inválido"))
        return "", err
    }

    userIdInterface, exists := tokenInfo["userId"]
    if !exists {
        w.WriteHeader(http.StatusUnauthorized)
        w.Write([]byte("userId not found in token"))
        return "", errors.New("userId not found in token")
    }

    userId, ok := userIdInterface.(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        w.Write([]byte("userId is not a string"))
        return "", errors.New("userId is not a string")
    }

    return userId, nil
}
```

**Analysis**:
1. ✅ Validates token is not empty
2. ✅ Calls `tokenService.ValidateToken()`
3. ✅ Extracts `userId` from claims
4. ✅ Type checks `userId` (must be string)
5. ✅ Writes HTTP errors directly to `ResponseWriter`
6. ✅ Returns only `userId` (not full claims)

**Important**: Only `userId` is used by the application, `username` is ignored after validation

---

## Secret Management

### Current Implementation

**Configuration Source** (`shared/config/config.go`):
```go
type Config struct {
    HTTPPort    string
    GRPCPort    string
    JWTSecret   string  // ← JWT secret stored here
    RedisHost   string
    RedisPort   string
    DatabaseURL string
}

func Load() *Config {
    once.Do(func() {
        // Try to load from config.env file
        err := godotenv.Load("config.env")
        if err != nil {
            log.Printf("Warning: Could not load config.env file: %v", err)
            log.Println("Using environment variables or default values...")
        }

        instance = &Config{
            HTTPPort:    getEnvWithDefault("HTTP_PORT", "localhost:8080"),
            GRPCPort:    getEnvWithDefault("GRPC_PORT", "localhost:50051"),
            JWTSecret:   getEnvWithDefault("MY_JWT_SECRET", "default-secret-key-change-in-production"),
            // ... other fields
        }

        // Validate required configuration
        if instance.JWTSecret == "default-secret-key-change-in-production" {
            log.Println("Warning: Using default JWT secret. Please set MY_JWT_SECRET environment variable for production.")
        }
    })

    return instance
}
```

**Secret Loading Priority**:
1. **Environment variable**: `MY_JWT_SECRET`
2. **Config file**: `config.env` (if exists)
3. **Default value**: `"default-secret-key-change-in-production"` (with warning)

---

### Current Secret Value

**From `config.env`**:
```bash
MY_JWT_SECRET=HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^
```

**Analysis**:
- ✅ **Strong secret**: 46 characters, mixed case, numbers, special characters
- ✅ **Not default**: Custom secret properly configured
- ✅ **Shared secret**: Same secret used for signing and verification

---

### Secret Access Pattern

**In Token Service**:
```go
cfg := config.Get()
token.SignedString([]byte(cfg.JWTSecret))
```

**In Token Parsing**:
```go
cfg := config.Get()
jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
    return []byte(cfg.JWTSecret), nil
})
```

**Pattern**: 
- ✅ Always accessed via `config.Get().JWTSecret`
- ✅ Converted to `[]byte` for JWT library
- ✅ Singleton pattern ensures same secret throughout app

---

## Token Lifecycle

### 1. Token Creation Flow

```
User Login Request
    │
    ▼
do_login.go: Execute login usecase
    │
    ▼
Validate credentials (username + password)
    │
    ▼
container.GetAuthService().CreateToken(email, userId)
    │
    ▼
auth_service.go: CreateToken()
    │
    ▼
token_service.go: CreateAndSignToken()
    │
    ├─> Load config.JWTSecret
    ├─> Create JWT claims (username, userId, exp)
    ├─> Sign with HS256 + secret
    └─> Return token string
    │
    ▼
Return to client: { "token": "eyJhbGci..." }
```

**Duration**: Token created with `exp = now + 10 minutes`

---

### 2. Token Validation Flow

```
Protected Endpoint Request
    │
    ▼
middleware.WithAuthentication()
    │
    ├─> Extract Authorization header
    └─> Call verifyToken(tokenString, w)
        │
        ▼
auth_service.go: VerifyToken()
        │
        ├─> Check if token is empty
        └─> Call tokenService.ValidateToken()
            │
            ▼
token_service.go: ValidateToken()
            │
            ├─> parseToken() - strips "Bearer " prefix
            ├─> Load config.JWTSecret
            ├─> jwt.Parse() - validates signature
            ├─> Check token.Valid (includes expiration)
            └─> Return claims
            │
            ▼
auth_service.go: Extract userId from claims
            │
            ▼
Return userId to handler
            │
            ▼
Business logic executes with userId
```

---

### 3. Token Expiration

**Expiration Time**: 10 minutes (600 seconds)

**Code** (token_service.go:33):
```go
"exp": time.Now().Add(time.Minute * 10).Unix()
```

**Expiration Handling**:
- ✅ JWT library automatically checks `exp` claim
- ✅ `token.Valid` returns false if expired
- ✅ Returns error: "invalid token"
- ✅ HTTP response: 401 Unauthorized

**Token Lifetime Example**:
```
Created:  2025-10-13 10:00:00 UTC
Expires:  2025-10-13 10:10:00 UTC
Duration: 600 seconds (10 minutes)
```

**No Refresh Token**: Current implementation does not support token refresh

---

## Compatibility Requirements

### Microservice MUST Match Monolith

#### ✅ **1. Algorithm Compatibility**
```go
// Monolith
jwt.SigningMethodHS256

// Microservice MUST use
jwt.SigningMethodHS256  // ← EXACT SAME
```

**Why Critical**: Different algorithms = signature validation fails

---

#### ✅ **2. Secret Compatibility**
```go
// Monolith
cfg.JWTSecret = "HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^"

// Microservice MUST use
cfg.JWTSecret = "HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^"  // ← EXACT SAME
```

**Why Critical**: Different secrets = all tokens invalid

**Strategy**: Shared environment variable `MY_JWT_SECRET`

---

#### ✅ **3. Claims Compatibility**
```go
// Monolith
jwt.MapClaims{
    "username": userName,  // User's email
    "userId":   userId,    // User's ID (string)
    "exp":      exp,       // Expiration (Unix timestamp)
}

// Microservice MUST use
jwt.MapClaims{
    "username": userName,  // ← EXACT SAME FIELD NAME
    "userId":   userId,    // ← EXACT SAME FIELD NAME
    "exp":      exp,       // ← EXACT SAME FIELD NAME
}
```

**Why Critical**: Monolith expects `userId` claim by name

---

#### ✅ **4. Expiration Compatibility**
```go
// Monolith
time.Now().Add(time.Minute * 10).Unix()

// Microservice MUST use
time.Now().Add(time.Minute * 10).Unix()  // ← EXACT SAME (10 minutes)
```

**Why Critical**: Consistency in user experience and security

---

#### ✅ **5. Token Format Compatibility**
```go
// Monolith expects
"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

// Microservice MUST return
"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  // ← WITH "Bearer " PREFIX
```

**Why Critical**: Token parsing strips "Bearer " prefix unconditionally

---

#### ✅ **6. Library Compatibility**
```go
// Monolith
github.com/golang-jwt/jwt v3.2.2+incompatible

// Microservice MUST use
github.com/golang-jwt/jwt v3.2.2+incompatible  // ← SAME VERSION
// OR
github.com/golang-jwt/jwt/v4  // ← Compatible upgrade (v4 is compatible)
```

**Why Critical**: Ensures same JWT parsing behavior

---

### Compatibility Verification Checklist

- [ ] **Algorithm**: HS256 (HMAC-SHA256)
- [ ] **Secret**: Loaded from `MY_JWT_SECRET` environment variable
- [ ] **Secret Value**: Exact match with monolith
- [ ] **Claims**: `username`, `userId`, `exp` (3 claims)
- [ ] **Expiration**: 10 minutes (600 seconds)
- [ ] **Token Format**: "Bearer " prefix in Authorization header
- [ ] **Library**: `github.com/golang-jwt/jwt` v3 or v4
- [ ] **Claim Types**: username=string, userId=string, exp=int64

---

## Migration Strategy

### Phase 1: Shared Secret (Current → Migration)

**Strategy**: Both monolith and microservice use same JWT secret

```
┌─────────────────────────────────────────────────────────────┐
│                    Environment Variable                      │
│              MY_JWT_SECRET=HubInv3stm3nts...                │
└─────────────────────────────────────────────────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │                            │
                ▼                            ▼
┌─────────────────────────┐   ┌─────────────────────────┐
│      MONOLITH           │   │   USER MICROSERVICE     │
│   config.Get().JWT      │   │  config.Get().JWT       │
│   Secret                │   │  Secret                 │
└─────────────────────────┘   └─────────────────────────┘
                │                            │
                ├─> Signs tokens ◄───────────┤
                └─> Verifies tokens ◄────────┘
```

**Benefits**:
- ✅ Simple implementation
- ✅ No token format changes
- ✅ Tokens interchangeable
- ✅ Low risk

**Requirements**:
- ✅ Same environment variable name
- ✅ Same secret value
- ✅ Both services have access

---

### Phase 2: Token Interoperability

**Scenario 1: Token Created by Monolith, Validated by Microservice**

```
Client → Monolith /login → Token (signed with shared secret)
                        ↓
Client → Monolith /getBalance → gRPC → Microservice ValidateToken
                                              ↓
                                    Validates with shared secret ✅
```

**Result**: ✅ **WORKS** (shared secret)

---

**Scenario 2: Token Created by Microservice, Validated by Monolith**

```
Client → Microservice Login (gRPC) → Token (signed with shared secret)
                                   ↓
Client → Monolith /getBalance → verifyToken (validates with shared secret) ✅
```

**Result**: ✅ **WORKS** (shared secret)

---

### Implementation in Microservice

**Required Code** (Go):
```go
package token

import (
    "time"
    "github.com/golang-jwt/jwt"
)

type TokenService struct {
    jwtSecret string
}

func NewTokenService(jwtSecret string) *TokenService {
    return &TokenService{jwtSecret: jwtSecret}
}

// CreateToken generates a JWT token
func (s *TokenService) CreateToken(username string, userID string) (string, error) {
    claims := jwt.MapClaims{
        "username": username,
        "userId":   userID,
        "exp":      time.Now().Add(time.Minute * 10).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.jwtSecret))
}

// ValidateToken verifies a JWT token
func (s *TokenService) ValidateToken(tokenString string) (string, error) {
    // Strip "Bearer " prefix if present
    if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
        tokenString = tokenString[7:]
    }

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(s.jwtSecret), nil
    })

    if err != nil {
        return "", err
    }

    if !token.Valid {
        return "", errors.New("invalid token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return "", errors.New("invalid claims")
    }

    userID, ok := claims["userId"].(string)
    if !ok {
        return "", errors.New("userId not found or invalid type")
    }

    return userID, nil
}
```

**Configuration**:
```go
// In main.go or config
jwtSecret := os.Getenv("MY_JWT_SECRET")
tokenService := token.NewTokenService(jwtSecret)
```

---

## Security Considerations

### Current Security Posture

#### ✅ **Strengths**

1. **Strong Secret**
   - 46 characters
   - Mixed case, numbers, special characters
   - Not using default value

2. **Short Expiration**
   - 10 minutes reduces window of attack
   - Forces re-authentication

3. **Industry Standard Algorithm**
   - HS256 (HMAC-SHA256)
   - Widely supported
   - Battle-tested

4. **Token Validation**
   - Signature verification
   - Expiration check
   - Claims validation

---

#### ⚠️ **Weaknesses and Recommendations**

1. **⚠️ No Token Revocation**
   - **Issue**: Tokens valid until expiration, even if user logs out
   - **Impact**: Medium
   - **Recommendation**: Implement token blacklist (Redis)

2. **⚠️ Short Expiration = Poor UX**
   - **Issue**: Users must re-login every 10 minutes
   - **Impact**: High
   - **Recommendation**: Implement refresh tokens (30 days) + access tokens (10 min)

3. **⚠️ Token Lifetime Not Configurable**
   - **Issue**: Hardcoded 10 minutes
   - **Impact**: Low
   - **Recommendation**: Move to environment variable `JWT_EXPIRATION_MINUTES`

4. **⚠️ No Token Rotation**
   - **Issue**: Same token used for entire session
   - **Impact**: Low
   - **Recommendation**: Implement token rotation on refresh

5. **⚠️ Bearer Prefix Handling Unsafe**
   - **Issue**: `token[len("Bearer "):]` panics if missing prefix
   - **Impact**: Medium
   - **Recommendation**: Check prefix exists before stripping

6. **⚠️ No Additional Claims**
   - **Issue**: Missing `iat` (issued at), `jti` (JWT ID)
   - **Impact**: Low
   - **Recommendation**: Add for audit trail

---

### Security Best Practices for Microservice

#### ✅ **MUST Implement**

1. **Same Secret Management**
   - Use same `MY_JWT_SECRET` environment variable
   - Never log or expose secret
   - Rotate secret periodically (with grace period)

2. **Validation Before Stripping Prefix**
   ```go
   // GOOD
   if strings.HasPrefix(tokenString, "Bearer ") {
       tokenString = tokenString[7:]
   }
   
   // BAD (current implementation)
   tokenString = tokenString[len("Bearer "):]  // ← Can panic
   ```

3. **Error Logging (Without Token)**
   ```go
   // GOOD
   log.Printf("Token validation failed: %v", err)
   
   // BAD
   log.Printf("Token validation failed for token %s: %v", token, err)  // ← Logs secret
   ```

4. **HTTPS Only**
   - Never send tokens over HTTP
   - Enforce TLS/SSL in production

---

#### 🔄 **Future Enhancements**

1. **Token Blacklist** (Redis)
   - Store revoked tokens
   - Check on validation
   - Auto-expire with token expiration

2. **Refresh Tokens**
   - Long-lived refresh token (30 days)
   - Short-lived access token (10 minutes)
   - Reduce login frequency

3. **Token Rotation**
   - Issue new token on refresh
   - Invalidate old token

4. **Rate Limiting**
   - Limit token validation requests
   - Prevent brute force attacks

---

## Testing Strategy

### Token Compatibility Tests

#### Test 1: Cross-Service Token Creation & Validation

```go
func TestTokenCompatibility_MonolithCreatesServiceValidates(t *testing.T) {
    // Arrange
    secret := "HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^"
    
    // Monolith creates token
    monolithService := NewMonolithTokenService(secret)
    token, err := monolithService.CreateToken("user@example.com", "123")
    assert.NoError(t, err)
    
    // Microservice validates token
    microserviceService := NewMicroserviceTokenService(secret)
    userID, err := microserviceService.ValidateToken(token)
    assert.NoError(t, err)
    assert.Equal(t, "123", userID)
}

func TestTokenCompatibility_ServiceCreatesMonolithValidates(t *testing.T) {
    // Arrange
    secret := "HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^"
    
    // Microservice creates token
    microserviceService := NewMicroserviceTokenService(secret)
    token, err := microserviceService.CreateToken("user@example.com", "123")
    assert.NoError(t, err)
    
    // Monolith validates token
    monolithService := NewMonolithTokenService(secret)
    userID, err := monolithService.ValidateToken("Bearer " + token)
    assert.NoError(t, err)
    assert.Equal(t, "123", userID)
}
```

---

#### Test 2: Claims Compatibility

```go
func TestTokenCompatibility_ClaimsStructure(t *testing.T) {
    secret := "HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^"
    
    service := NewTokenService(secret)
    token, err := service.CreateToken("user@example.com", "123")
    assert.NoError(t, err)
    
    // Parse token manually
    claims := parseTokenClaims(token)
    
    // Verify claims
    assert.Equal(t, "user@example.com", claims["username"])
    assert.Equal(t, "123", claims["userId"])
    assert.NotNil(t, claims["exp"])
    
    // Verify only expected claims
    assert.Len(t, claims, 3)
}
```

---

#### Test 3: Expiration Compatibility

```go
func TestTokenCompatibility_ExpirationTime(t *testing.T) {
    secret := "HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^"
    
    before := time.Now()
    service := NewTokenService(secret)
    token, err := service.CreateToken("user@example.com", "123")
    after := time.Now()
    
    assert.NoError(t, err)
    
    claims := parseTokenClaims(token)
    exp := int64(claims["exp"].(float64))
    
    // Verify expiration is 10 minutes from now
    expectedExp := before.Add(time.Minute * 10).Unix()
    tolerance := int64(5) // 5 seconds tolerance
    
    assert.InDelta(t, expectedExp, exp, float64(tolerance))
}
```

---

#### Test 4: Secret Mismatch Detection

```go
func TestTokenCompatibility_SecretMismatch(t *testing.T) {
    secret1 := "secret1"
    secret2 := "secret2"
    
    service1 := NewTokenService(secret1)
    token, err := service1.CreateToken("user@example.com", "123")
    assert.NoError(t, err)
    
    service2 := NewTokenService(secret2)
    _, err = service2.ValidateToken("Bearer " + token)
    
    // Should fail due to different secrets
    assert.Error(t, err)
}
```

---

### Integration Test: End-to-End Flow

```go
func TestTokenCompatibility_EndToEndFlow(t *testing.T) {
    // 1. User logs in via microservice
    loginResp, err := microserviceClient.Login(ctx, &pb.LoginRequest{
        Email:    "user@example.com",
        Password: "password123",
    })
    assert.NoError(t, err)
    token := loginResp.Token
    
    // 2. Use token to access monolith endpoint
    req := httptest.NewRequest("GET", "/getBalance", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    w := httptest.NewRecorder()
    monolithHandler.ServeHTTP(w, req)
    
    // 3. Verify response
    assert.Equal(t, http.StatusOK, w.Code)
    
    // 4. Use same token to access another monolith endpoint
    req2 := httptest.NewRequest("GET", "/getPortfolioSummary", nil)
    req2.Header.Set("Authorization", "Bearer "+token)
    
    w2 := httptest.NewRecorder()
    monolithHandler.ServeHTTP(w2, req2)
    
    assert.Equal(t, http.StatusOK, w2.Code)
}
```

---

## Summary

### JWT Token Specification

| Property | Value | Notes |
|----------|-------|-------|
| **Algorithm** | HS256 | HMAC-SHA256 |
| **Library** | `github.com/golang-jwt/jwt v3.2.2` | Monolith version |
| **Secret Source** | `MY_JWT_SECRET` env var | Shared secret |
| **Secret Value** | `HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^` | From config.env |
| **Claims** | `username`, `userId`, `exp` | 3 claims total |
| **Expiration** | 10 minutes | 600 seconds |
| **Token Format** | `Bearer <token>` | In Authorization header |

---

### Compatibility Requirements

✅ **MUST Match**:
1. Algorithm: HS256
2. Secret: Same value from `MY_JWT_SECRET`
3. Claims: `username`, `userId`, `exp`
4. Expiration: 10 minutes
5. Token format: "Bearer " prefix

---

### Migration Checklist

- [ ] **Microservice uses same JWT library** (v3 or compatible v4)
- [ ] **Microservice loads secret from `MY_JWT_SECRET`**
- [ ] **Microservice creates tokens with identical claims**
- [ ] **Microservice uses 10-minute expiration**
- [ ] **Microservice handles "Bearer " prefix correctly**
- [ ] **Integration tests verify cross-service compatibility**
- [ ] **Security: Never log JWT secret or tokens**
- [ ] **Security: Use HTTPS only**

---

### Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Secret mismatch | ⚠️ High | Environment variable validation |
| Claims structure mismatch | ⚠️ High | Integration tests |
| Different expiration time | ⚠️ Medium | Configuration constant |
| Library incompatibility | ⚠️ Low | Use same library version |
| Token format issues | ⚠️ Low | Prefix handling tests |

---

### Next Steps

✅ **Step 1.4**: JWT Token Compatibility Analysis - **COMPLETED**  
⏭️ **Step 1.5**: Test Inventory  
⏭️ **Week 2**: Microservice project setup  
⏭️ **Week 3**: Copy code AS-IS to microservice

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-13  
**Author**: AI Assistant  
**Status**: ✅ Ready for Step 1.5

