# Architecture Decision Records (ADR)

## Overview

This document captures key architectural decisions made during the microservices migration.

---

## ADR-001: Microservices Communication Protocol

**Date**: 2025-10-27  
**Status**: ✅ ACCEPTED  
**Context**: Deciding how microservices should expose their APIs

### Decision

**Microservices will ONLY expose gRPC APIs. HTTP REST APIs will NOT be implemented in microservices.**

### Rationale

1. **API Gateway Pattern**: The API Gateway (`hub-api-gateway`) acts as the single entry point for all external traffic
2. **Protocol Translation**: API Gateway handles HTTP → gRPC translation
3. **Separation of Concerns**: 
   - Frontend concerns (HTTP, WebSocket, JSON) → API Gateway
   - Business logic concerns (gRPC, Protobuf) → Microservices
4. **Performance**: gRPC is more efficient than HTTP REST for inter-service communication
5. **Type Safety**: Protobuf provides strong typing and versioning

### Architecture

```
┌──────────┐         ┌─────────────┐         ┌──────────────────┐
│ Frontend │◄───────►│ API Gateway │◄───────►│ Microservice     │
│          │  HTTP   │             │  gRPC   │                  │
│ (React)  │  JSON   │ (Go)        │ Protobuf│ (Go)             │
└──────────┘         └─────────────┘         └──────────────────┘
```

### Consequences

**Positive:**
- ✅ Microservices are simpler (no HTTP routing, middleware, etc.)
- ✅ Centralized security, rate limiting, and observability in API Gateway
- ✅ Easy to change frontend protocols without touching microservices
- ✅ Better performance for inter-service communication
- ✅ Consistent API contracts via Protobuf

**Negative:**
- ❌ API Gateway becomes a single point of failure (mitigated by load balancing)
- ❌ Additional hop adds ~5-10ms latency (acceptable for our use case)

### Implementation

- **Market Data Service**: ✅ Implements gRPC only (Step 2.3)
- **Future Microservices**: Will follow the same pattern
- **API Gateway**: Implements HTTP → gRPC translation

---

## ADR-002: WebSocket Communication Architecture

**Date**: 2025-10-27  
**Status**: ✅ ACCEPTED  
**Context**: Deciding how to handle real-time WebSocket communication in microservices

### Decision

**WebSocket servers will live in the API Gateway, NOT in microservices. Microservices will expose gRPC bidirectional streaming.**

### Rationale

1. **WebSocket Proxy Pattern**: API Gateway acts as a WebSocket-to-gRPC proxy
2. **Centralized Connection Management**: API Gateway handles WebSocket complexity
3. **Protocol Translation**: 
   - Frontend ↔ API Gateway: WebSocket (JSON)
   - API Gateway ↔ Microservice: gRPC Streaming (Protobuf)
   - Microservice ↔ Redis: Pub/Sub
4. **Security**: Centralized authentication, authorization, and rate limiting
5. **Scalability**: Easy to scale API Gateway and microservices independently

### Architecture

```
┌──────────┐         ┌─────────────┐         ┌──────────────────┐
│ Frontend │◄───────►│ API Gateway │◄───────►│ Market Data      │
│          │ WebSocket│             │ gRPC    │ Service          │
│ (React)  │  JSON   │ (Go)        │ Stream  │ (Go)             │
└──────────┘         └─────────────┘         └──────────────────┘
                            │                         │
                            │                         ▼
                            │                  ┌─────────────┐
                            │                  │ Redis       │
                            │                  │ Pub/Sub     │
                            │                  └─────────────┘
                            ▼
                     ┌─────────────┐
                     │ Auth        │
                     │ Rate Limit  │
                     │ Metrics     │
                     └─────────────┘
```

### Consequences

**Positive:**
- ✅ Microservices don't need to handle WebSocket complexity
- ✅ Centralized security and observability
- ✅ Easy to implement load balancing
- ✅ Protocol independence (can change WebSocket implementation without touching microservices)
- ✅ Consistent authentication and rate limiting

