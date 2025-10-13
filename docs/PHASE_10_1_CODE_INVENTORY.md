# Phase 10.1 - Code Inventory and Dependency Map
## Hub User Service Migration - Deep Code Analysis

**Date**: 2025-10-13  
**Status**: Step 1.1 - COMPLETED ✅  
**Deliverable**: Complete code inventory document with dependency map

---

## 📋 Table of Contents
1. [Module Overview](#module-overview)
2. [Authentication Module (`internal/auth/`)](#authentication-module)
3. [Login Module (`internal/login/`)](#login-module)
4. [External Dependencies](#external-dependencies)
5. [Integration Points](#integration-points)
6. [JWT Token Structure](#jwt-token-structure)
7. [Migration Checklist](#migration-checklist)

---

## Module Overview

The User Service encompasses two main modules:
- **Authentication Module** (`internal/auth/`): JWT token creation and validation
- **Login Module** (`internal/login/`): User authentication and domain logic

**Total Lines of Code**: ~800 lines  
**Test Coverage**: 8 test files with comprehensive coverage  
**External Dependencies**: 3 (jwt, config, database)

---

## Authentication Module (`internal/auth/`)

### 📁 File Structure
```
internal/auth/
├── auth_service.go (45 lines)
├── auth_service_test.go (273 lines)
└── token/
    ├── token_service.go (87 lines)
    └── token_service_test.go
```

### 🔍 `auth_service.go` Analysis

**Package**: `auth`  
**Lines**: 45  
**Purpose**: JWT token verification and creation facade

#### Interfaces
```go
type IAuthService interface {
    VerifyToken(tokenString string, w http.ResponseWriter) (string, error)
    CreateToken(userName string, userId string) (string, error)
}
```

#### Implementation
```go
type AuthService struct {
    tokenService token.ITokenService
}
```

#### Dependencies
- ✅ `internal/auth/token` - Token service for JWT operations
- ✅ `net/http` - HTTP response writer (standard library)
- ✅ `errors` - Error handling (standard library)
- ✅ `fmt` - String formatting (standard library)

#### Methods
1. **`NewAuthService(tokenService token.ITokenService) IAuthService`**
   - Constructor function
   - Dependency injection pattern
   - No business logic

2. **`VerifyToken(tokenString string, w http.ResponseWriter) (string, error)`**
   - Validates JWT token
   - Returns `userId` from token claims
   - Writes HTTP 401 if token is empty
   - **Issue**: Tightly coupled to `http.ResponseWriter`

3. **`CreateToken(userName string, userId string) (string, error)`**
   - Delegates to `tokenService.CreateAndSignToken()`
   - Simple pass-through method

#### Key Observations
- ✅ Clean interface-based design
- ✅ Dependency injection
- ⚠️ HTTP coupling in domain service (acceptable for migration)
- ✅ No database dependencies
- ✅ Simple, straightforward logic

---

### 🔍 `token/token_service.go` Analysis

**Package**: `token`  
**Lines**: 87  
**Purpose**: JWT token creation, signing, and validation

#### Interfaces
```go
type ITokenService interface {
    CreateAndSignToken(userName string, userId string) (string, error)
    ValidateToken(tokenString string) (map[string]interface{}, error)
}
```

#### Implementation
```go
type TokenService struct{}
```

#### Dependencies
- ✅ `HubInvestments/shared/config` - JWT secret configuration
- ✅ `github.com/golang-jwt/jwt` - JWT library (v3.2.2+)
- ✅ `time` - Token expiration (standard library)
- ✅ `errors` - Error handling (standard library)

#### JWT Token Structure
```go
jwt.MapClaims{
    "username": userName,  // User's email
    "userId":   userId,    // User's ID (string)
    "exp":      time.Now().Add(time.Minute * 10).Unix(), // 10 min expiration
}
```

#### Methods
1. **`CreateAndSignToken(userName string, userId string) (string, error)`**
   - Signing Algorithm: **HS256** (HMAC with SHA-256)
   - Token Expiration: **10 minutes**
   - Secret Source: `config.Get().JWTSecret`
   - Returns: Signed JWT token string

2. **`ValidateToken(tokenString string) (map[string]interface{}, error)`**
   - Parses JWT token
   - Validates signature
   - Returns claims map

3. **`parseToken(token string) (*jwt.Token, error)` (private)**
   - Strips "Bearer " prefix
   - Parses with secret key
   - Returns parsed token

4. **`validateToken(token *jwt.Token) (jwt.MapClaims, error)` (private)**
   - Checks token validity
   - Extracts claims
   - Returns error if invalid

#### Key Observations
- ✅ No stateless service (no struct fields)
- ✅ Clean separation of concerns
- ⚠️ **CRITICAL**: Hardcoded "Bearer " prefix stripping (line 64)
- ✅ JWT secret loaded from config
- ✅ Standard JWT library usage

---

## Login Module (`internal/login/`)

### 📁 File Structure
```
internal/login/
├── application/usecase/
│   ├── do_login_usecase.go (42 lines)
│   └── do_login_usecase_test.go (96 lines)
├── domain/
│   ├── model/
│   │   ├── user_model.go (84 lines)
│   │   └── user_model_test.go
│   ├── valueobject/
│   │   ├── email.go (128 lines)
│   │   ├── email_test.go
│   │   ├── password.go (276 lines)
│   │   └── password_test.go
│   └── repository/
│       └── i_login_repository.go (7 lines)
├── infra/
│   └── persistense/
│       ├── login_repository.go (40 lines)
│       └── login_repository_test.go
└── presentation/
    └── http/
        ├── do_login.go (43 lines)
        └── do_login_test.go
```

---

### 🔍 Domain Layer Analysis

#### `domain/model/user_model.go` (84 lines)

**Purpose**: User aggregate root with value objects

```go
type User struct {
    ID       string
    Email    *valueobject.Email
    Password *valueobject.Password
}
```

**Dependencies**:
- ✅ `HubInvestments/internal/login/domain/valueobject` - Email, Password VOs

**Methods**:
1. `NewUser(id, email, password string) (*User, error)` - With validation
2. `NewUserFromRepository(id, email, password string) *User` - Without validation
3. `GetEmailString() string`
4. `GetPasswordString() string`
5. `ChangeEmail(newEmail string) error`
6. `ChangePassword(newPassword string) error`

**Key Observations**:
- ✅ Pure domain model (no external dependencies)
- ✅ Value object composition
- ✅ Factory methods for different contexts
- ✅ JSON serialization tags (`json:"-"` for password)

---

#### `domain/valueobject/email.go` (128 lines)

**Purpose**: Email address validation and normalization

```go
type Email struct {
    value string
}
```

**Validation Rules**:
- ✅ Non-empty check
- ✅ RFC 5322 regex validation
- ✅ Length constraints (max 254 chars)
- ✅ Local part: 1-64 chars
- ✅ Domain part: 1-253 chars
- ✅ No consecutive dots
- ✅ No leading/trailing dots in local part
- ✅ Normalization: lowercase + trim

**Methods**:
1. `NewEmail(email string) (*Email, error)` - With validation
2. `NewEmailFromRepository(email string) *Email` - Without validation
3. `Value() string`
4. `Equals(other *Email) bool`
5. `Domain() string`
6. `LocalPart() string`
7. `IsValid() bool`

**Dependencies**:
- ✅ `regexp` - Email regex validation
- ✅ `strings` - String manipulation
- ✅ `errors` - Error handling

**Key Observations**:
- ✅ Comprehensive email validation
- ✅ No external dependencies
- ✅ Immutable value object
- ✅ Repository bypass for trusted data

---

#### `domain/valueobject/password.go` (276 lines)

**Purpose**: Password validation with security rules

```go
type Password struct {
    value string
}
```

**Validation Rules**:
- ✅ Min length: 8 characters
- ✅ Max length: 60 characters (prevent DOS)
- ✅ At least one uppercase letter
- ✅ At least one lowercase letter
- ✅ At least one digit
- ✅ At least one special character
- ✅ No common weak patterns
- ✅ No simple sequences (123456, abcdefgh)

**Weak Pattern Detection**:
- "password", "123456", "qwerty", "abc123", "admin", "user"
- "login", "welcome", "changeme", "default", "guest"
- Ascending/descending sequences

**Methods**:
1. `NewPassword(password string) (*Password, error)` - With validation
2. `NewPasswordFromRepository(password string) *Password` - Without validation
3. `Value() string`
4. `Equals(other *Password) bool`
5. `EqualsString(other string) bool`
6. `Length() int`
7. `Strength() int` - Returns 1-5 score
8. `HasUppercase() bool`
9. `HasSpecialChar() bool`
10. `HasDigit() bool`
11. `HasLowercase() bool`
12. `IsValid() bool`

**Dependencies**:
- ✅ `regexp` - Pattern matching
- ✅ `unicode` - Character type checking
- ✅ `errors` - Error handling

**Key Observations**:
- ✅ Strong password validation
- ✅ No external dependencies
- ✅ Password strength scoring
- ✅ Immutable value object
- ⚠️ **NOTE**: Passwords stored as plain text in value object (hashing should be at application layer)

---

#### `domain/repository/i_login_repository.go` (7 lines)

**Purpose**: Repository interface for user data access

```go
type ILoginRepository interface {
    GetUserByEmail(email string) (*model.User, error)
}
```

**Key Observations**:
- ✅ Minimal interface (Single Responsibility Principle)
- ✅ Clean dependency inversion
- ✅ No implementation details leaked

---

### 🔍 Application Layer Analysis

#### `application/usecase/do_login_usecase.go` (42 lines)

**Purpose**: User login business logic

```go
type IDoLoginUsecase interface {
    Execute(email string, password string) (*model.User, error)
}

type DoLoginUsecase struct {
    repo repository.ILoginRepository
}
```

**Dependencies**:
- ✅ `HubInvestments/internal/login/domain/model` - User model
- ✅ `HubInvestments/internal/login/domain/repository` - Repository interface
- ✅ `errors` - Error handling

**Business Logic**:
1. Fetch user by email from repository
2. Check if user exists
3. Check if user has password
4. Validate password matches
5. Return user or error

**Error Cases**:
- User not found → "user not found"
- Nil user returned → "user not found"
- Nil password → "user password not found"
- Password mismatch → "invalid password"

**Key Observations**:
- ✅ Clean use case pattern
- ✅ Single responsibility
- ✅ Proper error handling
- ✅ No direct database dependencies
- ✅ Simple, testable logic

---

### 🔍 Infrastructure Layer Analysis

#### `infra/persistense/login_repository.go` (40 lines)

**Purpose**: PostgreSQL implementation of login repository

```go
type LoginRepository struct {
    db database.Database
}

type userDTO struct {
    ID       string `db:"id"`
    Email    string `db:"email"`
    Password string `db:"password"`
}
```

**Dependencies**:
- ✅ `HubInvestments/internal/login/domain/model` - User model
- ✅ `HubInvestments/internal/login/domain/repository` - Repository interface
- ✅ `HubInvestments/shared/infra/database` - Database abstraction
- ✅ `fmt` - Error formatting

**Database Query**:
```sql
SELECT id, email, password FROM users WHERE email = $1
```

**Methods**:
1. `NewLoginRepository(db database.Database) repository.ILoginRepository`
   - Constructor with database injection

2. `GetUserByEmail(email string) (*model.User, error)`
   - Executes parameterized query (SQL injection safe ✅)
   - Maps DTO to domain model
   - Uses `NewUserFromRepository` (bypasses validation)
   - Returns formatted error on failure

**Key Observations**:
- ✅ SQL injection safe (parameterized queries)
- ✅ Clean DTO pattern
- ✅ Database abstraction (not coupled to specific driver)
- ✅ Proper error wrapping
- ✅ Repository pattern implementation

---

### 🔍 Presentation Layer Analysis

#### `presentation/http/do_login.go` (43 lines)

**Purpose**: HTTP endpoint for user login

**Dependencies**:
- ✅ `HubInvestments/pck` - Dependency injection container
- ✅ `encoding/json` - JSON serialization
- ✅ `net/http` - HTTP handling

**Request Format**:
```json
{
    "email": "user@example.com",
    "password": "SecurePassword123!"
}
```

**Response Format**:
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Flow**:
1. Decode JSON request body
2. Execute login use case
3. Generate JWT token via auth service
4. Return token in JSON response

**Error Handling**:
- Invalid JSON → 400 Bad Request
- Invalid credentials → 401 Unauthorized
- Token generation failure → 500 Internal Server Error

**Key Observations**:
- ✅ Clean HTTP handler
- ✅ Proper error status codes
- ✅ JSON response format
- ✅ Dependency injection via container
- ✅ Separation of concerns

---

## External Dependencies

### 1. JWT Library
**Package**: `github.com/golang-jwt/jwt`  
**Version**: v3.2.2+ (check go.mod)  
**Usage**: Token signing and validation  
**Migration Note**: Must be included in microservice dependencies

### 2. Configuration Package
**Package**: `HubInvestments/shared/config`  
**Purpose**: JWT secret management  
**Key Config**:
- `JWTSecret`: Environment variable `MY_JWT_SECRET`
- Default: `"default-secret-key-change-in-production"`
- **CRITICAL**: Must match between monolith and microservice

### 3. Database Abstraction
**Package**: `HubInvestments/shared/infra/database`  
**Purpose**: Database connection and query execution  
**Migration Note**: Can reuse or create microservice-specific implementation

### 4. Dependency Injection Container
**Package**: `HubInvestments/pck`  
**Purpose**: Service instantiation and wiring  
**Migration Note**: Microservice will need its own DI container

---

## Integration Points

### 1. Main Application (`main.go`)

**Current Integration** (lines 47-52):
```go
tokenService := token.NewTokenService()
aucService := auth.NewAuthService(tokenService)

verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
    return aucService.VerifyToken(token, w)
})
```

**Usage**:
- Login endpoint: `/login` → `doLoginHandler.DoLogin()`
- Protected endpoints use `verifyToken` middleware

**Migration Impact**:
- ⚠️ This will need to be replaced with gRPC client
- ⚠️ All protected endpoints depend on this

---

### 2. Dependency Injection Container (`pck/container.go`)

**Current Dependencies**:
```go
type Container interface {
    DoLoginUsecase() doLoginUsecase.IDoLoginUsecase
    GetAuthService() auth.IAuthService
    // ... other services
}
```

**Implementation** (lines 296-306):
```go
loginRepo := loginPersistence.NewLoginRepository(db)
loginUsecase := doLoginUsecase.NewDoLoginUsecase(loginRepo)
tokenService := token.NewTokenService()
authService := auth.NewAuthService(tokenService)
```

**Migration Impact**:
- ⚠️ Container will need to be updated to use gRPC client
- ⚠️ Login use case will be removed from monolith
- ⚠️ Auth service will be replaced with adapter

---

### 3. Protected Endpoints

**Current Usage**:
All these endpoints use `verifyToken`:
- `/getAucAggregation`
- `/getBalance`
- `/getPortfolioSummary`
- `/getMarketData`
- `/getWatchlist`
- `/orders/*`
- `/admin/*`

**Implementation Pattern**:
```go
http.HandleFunc("/getBalance", balanceHandler.GetBalanceWithAuth(verifyToken, container))
```

**Migration Impact**:
- ✅ No changes needed to these endpoints
- ✅ They will continue to use same `verifyToken` interface
- ✅ Only the implementation behind `verifyToken` changes

---

### 4. Middleware (`shared/middleware/auth_middleware.go`)

**Current Implementation**:
```go
type TokenVerifier func(string, http.ResponseWriter) (string, error)

func WithAuthentication(verifyToken TokenVerifier, handler AuthenticatedHandler) http.HandlerFunc {
    tokenString := r.Header.Get("Authorization")
    userId, err := verifyToken(tokenString, w)
    // ... handle authentication
}
```

**Migration Impact**:
- ✅ No changes needed to middleware
- ✅ Interface remains the same
- ✅ gRPC adapter will implement `TokenVerifier` signature

---

## JWT Token Structure

### Token Claims
```go
{
    "username": "user@example.com",  // Email address
    "userId":   "user-uuid-string",  // User ID (string, not int)
    "exp":      1234567890           // Unix timestamp
}
```

### Token Format
- **Algorithm**: HS256 (HMAC-SHA256)
- **Expiration**: 10 minutes
- **Header**: `Authorization: Bearer <token>`
- **Secret**: Environment variable `MY_JWT_SECRET`

### Critical Requirements for Migration
1. ✅ **Secret must match** between monolith and microservice
2. ✅ **Claims structure must match** exactly
3. ✅ **Algorithm must be HS256**
4. ✅ **Expiration must be 10 minutes**
5. ✅ **Bearer prefix must be handled** correctly

---

## Database Schema

### Users Table
```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT non_empty_name CHECK (LENGTH(TRIM(name)) > 0),
    CONSTRAINT non_empty_password CHECK (LENGTH(password) >= 6)
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
```

**Migration Note**:
- ✅ Migration file exists: `shared/infra/migration/sql/000001_create_users_table.up.sql`
- ✅ Can be copied as-is to microservice
- ✅ Trigger for `updated_at` included

---

## Migration Checklist

### Files to Copy (AS-IS)

#### Authentication Module
- [ ] `internal/auth/auth_service.go` → `internal/core/auth_service.go`
- [ ] `internal/auth/auth_service_test.go` → `internal/core/auth_service_test.go`
- [ ] `internal/auth/token/token_service.go` → `internal/core/token_service.go`
- [ ] `internal/auth/token/token_service_test.go` → `internal/core/token_service_test.go`

#### Login Module - Domain
- [ ] `internal/login/domain/model/user_model.go` → `internal/domain/model/user.go`
- [ ] `internal/login/domain/model/user_model_test.go` → `internal/domain/model/user_test.go`
- [ ] `internal/login/domain/valueobject/email.go` → `internal/domain/valueobject/email.go`
- [ ] `internal/login/domain/valueobject/email_test.go` → `internal/domain/valueobject/email_test.go`
- [ ] `internal/login/domain/valueobject/password.go` → `internal/domain/valueobject/password.go`
- [ ] `internal/login/domain/valueobject/password_test.go` → `internal/domain/valueobject/password_test.go`
- [ ] `internal/login/domain/repository/i_login_repository.go` → `internal/domain/repository/user_repository.go`

#### Login Module - Application
- [ ] `internal/login/application/usecase/do_login_usecase.go` → `internal/usecase/login_usecase.go`
- [ ] `internal/login/application/usecase/do_login_usecase_test.go` → `internal/usecase/login_usecase_test.go`

#### Login Module - Infrastructure
- [ ] `internal/login/infra/persistense/login_repository.go` → `internal/repository/postgres_user_repository.go`
- [ ] `internal/login/infra/persistense/login_repository_test.go` → `internal/repository/postgres_user_repository_test.go`

#### Database Migration
- [ ] `shared/infra/migration/sql/000001_create_users_table.up.sql` → `migrations/000001_create_users_table.up.sql`
- [ ] `shared/infra/migration/sql/000001_create_users_table.down.sql` → `migrations/000001_create_users_table.down.sql`

#### gRPC Proto
- [ ] `shared/grpc/proto/auth_service.proto` → `internal/grpc/proto/auth_service.proto`

### Import Path Changes Required

**Old Import Paths**:
```go
"HubInvestments/internal/auth"
"HubInvestments/internal/auth/token"
"HubInvestments/internal/login/domain/model"
"HubInvestments/internal/login/domain/valueobject"
"HubInvestments/internal/login/domain/repository"
"HubInvestments/shared/config"
"HubInvestments/shared/infra/database"
```

**New Import Paths**:
```go
"hub-user-service/internal/core"
"hub-user-service/internal/domain/model"
"hub-user-service/internal/domain/valueobject"
"hub-user-service/internal/domain/repository"
"hub-user-service/config"
"hub-user-service/internal/repository"
```

### Dependencies to Install

```bash
go get github.com/golang-jwt/jwt@latest
go get github.com/lib/pq@latest
go get google.golang.org/grpc@latest
go get google.golang.org/protobuf@latest
```

---

## Summary

### Total Code to Migrate
- **Source Files**: 16 files
- **Test Files**: 8 files
- **Total Lines**: ~800 lines
- **External Dependencies**: 3 packages

### Complexity Assessment
- ✅ **Low Complexity**: Clean architecture, minimal dependencies
- ✅ **Well-Tested**: Comprehensive test coverage
- ✅ **Clear Boundaries**: No tight coupling
- ✅ **Reusable Migration**: Existing migration files

### Risk Assessment
- ✅ **Low Risk**: Simple domain logic
- ⚠️ **Medium Risk**: JWT token compatibility (must test thoroughly)
- ✅ **Low Risk**: Database schema (already defined)

### Estimated Effort
- **Code Migration**: 2-3 hours (copy + update imports)
- **gRPC Implementation**: 4-6 hours
- **Testing**: 4-6 hours
- **Total**: ~12-15 hours

---

## Next Steps

✅ **Step 1.1**: Deep Code Analysis - **COMPLETED**  
⏭️ **Step 1.2**: Database Schema Analysis  
⏭️ **Step 1.3**: Integration Point Mapping  
⏭️ **Step 1.4**: JWT Token Compatibility Analysis  
⏭️ **Step 1.5**: Test Inventory

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-13  
**Author**: AI Assistant  
**Status**: ✅ Ready for Step 1.2

