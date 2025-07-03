- ver artigo do bryan de novo pra ver se eu to implementando interface corretamente

## DDD Implementation Improvement Plan

### ✅ Step 1: Move HTTP Handler to Proper Location (COMPLETED)
- [x] Move `get_aggregation.go` from root of position module to `position/presentation/http/`
- [x] Rename package from `get_aggregation` to `http`
- [x] Update import paths in `main.go`

### ✅ Step 2: Extract Business Logic from HTTP Handler (COMPLETED)
- [x] Create `position/application/usecase/get_position_aggregation_usecase.go`
- [x] Move all business logic (aggregation, calculations, sorting) from HTTP handler to use case
- [x] Update HTTP handler to only handle HTTP concerns (auth, serialization, error responses)
- [x] Update dependency injection container to support the new use case
- [x] Fix tests to use the new use case instead of the old service

### ✅ Step 3: Missing Use Case Layer (COMPLETED)
- [x] Replace thin `AucService` (which was just delegating to repository) with proper use case
- [x] Move business logic from HTTP handler to dedicated use case
- [x] Implement proper aggregation, calculations, and business rules in use case layer
- [x] Update dependency injection to provide use case instead of thin service
- **Result**: Now we have a proper application layer with rich business logic instead of anemic services

### ✅ Step 4: Clean Up Domain Models (COMPLETED)
- [x] Remove infrastructure concerns from domain models (database tags like `db:"symbol"`)
- [x] Create separate DTOs in `position/infra/dto/` for database mapping
- [x] Keep domain models pure without external dependencies (only JSON tags for HTTP serialization)
- [x] Update repository implementations to use DTOs and map to/from domain models
- [x] Create proper mapper to convert between DTOs and domain models
- **Result**: Domain models are now pure and separated from infrastructure concerns

### ✅ Step 5: Improve Repository Interface Design (COMPLETED)
- [x] Make repository interfaces more domain-focused
- [x] Rename `AucRepository` to `PositionRepository` for better domain alignment
- [x] Add more specific methods: `GetPositionsByUserId`, `GetPositionsByCategory`, `GetPositionBySymbol`
- [x] Keep repository interfaces in `position/domain/repository/`
- [x] Update all implementations and dependencies to use new interface
- [x] Create new `SQLXPositionRepository` with improved error messages and domain-focused methods
- [x] Update use cases, services, and tests to use new repository interface
- **Result**: Repository interface is now more domain-focused with specific, well-named methods that clearly express business intent

### ⏳ Step 6: Restructure Package Organization
- [ ] Create proper DDD directory structure:
  ```
  position/
  ├── presentation/http/         # HTTP handlers (✅ Done)
  ├── application/
  │   ├── usecase/              # Use cases (✅ Done)
  │   └── service/              # Application services
  ├── domain/
  │   ├── model/                # Pure domain models (✅ Done)
  │   ├── service/              # Domain services
  │   └── repository/           # Repository interfaces (✅ Done)
  └── infra/
      ├── persistence/          # Repository implementations (✅ Done)
      └── dto/                  # Data transfer objects (✅ Done)
  ```
- [ ] Move existing files to appropriate locations
- [ ] Update all import paths throughout the codebase
- [ ] Ensure proper dependency direction (infra → app → domain)

### Additional Improvements to Consider:
- [ ] Add proper error handling with domain-specific errors
- [ ] Implement input validation in use cases
- [ ] Add logging and monitoring
- [ ] Consider adding domain events for complex workflows
- [ ] Add integration tests for the complete flow