**Negative:**
- ❌ API Gateway must handle WebSocket connection state
- ❌ Additional complexity in API Gateway (mitigated by clear separation of concerns)

### Implementation

- **API Gateway**: Implements WebSocket server + gRPC client (Step 2.5)
- **Market Data Service**: Implements gRPC bidirectional streaming (Step 2.5)
- **Proto Contract**: Add `StreamQuotes` RPC to `market_data_service.proto`

### Alternative Considered

**Direct WebSocket from Microservice:**
```
Frontend → API Gateway (HTTP Proxy) → Microservice (WebSocket Server)
```

**Why Rejected:**
- ❌ Breaks API Gateway pattern (no centralized auth/rate limiting)
- ❌ Requires exposing microservice ports
- ❌ Harder to implement load balancing
- ❌ Loses API Gateway benefits (metrics, tracing, auth)

---

## ADR-003: Shared Infrastructure Abstractions

**Date**: 2025-10-27  
**Status**: ✅ ACCEPTED  
**Context**: Avoiding direct dependencies on infrastructure tools (Redis, RabbitMQ, PostgreSQL, Elasticsearch)

### Decision

**Create a single `hub-investments-common` repository containing all infrastructure abstractions (cache, messaging, database, search).**

### Rationale

1. **Dependency Inversion Principle**: Services depend on interfaces, not implementations
2. **Single Repository**: Easier to maintain than multiple separate repos
3. **Atomic Updates**: Update all infrastructure together with single version
4. **Consistency**: Standardized patterns across all services
5. **Testability**: Mock implementations for unit tests
6. **Flexibility**: Easy to swap implementations (Redis → DragonflyDB, RabbitMQ → Kafka)

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              hub-investments-common                         │
│                                                             │
│  ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐  │
│  │ cache/  │  │messaging/│  │database/ │  │  search/   │  │
│  │         │  │          │  │          │  │            │  │
│  │ Redis   │  │ RabbitMQ │  │PostgreSQL│  │Elasticsearch│ │
│  │ Memory  │  │ Kafka    │  │ Mock     │  │  Mock      │  │
│  │ Mock    │  │ Mock     │  │          │  │            │  │
│  └─────────┘  └──────────┘  └──────────┘  └────────────┘  │
│                                                             │
│  ┌─────────┐  ┌──────────┐  ┌──────────┐                  │
│  │ logger/ │  │ config/  │  │ errors/  │                  │
│  └─────────┘  └──────────┘  └──────────┘                  │
│                                                             │
│  ┌─────────────────────────────────┐                       │
│  │ observability/                  │                       │
│  │   - metrics/ (Prometheus)       │                       │
│  │   - tracing/ (OpenTelemetry)    │                       │
│  └─────────────────────────────────┘                       │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ import
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Microservices                            │
│                                                             │
│  hub-market-data-service                                    │
│  hub-user-service                                           │
│  hub-order-service                                          │
│  hub-portfolio-service                                      │
│  HubInvestmentsServer (monolith)                            │
└─────────────────────────────────────────────────────────────┘
```

### Consequences

**Positive:**
- ✅ Single repository to maintain (vs 4+ separate repos)
- ✅ Loose coupling (services depend on interfaces)
- ✅ Easy testing (mock implementations)
- ✅ Easy swapping (change Redis → DragonflyDB without touching services)
- ✅ Consistent patterns across all services
- ✅ Reusability (write once, use everywhere)
- ✅ Single semantic version for all infrastructure
- ✅ Bug fixes apply to all services via `go get -u`
- ✅ Atomic updates (cache + messaging + database together)
- ✅ Simplified dependency management (one `go.mod` entry)

**Negative:**
- ❌ Requires upfront design and implementation (3 weeks)
- ❌ Breaking changes affect all services (mitigated by semantic versioning)

### Implementation Plan

**Phase 1: Create Library** (3 weeks)
- Define interfaces for cache, messaging, database, search
- Implement adapters (Redis, RabbitMQ, PostgreSQL, Elasticsearch)
- Create mock implementations
- Write comprehensive tests and documentation

**Phase 2: Migrate Services** (2 weeks)
- Migrate `hub-market-data-service`
- Migrate `HubInvestmentsServer` (monolith)
- Migrate future microservices

**Phase 3: Advanced Features** (3 weeks)
- Add distributed locking, cache warming, query builders
- Add Kafka adapter, message batching, saga patterns
- Add Prometheus metrics, OpenTelemetry tracing

### Priority

**CRITICAL** - Must be done before creating 3rd microservice

### Alternative Considered

**Multiple Separate Repositories:**
- `hub-cache-client`
- `hub-messaging-client`
- `hub-database-client`
- `hub-search-client`

**Why Rejected:**
- ❌ More complex to maintain (4+ repos vs 1)
- ❌ Harder to keep versions in sync
- ❌ More complex dependency management
- ❌ Can't do atomic updates across infrastructure

---

## ADR-004: Strangler Fig Pattern for Microservices Migration

**Date**: 2025-10-27  
**Status**: ✅ ACCEPTED  
**Context**: Migrating from monolith to microservices without downtime

### Decision

**Use the Strangler Fig Pattern to gradually migrate functionality from the monolith to microservices.**

### Rationale

1. **Zero Downtime**: Gradual migration without service interruption
2. **Risk Mitigation**: Easy rollback if issues arise
3. **Incremental Validation**: Test each microservice before full cutover
4. **Team Learning**: Build confidence and experience incrementally

### Pattern

```
Phase 1: Monolith Only
┌─────────────────────────────────────┐
│         Monolith                    │
│  ┌──────────┐  ┌──────────────┐    │
│  │  Users   │  │ Market Data  │    │
│  └──────────┘  └──────────────┘    │
└─────────────────────────────────────┘

