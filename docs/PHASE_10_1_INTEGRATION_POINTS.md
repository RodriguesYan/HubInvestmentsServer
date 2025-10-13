# Phase 10.1 - Integration Point Mapping
## Hub User Service Migration - Integration Analysis

**Date**: 2025-10-13  
**Status**: Step 1.3 - COMPLETED ✅  
**Deliverable**: Integration point diagram showing all auth touchpoints

---

## 📋 Table of Contents
1. [Executive Summary](#executive-summary)
2. [AuthService Integration Points](#authservice-integration-points)
3. [Token Verifier Integration Points](#token-verifier-integration-points)
4. [Container Integration Points](#container-integration-points)
5. [Authentication Flow Diagram](#authentication-flow-diagram)
6. [Migration Impact Analysis](#migration-impact-analysis)
7. [Code Changes Required](#code-changes-required)

---

## Executive Summary

### Total Integration Points Identified
- **`VerifyToken()` calls**: 3 direct calls + 1 WebSocket
- **`CreateToken()` calls**: 2 direct calls
- **`TokenVerifier` usage**: 12 protected endpoints
- **Container dependencies**: 2 methods

### Migration Impact
- ✅ **Minimal Changes Required**: Only main.go and container.go need updates
- ✅ **Protected Endpoints**: No changes needed (interface remains same)
- ✅ **Test Files**: Can be ignored (won't be in production)

---

## AuthService Integration Points

### 1. Direct `VerifyToken()` Calls

#### **Location 1: main.go (Line 50-52)**
```go
verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
    return aucService.VerifyToken(token, w)
})
```

**Purpose**: Creates token verifier function for middleware  
**Migration Impact**: ⚠️ **HIGH** - Must replace with gRPC adapter  
**Change Required**: ✅ YES

---

#### **Location 2: realtime_quotes/infra/websocket/realtime_quotes_websocket_handler.go (Line 69)**
```go
userId, err := h.authService.VerifyToken(tokenString, w)
if err != nil {
    log.Printf("WebSocket Debug: Authentication failed with error: %v", err)
    // authService.VerifyToken already wrote the HTTP error response
    return
}
```

**Purpose**: WebSocket authentication  
**Context**: WebSocket handler for real-time quotes  
**Migration Impact**: ⚠️ **MEDIUM** - Will use injected auth service from container  
**Change Required**: ✅ NO (uses container's auth service)

---

### 2. Direct `CreateToken()` Calls

#### **Location 1: login/presentation/http/do_login.go (Line 32)**
```go
// Generate token
tokenString, err := container.GetAuthService().CreateToken(user.Email.Value(), user.ID)
if err != nil {
    http.Error(w, "Failed to generate token", http.StatusInternalServerError)
    return
}
```

**Purpose**: Token generation after successful login  
**Flow**: User login → Validate credentials → Create token → Return token  
**Migration Impact**: ⚠️ **LOW** - Uses container's auth service  
**Change Required**: ✅ NO (uses container's auth service)

---

#### **Location 2: shared/grpc/auth_server.go (Line 55)**
```go
authService := s.container.GetAuthService()
token, err := authService.CreateToken(user.Email.Value(), user.ID)
if err != nil {
    return &proto.LoginResponse{
        ApiResponse: &proto.APIResponse{
            Success: false,
            Message: "Failed to generate token",
        },
    }, nil
}
```

**Purpose**: Token generation in gRPC auth service  
**Context**: gRPC endpoint for authentication  
**Migration Impact**: ⚠️ **LOW** - Uses container's auth service  
**Change Required**: ✅ NO (uses container's auth service)

---

### 3. Login Use Case Integration

#### **Location 1: login/presentation/http/do_login.go (Line 24)**
```go
// Authenticate user
user, err := container.DoLoginUsecase().Execute(loginRequest.Email, loginRequest.Password)

if err != nil {
    http.Error(w, "Invalid credentials", http.StatusUnauthorized)
    return
}
```

**Purpose**: User authentication  
**Migration Impact**: ⚠️ **HIGH** - Will need to call microservice  
**Change Required**: ✅ YES - Move to microservice

---

#### **Location 2: shared/grpc/auth_server.go (Line 41-42)**
```go
loginUseCase := s.container.DoLoginUsecase()
user, err := loginUseCase.Execute(req.Email, req.Password)
if err != nil {
    log.Printf("Login failed for user %s: %v", req.Email, err)
    return &proto.LoginResponse{
        ApiResponse: &proto.APIResponse{
            Success: false,
            Message: "Invalid credentials",
        },
    }, nil
}
```

**Purpose**: gRPC login implementation  
**Migration Impact**: ⚠️ **MEDIUM** - Already isolated in gRPC service  
**Change Required**: ✅ NO (this is part of existing gRPC, may be replaced)

---

## Token Verifier Integration Points

### Protected HTTP Endpoints Using `verifyToken`

All these endpoints in **main.go** use the `verifyToken` middleware:

#### **1. Position/Portfolio Endpoints**
```go
// Line 69
http.HandleFunc("/getAucAggregation", 
    positionHandler.GetAucAggregationWithAuth(verifyToken, container))
```

**Purpose**: Get position aggregation  
**Migration Impact**: ✅ **NONE** - Interface remains same  
**Change Required**: ✅ NO

---

#### **2. Balance Endpoint**
```go
// Line 70
http.HandleFunc("/getBalance", 
    balanceHandler.GetBalanceWithAuth(verifyToken, container))
```

**Purpose**: Get user balance  
**Migration Impact**: ✅ **NONE** - Interface remains same  
**Change Required**: ✅ NO

---

#### **3. Portfolio Summary Endpoint**
```go
// Line 71
http.HandleFunc("/getPortfolioSummary", 
    portfolioSummaryHandler.GetPortfolioSummaryWithAuth(verifyToken, container))
```

**Purpose**: Get portfolio summary  
**Migration Impact**: ✅ **NONE** - Interface remains same  
**Change Required**: ✅ NO

---

#### **4. Market Data Endpoint**
```go
// Line 72
http.HandleFunc("/getMarketData", 
    marketDataHandler.GetMarketDataWithAuth(verifyToken, container))
```

**Purpose**: Get market data  
**Migration Impact**: ✅ **NONE** - Interface remains same  
**Change Required**: ✅ NO

---

#### **5. Watchlist Endpoint**
```go
// Line 73
http.HandleFunc("/getWatchlist", 
    watchlistHandler.GetWatchlistWithAuth(verifyToken, container))
```

**Purpose**: Get user watchlist  
**Migration Impact**: ✅ **NONE** - Interface remains same  
**Change Required**: ✅ NO

---

#### **6. Order Management Endpoints**

**Submit Order** (Line 76):
```go
http.HandleFunc("/orders", 
    orderHandler.SubmitOrderWithAuth(verifyToken, container))
```

**Order Details** (Line 77-86):
```go
http.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    if strings.HasSuffix(path, "/status") {
        orderHandler.GetOrderStatusWithAuth(verifyToken, container)(w, r)
    } else if strings.HasSuffix(path, "/cancel") {
        orderHandler.CancelOrderWithAuth(verifyToken, container)(w, r)
    } else {
        orderHandler.GetOrderDetailsWithAuth(verifyToken, container)(w, r)
    }
})
```

**Order History** (Line 87):
```go
http.HandleFunc("/orders/history", 
    orderHandler.GetOrderHistoryWithAuth(verifyToken, container))
```

**Total Order Endpoints**: 5 endpoints  
**Migration Impact**: ✅ **NONE** - Interface remains same  
**Change Required**: ✅ NO

---

#### **7. Admin Cache Management Endpoints**

**Cache Invalidation** (Line 90):
```go
http.HandleFunc("/admin/market-data/cache/invalidate", 
    adminHandler.AdminInvalidateCacheWithAuth(verifyToken, container))
```

**Cache Warming** (Line 91):
```go
http.HandleFunc("/admin/market-data/cache/warm", 
    adminHandler.AdminWarmCacheWithAuth(verifyToken, container))
```

**Total Admin Endpoints**: 2 endpoints  
**Migration Impact**: ✅ **NONE** - Interface remains same  
**Change Required**: ✅ NO

---

### Summary of Protected Endpoints

| Endpoint | Handler | Migration Impact |
|----------|---------|------------------|
| `/getAucAggregation` | Position | ✅ No change |
| `/getBalance` | Balance | ✅ No change |
| `/getPortfolioSummary` | Portfolio | ✅ No change |
| `/getMarketData` | Market Data | ✅ No change |
| `/getWatchlist` | Watchlist | ✅ No change |
| `/orders` (POST) | Submit Order | ✅ No change |
| `/orders/{id}` (GET) | Order Details | ✅ No change |
| `/orders/{id}/status` (GET) | Order Status | ✅ No change |
| `/orders/{id}/cancel` (PUT) | Cancel Order | ✅ No change |
| `/orders/history` (GET) | Order History | ✅ No change |
| `/admin/market-data/cache/invalidate` | Admin | ✅ No change |
| `/admin/market-data/cache/warm` | Admin | ✅ No change |

**Total Protected Endpoints**: 12  
**Endpoints Requiring Changes**: 0 ✅

---

## Container Integration Points

### Container Interface Methods

#### **1. GetAuthService() Method**

**Interface Definition** (pck/container.go Line 47):
```go
type Container interface {
    GetAuthService() auth.IAuthService
    // ... other methods
}
```

**Implementation** (pck/container.go Line 140-142):
```go
func (c *containerImpl) GetAuthService() auth.IAuthService {
    return c.AuthService
}
```

**Initialization** (pck/container.go Line 305-306):
```go
tokenService := token.NewTokenService()
authService := auth.NewAuthService(tokenService)
```

**Used By**:
- ✅ `do_login.go` - Token creation after login
- ✅ `auth_server.go` - gRPC token creation
- ✅ `realtime_quotes_websocket_handler.go` - WebSocket authentication

**Migration Impact**: ⚠️ **HIGH**  
**Change Required**: ✅ YES - Replace with gRPC adapter

---

#### **2. DoLoginUsecase() Method**

**Interface Definition** (pck/container.go Line 46):
```go
type Container interface {
    DoLoginUsecase() doLoginUsecase.IDoLoginUsecase
    // ... other methods
}
```

**Implementation** (pck/container.go Line 172-174):
```go
func (c *containerImpl) DoLoginUsecase() doLoginUsecase.IDoLoginUsecase {
    return c.LoginUsecase
}
```

**Initialization** (pck/container.go Line 303-304):
```go
loginRepo := loginPersistence.NewLoginRepository(db)
loginUsecase := doLoginUsecase.NewDoLoginUsecase(loginRepo)
```

**Used By**:
- ✅ `do_login.go` - User authentication
- ✅ `auth_server.go` - gRPC login

**Migration Impact**: ⚠️ **HIGH**  
**Change Required**: ✅ YES - Will be handled by microservice

---

## Authentication Flow Diagram

### Current Authentication Flow (Monolith)

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT REQUEST                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    1. LOGIN ENDPOINT                             │
│                   POST /login                                    │
│   { "email": "user@example.com", "password": "pass" }          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              2. DO LOGIN HANDLER (do_login.go)                   │
│   - Decodes JSON request                                        │
│   - Calls container.DoLoginUsecase().Execute()                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│           3. LOGIN USE CASE (do_login_usecase.go)                │
│   - Fetches user from repository by email                       │
│   - Validates password                                           │
│   - Returns User model                                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│        4. LOGIN REPOSITORY (login_repository.go)                 │
│   - Queries database: SELECT id, email, password FROM users     │
│   - Returns User DTO → User model                               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    5. AUTH SERVICE                               │
│   container.GetAuthService().CreateToken(email, userId)         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              6. TOKEN SERVICE (token_service.go)                 │
│   - Creates JWT token with HS256                                │
│   - Claims: username, userId, exp (10 min)                      │
│   - Signs with config.JWTSecret                                 │
│   - Returns token string                                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  7. RESPONSE TO CLIENT                           │
│           { "token": "eyJhbGci..." }                            │
└─────────────────────────────────────────────────────────────────┘
```

---

### Current Protected Endpoint Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                   CLIENT REQUEST                                 │
│   GET /getBalance                                                │
│   Authorization: Bearer eyJhbGci...                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              1. HTTP HANDLER (main.go)                           │
│   http.HandleFunc("/getBalance",                                │
│       balanceHandler.GetBalanceWithAuth(verifyToken, container))│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│         2. MIDDLEWARE (auth_middleware.go)                       │
│   WithAuthentication(verifyToken, handler)                      │
│   - Extracts "Authorization" header                             │
│   - Calls verifyToken(tokenString, w)                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│         3. VERIFY TOKEN FUNCTION (main.go)                       │
│   verifyToken := func(token, w) {                               │
│       return aucService.VerifyToken(token, w)                   │
│   }                                                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│           4. AUTH SERVICE (auth_service.go)                      │
│   - Checks if token is empty                                    │
│   - Calls tokenService.ValidateToken()                          │
│   - Extracts userId from claims                                 │
│   - Returns userId                                               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│          5. TOKEN SERVICE (token_service.go)                     │
│   - Strips "Bearer " prefix                                     │
│   - Parses JWT with config.JWTSecret                            │
│   - Validates signature and expiration                          │
│   - Returns claims map                                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│         6. BUSINESS HANDLER (balance_handler.go)                 │
│   GetBalance(w, r, userId, container)                           │
│   - Uses userId to fetch user data                              │
│   - Returns business response                                    │
└─────────────────────────────────────────────────────────────────┘
```

---

### Future Authentication Flow (After Migration)

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT REQUEST                            │
│                      POST /login                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    MONOLITH (main.go)                            │
│              DO LOGIN HANDLER (proxy)                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ gRPC Call
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              HUB-USER-SERVICE (Microservice)                     │
│                   gRPC: Login(email, password)                   │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │  1. Validate credentials                                 │  │
│   │  2. Query database                                       │  │
│   │  3. Generate JWT token                                   │  │
│   │  4. Return token                                         │  │
│   └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  RESPONSE TO CLIENT                              │
│              { "token": "eyJhbGci..." }                         │
└─────────────────────────────────────────────────────────────────┘


For Protected Endpoints:
┌─────────────────────────────────────────────────────────────────┐
│   CLIENT REQUEST: GET /getBalance                                │
│   Authorization: Bearer eyJhbGci...                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│           MONOLITH MIDDLEWARE (verifyToken)                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ gRPC Call
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│         HUB-USER-SERVICE (Microservice)                          │
│         gRPC: ValidateToken(token)                               │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │  1. Parse JWT token                                      │  │
│   │  2. Validate signature                                   │  │
│   │  3. Check expiration                                     │  │
│   │  4. Return userId                                        │  │
│   └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│      MONOLITH BUSINESS HANDLER (balance_handler.go)              │
│      Continues with userId → fetch data → respond                │
└─────────────────────────────────────────────────────────────────┘
```

---

## Migration Impact Analysis

### High Impact (Requires Changes)

#### **1. main.go - Authentication Service Initialization**

**Current Code** (Lines 47-52):
```go
tokenService := token.NewTokenService()
aucService := auth.NewAuthService(tokenService)

verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
    return aucService.VerifyToken(token, w)
})
```

**New Code** (After Migration):
```go
// Connect to user microservice via gRPC
authClient, err := grpc.NewAuthClient("localhost:50051")
if err != nil {
    log.Fatal("Failed to connect to auth service:", err)
}
aucService := auth.NewGRPCAuthServiceAdapter(authClient)

verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
    return aucService.VerifyToken(token, w)
})
```

**Lines Changed**: ~6 lines  
**Complexity**: Low  
**Risk**: Medium (core authentication)

---

#### **2. pck/container.go - Auth Service Initialization**

**Current Code** (Lines 305-306):
```go
tokenService := token.NewTokenService()
authService := auth.NewAuthService(tokenService)
```

**New Code** (After Migration):
```go
// Initialize gRPC client for user microservice
authClient, err := grpc.NewAuthClient(os.Getenv("USER_SERVICE_GRPC_ADDR"))
if err != nil {
    return nil, fmt.Errorf("failed to connect to user service: %w", err)
}
authService := auth.NewGRPCAuthServiceAdapter(authClient)
```

**Lines Changed**: ~5 lines  
**Complexity**: Low  
**Risk**: Low (isolated in container)

---

#### **3. pck/container.go - Login Use Case Removal**

**Current Code** (Lines 303-304):
```go
loginRepo := loginPersistence.NewLoginRepository(db)
loginUsecase := doLoginUsecase.NewDoLoginUsecase(loginRepo)
```

**Action**: Remove these lines (login moves to microservice)

**Lines Changed**: -2 lines (deletion)  
**Complexity**: Low  
**Risk**: Low (unused after migration)

---

### Medium Impact (Indirect Changes)

#### **4. Create gRPC Adapter**

**New File**: `internal/auth/grpc_auth_adapter.go`

**Purpose**: Adapter that implements `auth.IAuthService` interface and calls microservice via gRPC

**Code** (New):
```go
package auth

import (
    "context"
    "net/http"
    "HubInvestments/shared/grpc"
)

type GRPCAuthServiceAdapter struct {
    grpcClient *grpc.AuthClient
}

func NewGRPCAuthServiceAdapter(client *grpc.AuthClient) IAuthService {
    return &GRPCAuthServiceAdapter{grpcClient: client}
}

func (a *GRPCAuthServiceAdapter) VerifyToken(tokenString string, w http.ResponseWriter) (string, error) {
    resp, err := a.grpcClient.ValidateToken(context.Background(), tokenString)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        return "", err
    }
    return resp.UserInfo.UserId, nil
}

func (a *GRPCAuthServiceAdapter) CreateToken(userName string, userId string) (string, error) {
    // This method may not be needed if login is fully in microservice
    // For now, forward to microservice
    resp, err := a.grpcClient.Login(context.Background(), &pb.LoginRequest{
        Email: userName,
        // Password not needed for token creation
    })
    if err != nil {
        return "", err
    }
    return resp.Token, nil
}
```

**Lines Added**: ~40 lines  
**Complexity**: Low  
**Risk**: Low (wrapper only)

---

### Low/No Impact (No Changes Required)

#### **5. All Protected Endpoints**

**Endpoints**: 12 endpoints (balance, portfolio, orders, admin, etc.)

**Why No Changes**:
- ✅ All use `verifyToken` middleware
- ✅ `verifyToken` implements `TokenVerifier` interface
- ✅ Interface signature unchanged
- ✅ Only the implementation behind `verifyToken` changes

**Impact**: ✅ **ZERO** - Transparent to business logic

---

#### **6. Middleware (auth_middleware.go)**

**Current Code**:
```go
type TokenVerifier func(string, http.ResponseWriter) (string, error)

func WithAuthentication(verifyToken TokenVerifier, handler AuthenticatedHandler) http.HandlerFunc {
    // ... implementation
}
```

**Why No Changes**:
- ✅ Interface remains exactly the same
- ✅ Middleware doesn't care about implementation
- ✅ Adapter pattern ensures compatibility

**Impact**: ✅ **ZERO** - No changes needed

---

## Code Changes Required

### Summary of Changes

| File | Action | Lines Changed | Complexity | Risk |
|------|--------|---------------|------------|------|
| `main.go` | Modify | ~6 lines | Low | Medium |
| `pck/container.go` | Modify | ~7 lines | Low | Low |
| `internal/auth/grpc_auth_adapter.go` | Create | ~40 lines | Low | Low |
| **Protected Endpoints** | **None** | **0 lines** | **N/A** | **None** |
| **Middleware** | **None** | **0 lines** | **N/A** | **None** |
| **Total** | **3 files** | **~53 lines** | **Low** | **Low** |

---

### Detailed Change Checklist

#### **Phase 1: Create Adapter (Week 6)**

- [ ] **Step 1**: Create `internal/auth/grpc_auth_adapter.go`
  - [ ] Implement `GRPCAuthServiceAdapter` struct
  - [ ] Implement `VerifyToken()` method (calls gRPC)
  - [ ] Implement `CreateToken()` method (calls gRPC)
  - [ ] Implement `IAuthService` interface
  - [ ] Add error handling
  - [ ] **Deliverable**: Adapter file created

#### **Phase 2: Update main.go (Week 6)**

- [ ] **Step 2**: Modify `main.go` authentication initialization
  - [ ] Remove local `tokenService` and `authService` creation
  - [ ] Add gRPC client initialization
  - [ ] Create `GRPCAuthServiceAdapter` instance
  - [ ] Update `verifyToken` to use adapter
  - [ ] Test authentication flow
  - [ ] **Deliverable**: main.go updated

#### **Phase 3: Update Container (Week 6)**

- [ ] **Step 3**: Modify `pck/container.go`
  - [ ] Add gRPC client initialization in `NewContainer()`
  - [ ] Replace local auth service with gRPC adapter
  - [ ] Remove login repository initialization
  - [ ] Remove login use case initialization
  - [ ] Update container struct (remove login fields)
  - [ ] **Deliverable**: Container updated

#### **Phase 4: Testing (Week 6)**

- [ ] **Step 4**: Test all integration points
  - [ ] Test login endpoint
  - [ ] Test all 12 protected endpoints
  - [ ] Test WebSocket authentication
  - [ ] Test token expiration
  - [ ] Test error scenarios
  - [ ] **Deliverable**: All tests passing

---

## Dependency Graph

```
┌─────────────────────────────────────────────────────────────────┐
│                           main.go                                │
│                                                                  │
│   ┌──────────────────────────────────────────────────────────┐ │
│   │  1. Creates verifyToken function                         │ │
│   │  2. Uses container.GetAuthService()                      │ │
│   │  3. Passes verifyToken to all protected endpoints       │ │
│   └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ depends on
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      pck/container.go                            │
│                                                                  │
│   ┌──────────────────────────────────────────────────────────┐ │
│   │  1. Initializes AuthService                              │ │
│   │  2. Initializes LoginUsecase                             │ │
│   │  3. Provides GetAuthService() method                     │ │
│   │  4. Provides DoLoginUsecase() method                     │ │
│   └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ provides
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  internal/auth/auth_service.go                   │
│                                                                  │
│   ┌──────────────────────────────────────────────────────────┐ │
│   │  IAuthService interface:                                 │ │
│   │    - VerifyToken(token, w) → userId                      │ │
│   │    - CreateToken(username, userId) → token              │ │
│   └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ used by
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Protected Endpoints (12 total)                  │
│                                                                  │
│   ┌──────────────────────────────────────────────────────────┐ │
│   │  - /getBalance                                           │ │
│   │  - /getPortfolioSummary                                  │ │
│   │  - /orders                                               │ │
│   │  - /admin/market-data/cache/*                           │ │
│   │  - ... and 8 more endpoints                             │ │
│   └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## Summary

### Integration Points Summary

| Category | Count | Migration Impact |
|----------|-------|------------------|
| **Direct Auth Calls** | 4 calls | Low (container handles it) |
| **Protected Endpoints** | 12 endpoints | ✅ None |
| **Container Methods** | 2 methods | Medium (needs update) |
| **Files Requiring Changes** | 3 files | Low complexity |

### Risk Assessment

| Risk Level | Count | Details |
|------------|-------|---------|
| **High** | 0 | No high-risk changes |
| **Medium** | 2 | main.go, container.go |
| **Low** | 1 | Adapter creation |
| **None** | 12 | Protected endpoints |

### Effort Estimation

| Task | Estimated Time | Complexity |
|------|---------------|------------|
| Create gRPC adapter | 2-3 hours | Low |
| Update main.go | 1 hour | Low |
| Update container.go | 1-2 hours | Low |
| Testing | 3-4 hours | Medium |
| **Total** | **7-10 hours** | **Low** |

### Key Insights

1. ✅ **Minimal Impact**: Only 3 files need changes
2. ✅ **Interface Stability**: `IAuthService` interface doesn't change
3. ✅ **Protected Endpoints**: Zero changes needed (12 endpoints)
4. ✅ **Adapter Pattern**: Clean separation via adapter
5. ✅ **Low Risk**: Most code remains unchanged

---

## Next Steps

✅ **Step 1.3**: Integration Point Mapping - **COMPLETED**  
⏭️ **Step 1.4**: JWT Token Compatibility Analysis  
⏭️ **Step 1.5**: Test Inventory

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-13  
**Author**: AI Assistant  
**Status**: ✅ Ready for Step 1.4

