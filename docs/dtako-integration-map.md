# dtako_mod Integration Mapping

**Date**: 2025-09-12  
**Target**: ryohi_sub_cal2 Router Module  
**Module**: github.com/yhonda-ohishi/dtako_mod  

## Integration Architecture

```
┌─────────────────────────────────────────────────┐
│           Client Requests                        │
└─────────────┬───────────────────────────────────┘
              │
┌─────────────▼───────────────────────────────────┐
│         Router Module (gorilla/mux)              │
├──────────────────────────────────────────────────┤
│  Middleware Chain:                               │
│  ├─ Authentication                               │
│  ├─ Rate Limiting                                │
│  ├─ Logging                                      │
│  ├─ Metrics                                      │
│  └─ CORS                                         │
├──────────────────────────────────────────────────┤
│  Routes:                                         │
│  ├─ /health                                      │
│  ├─ /metrics                                     │
│  ├─ /admin/*                                     │
│  └─ /dtako/* ──────► dtako_mod Routes           │
└──────────────────────────────────────────────────┘
```

## Integration Points

### 1. Route Integration
**Current Setup**: gorilla/mux router  
**dtako_mod Requirement**: chi router  
**Solution**: Create chi-to-mux adapter

```go
// src/lib/adapters/chi_mux_adapter.go
func AdaptChiToMux(muxRouter *mux.Router, chiRoutes func(chi.Router)) {
    // Create chi router instance
    // Register dtako routes
    // Bridge to mux router
}
```

**Integration Location**: `src/server.go`
```go
// Add dtako routes to main router
dtako_mod.RegisterRoutes(adaptedRouter)
```

### 2. Handler Integration
**Current Middleware Chain**:
```
Request → Auth → RateLimit → Logging → Handler → Response
```

**dtako_mod Handler Wrapping**:
```go
// Wrap dtako handlers with existing middleware
wrappedHandler := middleware.Chain(
    middleware.Auth,
    middleware.RateLimit,
    middleware.Logging,
    dtakoHandler,
)
```

### 3. Service Layer Integration

**Current Services**:
- router/router.go
- proxy/proxy.go  
- health/health.go
- loadbalancer/loadbalancer.go
- circuit/circuit.go

**New Service Integration**:
```go
// src/services/dtako/dtako_service.go
type DtakoService struct {
    rowsService   *dtako_mod.RowsService
    eventsService *dtako_mod.EventsService
    ferryService  *dtako_mod.FerryService
}
```

### 4. Model Compatibility

**Existing Models**:
- RouteConfig
- BackendService
- HealthCheckConfig
- CircuitBreakerConfig
- RateLimitConfig
- AuthConfig
- Metrics

**dtako_mod Models Integration**:
```go
// src/models/dtako_models.go
// Re-export dtako_mod models
type (
    DtakoRow   = models.DtakoRow
    DtakoEvent = models.DtakoEvent
    DtakoFerry = models.DtakoFerry
)
```

### 5. Database Integration

**Current**: Stateless (config from YAML/JSON)  
**dtako_mod Needs**: Database connection  

**Solution**: Add database configuration
```go
// src/lib/config/database.go
type DatabaseConfig struct {
    Host     string
    Port     int
    Database string
    User     string
    Password string
}
```

### 6. Middleware Enhancement Opportunities

While dtako_mod doesn't provide traditional middleware, we can enhance:

**Request Validation**:
```go
// Use dtako_mod models for validation
func ValidateDtakoRequest(next http.Handler) http.Handler {
    // Validate using ImportRequest model
}
```

**Error Handling**:
```go
// Standardize error responses
func DtakoErrorHandler(next http.Handler) http.Handler {
    // Handle dtako-specific errors
}
```

### 7. Configuration Management

**Current Config** (viper):
```yaml
router:
  port: 8080
  routes: []
```

**Extended Config** for dtako:
```yaml
router:
  port: 8080
  routes: []
dtako:
  enabled: true
  database:
    host: localhost
    port: 5432
    name: dtako_db
  import:
    batch_size: 1000
    timeout: 30s
```

## Implementation Strategy

### Phase 1: Adapter Layer
1. Create chi-to-mux router adapter
2. Test adapter with simple routes
3. Validate middleware chain compatibility

### Phase 2: Route Registration
1. Register dtako routes under `/dtako` prefix
2. Apply authentication to import endpoints
3. Add rate limiting to list endpoints

### Phase 3: Service Integration
1. Initialize dtako services
2. Configure database connection
3. Add service health checks

### Phase 4: Monitoring
1. Add dtako-specific metrics
2. Include in health check endpoint
3. Add structured logging for dtako operations

## Compatibility Matrix

| Component | Router Module | dtako_mod | Compatible | Notes |
|-----------|--------------|-----------|------------|-------|
| Router | gorilla/mux | chi/v5 | ⚠️ | Needs adapter |
| Go Version | 1.23.0 | 1.21 | ✅ | Forward compatible |
| HTTP | net/http | net/http | ✅ | Same interface |
| Models | Custom | Custom | ✅ | Can coexist |
| Database | None | Required | ⚠️ | Need to add DB |
| Config | viper | Unknown | ✅ | Can extend |

## Migration Path

1. **No Breaking Changes**: Existing routes continue to work
2. **Additive Only**: dtako routes added under new prefix
3. **Gradual Adoption**: Can enable/disable via config
4. **Rollback Safe**: Can remove dtako routes without affecting core

## Testing Strategy

1. **Unit Tests**: Test adapter and wrapper functions
2. **Integration Tests**: Test dtako routes with middleware
3. **Contract Tests**: Verify dtako API contracts
4. **Performance Tests**: Ensure no degradation

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Router incompatibility | Medium | Use adapter pattern |
| Database requirement | High | Make optional via config |
| Performance impact | Low | Monitor metrics |
| Maintenance burden | Medium | Clear separation of concerns |

---
*Mapping completed: 2025-09-12*