Phase 2: Parallel Run (Strangler Fig)
┌─────────────────────────────────────┐
│         Monolith                    │
│  ┌──────────┐  ┌──────────────┐    │
│  │  Users   │  │ Market Data  │    │ ← Still exists
│  └──────────┘  └──────────────┘    │
└─────────────────────────────────────┘
                    │
                    │ (API Gateway routes to microservice)
                    ▼
         ┌──────────────────┐
         │ Market Data      │
         │ Microservice     │ ← New microservice
         └──────────────────┘

Phase 3: Monolith Code Removal (Manual)
┌─────────────────────────────────────┐
│         Monolith                    │
│  ┌──────────┐                       │
│  │  Users   │  (Market Data removed)│
│  └──────────┘                       │
└─────────────────────────────────────┘
                    │
                    ▼
         ┌──────────────────┐
         │ Market Data      │
         │ Microservice     │
         └──────────────────┘
```

### Consequences

**Positive:**
- ✅ Zero downtime during migration
- ✅ Easy rollback (just route traffic back to monolith)
- ✅ Incremental validation
- ✅ Team learning and confidence building

**Negative:**
- ❌ Temporary code duplication (monolith + microservice)
- ❌ Requires manual cleanup after validation

### Implementation

**⚠️ IMPORTANT NOTE**: The market data code in the monolith (`internal/market_data/`, `internal/realtime_quotes/`) will **remain in place** after the microservice is created. The monolith will continue to have this code until manual removal is performed after full validation. This allows for easy rollback and gradual traffic migration.

---

## Summary

| ADR | Decision | Status | Impact |
|-----|----------|--------|--------|
| ADR-001 | Microservices expose gRPC only | ✅ ACCEPTED | High |
| ADR-002 | WebSocket in API Gateway, not microservices | ✅ ACCEPTED | High |
| ADR-003 | Single `hub-investments-common` repo | ✅ ACCEPTED | Critical |
| ADR-004 | Strangler Fig Pattern for migration | ✅ ACCEPTED | High |

**Next Steps:**
1. ✅ Complete Market Data Service gRPC implementation (Step 2.3)
2. [ ] Implement gRPC streaming for real-time quotes (Step 2.5)
3. [ ] Create `hub-investments-common` repository (Phase 1)
4. [ ] Migrate Market Data Service to use `hub-investments-common` (Phase 2)
5. [ ] Integrate Market Data Service with API Gateway (Step 3)

