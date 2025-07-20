# Database Abstraction Layer

This package provides a database abstraction layer that encapsulates SQL operations, making it easy to switch between different SQL packages in the future without changing repository implementations.

## Problem Solved

Previously, repositories directly depended on specific SQL packages like SQLX:

```go
// Old approach - direct SQLX dependency
type SQLXPositionRepository struct {
    db *sqlx.DB  // Direct dependency on SQLX
}

func (r *SQLXPositionRepository) GetPositions(userId string) ([]Position, error) {
    return r.db.Select(&positions, query, userId)  // SQLX-specific method
}
```

If you wanted to switch from SQLX to another SQL package, you would need to:
1. Change all repository implementations
2. Update dependency injection containers
3. Modify every place where database connections are created
4. Potentially break existing tests

## Solution

The database abstraction layer provides:

1. **Generic Database Interface**: All repositories depend on the `Database` interface, not SQLX directly
2. **Single Point of Change**: To switch SQL packages, you only need to change the connection factory
3. **Future-Proof**: Easy to add support for GORM, Ent, or any other SQL package when needed
4. **Backward Compatibility**: Existing code continues to work without changes

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Repository    │    │   Database       │    │ SQLX            │
│                 │───▶│   Interface      │◄───│ Implementation  │
│ (Position,      │    │                  │    │                 │
│  Balance, etc.) │    │ - Query()        │    │ (Currently)     │
└─────────────────┘    │ - Get()          │    └─────────────────┘
                       │ - Select()       │
                       │ - Exec()         │
                       └──────────────────┘
```

## Usage

### Basic Usage (Drop-in Replacement)

The abstraction layer is designed to be a drop-in replacement for existing SQLX usage:

```go
// Before
db, err := sqlx.Connect("postgres", dsn)
repo := NewSQLXPositionRepository(db)

// After - using abstraction
db, err := database.CreateDatabaseConnection()
repo := positionPersistence.NewPositionRepository(db)
```

### Creating Repositories

```go
package persistence

import (
    "HubInvestments/shared/infra/database"
    domain "HubInvestments/position/domain/model"
)

type PositionRepository struct {
    db database.Database  // Use the interface, not *sqlx.DB directly
}

func NewPositionRepository(db database.Database) repository.PositionRepository {
    return &PositionRepository{db: db}
}

func (r *PositionRepository) GetPositions(userId string) ([]domain.Position, error) {
    query := "SELECT * FROM positions WHERE user_id = $1"
    
    var positions []domain.Position
    err := r.db.Select(&positions, query, userId)  // Same API as SQLX
    return positions, err
}
```

## Current Implementation

### SQLX Implementation
- **File**: `sqlx_database.go`
- **Benefits**: Full SQLX feature set, struct scanning, named parameters
- **Status**: Current implementation, maintains 100% compatibility with existing SQLX code

## Future Extensibility

When you need to add support for a new SQL package (e.g., GORM), you would:

1. **Create the implementation file** (`gorm_database.go`):
```go
type GORMDatabase struct {
    db *gorm.DB
}

func (g *GORMDatabase) Get(dest interface{}, query string, args ...interface{}) error {
    return g.db.Raw(query, args...).Scan(dest).Error
}

// Implement all other Database interface methods...
```

2. **Update the connection factory** to support the new implementation:
```go
func (cf *ConnectionFactory) CreateConnection() (Database, error) {
    // Add logic to choose between SQLX, GORM, etc.
    // For now, it only creates SQLX connections
}
```

3. **All existing repositories continue to work** without any changes!

## Migration Guide

### Migrating Existing Repositories

1. **Update imports**:
```go
// Remove
import "github.com/jmoiron/sqlx"

// Add
import "HubInvestments/shared/infra/database"
```

2. **Change struct fields**:
```go
// Before
type Repository struct {
    db *sqlx.DB
}

// After
type Repository struct {
    db database.Database
}
```

3. **Update constructor**:
```go
// Before
func NewRepository(db *sqlx.DB) RepositoryInterface {
    return &Repository{db: db}
}

// After
func NewRepository(db database.Database) RepositoryInterface {
    return &Repository{db: db}
}
```

4. **Method calls remain exactly the same**:
```go
// These calls work identically with the abstraction
err := r.db.Get(&result, query, args...)
err := r.db.Select(&results, query, args...)
err := r.db.Exec(query, args...)
```

### Updating Dependency Injection

```go
// Before
db, err := sqlx.Connect("postgres", dsn)
repo := persistence.NewSQLXRepository(db)

// After
db, err := database.CreateDatabaseConnection()
repo := persistence.NewRepository(db)
```

## Testing

The abstraction layer makes testing easier by allowing you to inject mock implementations:

```go
type MockDatabase struct {
    mock.Mock
}

func (m *MockDatabase) Get(dest interface{}, query string, args ...interface{}) error {
    args := m.Called(dest, query, args)
    return args.Error(0)
}

// In tests
mockDB := &MockDatabase{}
repo := NewRepository(mockDB)
```

## Benefits

1. **Single Point of Change**: Switch SQL packages by changing only the connection factory
2. **Technology Independence**: Repositories don't depend on specific SQL packages
3. **Easy Testing**: Inject mock implementations for unit tests
4. **Incremental Migration**: Migrate repositories one at a time
5. **Future-Proof**: Add new SQL packages without breaking existing code
6. **Performance**: No runtime overhead, interfaces are compiled away
7. **100% SQLX Compatibility**: All existing SQLX features work unchanged

## File Structure

```
shared/infra/database/
├── README.md                 # This documentation
├── database.go              # Interface definitions
├── connection_factory.go    # Connection management and factory
└── sqlx_database.go         # SQLX implementation
```

## Examples

See the updated repositories in:
- `position/infra/persistence/position_repository.go`
- `balance/infra/persistence/balance_repository.go`
- `login/login_refactored.go`

These show how to use the database abstraction in practice.

## Summary

You now have a clean database abstraction layer that:
- ✅ **Encapsulates SQLX** in one place
- ✅ **Maintains full compatibility** with existing code
- ✅ **Provides a single point of change** for future SQL package switches
- ✅ **Makes testing easier** with dependency injection
- ✅ **Future-proofs your codebase** for when you need to add new SQL packages

When you're ready to add support for GORM, Ent, or any other SQL package, you can do so without touching any repository code! 