# Marshal - AI Agent Instructions

This Go backend implements **Hexagonal Architecture** with **Domain-Driven Design**. Understanding these patterns is critical for contributing effectively.

## Architecture Overview

### Hexagonal Architecture Layers

The codebase has strict layer separation:

```
internal/{context}/
├── domain/           # Pure business logic, no external dependencies
│   ├── {entity}.go   # Rich domain models with private fields, factory functions
│   ├── repository.go # Repository interface (port) - NEVER implementation
│   └── errors.go     # Domain-specific error instances
├── application/      # Use cases, orchestrates domain logic
│   ├── service.go    # Business workflows, transaction boundaries
│   └── ports/        # Interface definitions for external dependencies
└── adapters/         # Implementations of ports (HTTP, Postgres, Redis, OAuth)
    ├── http/         # HTTP handlers, request/response DTOs, mappers
    ├── postgres/     # Repository implementations with encryption
    └── {other}/      # OAuth providers, token services, caches
```

**Never import domain packages across contexts.** Use Anti-Corruption Layer (ACL) adapters instead (see `internal/auth/adapters/user_service_adapter.go`).

### Bounded Contexts

Three distinct contexts with strict boundaries:

1. **Users** (`internal/users/`) - User profiles (email, name, username, avatar)
2. **Identities** (`internal/identities/`) - Authentication credentials (passwords, OAuth links)
3. **Auth** (`internal/auth/`) - Orchestration layer (tokens, sessions, login/register flows)

**Key Insight**: Auth NEVER imports User or Identity domain packages. It only depends on their service interfaces and uses DTOs (`ports.UserInfo`, `ports.IdentityInfo`). See ACL pattern below.

## Critical Patterns

### 1. Domain Models Are Rich

Never create anemic domain models. Entities encapsulate behavior and validation:

```go
// ✅ CORRECT - Rich domain model
// internal/identities/domain/identity.go
type Identity struct {
    id           int64
    uid          string
    passwordHash string  // Private fields
}

func NewLocalIdentity(userID int64, userUID string, password Password) (*Identity, error) {
    // Factory function with validation
}

func (i *Identity) VerifyPassword(password string, hasher PasswordHasher) error {
    // Business logic lives in domain
}

// ❌ WRONG - Don't expose domain entities in HTTP responses
func (h *Handler) GetIdentity(c echo.Context) error {
    identity, _ := h.service.GetIdentity(ctx, id)
    return c.JSON(200, identity)  // NEVER do this
}
```

Always use request/response DTOs and mappers in HTTP adapters (`ToUserResponse`, `toUserRequest`).

### 2. Transparent Encryption in Repositories

**PII is ALWAYS encrypted** at the repository adapter layer using AES-256-GCM. Domain entities work with plaintext.

**Encryption on Write** (`internal/users/adapters/postgres/repository.go:35-60`):
```go
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
    // Encrypt PII
    emailEncrypted, _ := r.cryptoService.Encrypt(u.Email().String())
    nameEncrypted, _ := r.cryptoService.Encrypt(u.Name())

    // Generate HMAC for searchable lookups (deterministic)
    emailLookupHash, _ := r.cryptoService.HMAC(u.Email().String())

    // Store encrypted data + lookup hash
    query := `INSERT INTO users (
        email_encrypted, email_lookup_hash, name_encrypted, ...
    ) VALUES ($1, $2, $3, ...)`
}
```

**Search via HMAC** (`internal/users/adapters/postgres/repository.go:153`):
```go
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    // Search using deterministic hash (not ciphertext)
    emailLookupHash, _ := r.cryptoService.HMAC(email)
    query := `SELECT * FROM users WHERE email_lookup_hash = $1`
    // Decrypt in toDomain() mapper
}
```

**Decryption on Read** (`internal/users/adapters/postgres/mapper.go:29`):
```go
func (r *UserRepository) toDomain(row *userRow) (*domain.User, error) {
    emailPlaintext, _ := r.cryptoService.Decrypt(row.EmailEncrypted)
    namePlaintext, _ := r.cryptoService.Decrypt(row.NameEncrypted)
    return domain.ReconstructUser(...)  // Domain entity gets plaintext
}
```

**Domain-Specific Keys**: Each context has its own encryption keys (`USER_ENCRYPTION_KEY`, `IDENTITY_ENCRYPTION_KEY`). Never share keys between domains.

### 3. Anti-Corruption Layer (ACL) Between Contexts

When Auth needs to call User or Identity services, it uses ACL adapters that:
- Translate domain entities → DTOs
- Translate errors (prevents error type leakage)
- Define clean port interfaces

**Example** (`internal/auth/adapters/user_service_adapter.go:25-45`):
```go
type UserServiceAdapter struct {
    userService usersapp.Service  // Depends on service interface, NOT domain
}

func (a *UserServiceAdapter) GetUserByEmail(ctx context.Context, email string) (*ports.UserInfo, error) {
    user, err := a.userService.GetByEmail(ctx, email)  // User domain entity
    if err != nil {
        return nil, a.translateError(err)  // Error translation!
    }

    // Convert to Auth's DTO (not User domain type)
    return &ports.UserInfo{
        ID:       user.ID(),
        Email:    user.Email().String(),
        Name:     user.Name(),
        Username: user.Username().String(),
    }, nil
}

func (a *UserServiceAdapter) translateError(err error) error {
    switch {
    case errors.Is(err, usersdomain.ErrUserNotFound):
        return authdomain.ErrUserNotFound  // Auth's error, not User's
    case errors.Is(err, usersdomain.ErrEmailRegistered):
        return authdomain.ErrEmailAlreadyUsed
    default:
        return err
    }
}
```

**Never do this**:
```go
// ❌ WRONG - Direct domain dependency across contexts
import usersdomain "github.com/.../users/domain"

func (s *AuthService) Login(ctx context.Context) {
    user, err := s.userRepo.FindByEmail(ctx, email)  // usersdomain.User
}
```

### 4. Transaction Management via Context

Use `TransactionManager.WithTransaction` for atomic multi-repository operations. Transactions propagate via `context.Context`.

**Service Layer** (`internal/auth/application/local.go:21-45`):
```go
func (s *AuthService) RegisterLocal(ctx context.Context, email, password, name, username string) (*AuthResponse, error) {
    var userInfo *ports.UserInfo
    var identityInfo *ports.IdentityInfo

    // User + Identity must be created atomically
    err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
        userInfo, err = s.userService.CreateUser(txCtx, email, name, username)
        if err != nil {
            return err  // Rollback
        }

        identityInfo, err = s.identityService.CreateLocalIdentity(txCtx, &ports.IdentityCreateLocalRequest{...})
        if err != nil {
            return err  // Rollback
        }

        return nil  // Commit
    })

    // Token generation happens OUTSIDE transaction
    tokenPair, _ := s.tokenService.GenerateTokenPair(ctx, userInfo.ID, userInfo.UID)
}
```

**Repository Layer** (`internal/users/adapters/postgres/repository.go:58`):
```go
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
    // GetQuerier extracts transaction from context if available
    q := database.GetQuerier(ctx, r.db)  // Returns *sqlx.Tx or *sqlx.DB

    // Automatically participates in transaction if ctx comes from WithTransaction
    err := q.QueryRowContext(ctx, query, ...).Scan(&id)
}
```

**Never manually manage transactions** in repositories. Always use `database.GetQuerier(ctx, r.db)` to respect transaction context.

### 5. Error Handling

Use typed errors from `internal/shared/errors`:

```go
// In domain layer
var (
    ErrUserNotFound = errors.NewNotFoundError("user")
    ErrEmailRegistered = errors.NewConflictError("email already registered")
)

// In application service
func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    user, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, domain.ErrUserNotFound  // Translate infrastructure error
        }
        return nil, errors.NewInternalError("failed to query user", err)
    }
    return user, nil
}
```

**Error Middleware** (`internal/adapters/http/middleware/error.go:19`) automatically:
- Maps `AppError.StatusCode()` to HTTP status
- Logs errors with trace context
- Masks internal error details from clients (5xx errors)
- Returns structured JSON: `{"error": "error_code", "message": "...", "details": {...}}`

**Each context defines its own error instances**, even for similar concepts (e.g., `usersdomain.ErrUserNotFound` vs `authdomain.ErrUserNotFound`). Use ACL adapters to translate between them.

### 6. Validation

Use struct tags + custom validators (`internal/shared/validation/custom.go`):

```go
// In request DTO
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email_format"`
    Password string `json:"password" validate:"required,password_complexity"`
    Username string `json:"username" validate:"required,username_format"`
}

// In handler
func (h *AuthHandler) Register(c echo.Context) error {
    var req RegisterRequest
    if err := c.Bind(&req); err != nil {
        return errors.NewBadRequestError("invalid request body")
    }

    if err := h.validator.Validate(&req); err != nil {
        return err  // Returns ValidationError with field details
    }
}
```

**Custom validators** are registered in `validation.NewValidator()`:
- `username_format` - 3-30 chars, alphanumeric + underscore
- `password_complexity` - 8+ chars, uppercase, lowercase, digit
- `email_format` - Valid RFC 5322 email

### 7. HTTP Handler Structure

All HTTP handlers follow this pattern (`internal/users/adapters/http/handler.go`):

1. **Extract user context** (if authenticated): `userID := middleware.GetUserID(c)`
2. **Bind and validate request**: `c.Bind(&req)` + `h.validator.Validate(&req)`
3. **Call application service**: `user, err := h.userService.GetByUID(c.Request().Context(), userID)`
4. **Map to response DTO**: `return c.JSON(200, ToUserResponse(user))`

**Never expose domain entities** in HTTP responses. Always use mappers:
```go
// mapper.go
func ToUserResponse(u *domain.User) *UserResponse {
    return &UserResponse{
        ID:        u.UID(),        // External UUID (not internal int64 ID)
        Email:     u.Email().String(),
        Name:      u.Name(),
        Username:  u.Username().String(),
        CreatedAt: u.CreatedAt(),
        UpdatedAt: u.UpdatedAt(),
    }
}
```

## Bootstrap & Dependency Injection

All wiring happens in `internal/bootstrap/` in strict initialization order (`app.go:31-60`):

1. **Config** → Environment variables
2. **Observability** → OpenTelemetry + Zap logger (BEFORE everything else)
3. **Infrastructure** → PostgreSQL + Redis clients
4. **Crypto Services** → Domain-specific encryption keys (User, Identity)
5. **Repositories** → Inject DB, logger, crypto service
6. **Application Services** → Inject repositories, ports, logger
7. **HTTP Server** → Register routes, middleware

**Key Files**:
- `bootstrap/app.go` - Orchestrates initialization
- `bootstrap/repositories.go` - Repository instantiation
- `bootstrap/services.go` - Service layer wiring (includes ACL adapters)
- `bootstrap/server.go` - Route registration

**Example** (`bootstrap/services.go:36-60`):
```go
func initServices(...) *Services {
    // Infrastructure adapters
    txManager := postgres.NewTransactionManager(db, logger)
    passwordHasher := crypto.NewPasswordHasher()

    // Application services (bounded contexts)
    userSvc := usersapp.NewUserService(repos.UserRepo, logger)
    identitySvc := identitiesapp.NewIdentityService(repos.IdentityRepo, passwordHasher, logger)

    // ACL adapters (Auth → User/Identity communication)
    userServiceAdapter := authadapters.NewUserServiceAdapter(userSvc)
    identityServiceAdapter := authadapters.NewIdentityServiceAdapter(identitySvc)

    // Auth service (orchestration layer)
    authSvc := authapp.NewAuthService(
        txManager,              // Port
        userServiceAdapter,     // Port with ACL
        identityServiceAdapter, // Port with ACL
        tokenService,           // Port
        refreshTokenRepo,       // Domain repository
        logger,                 // Port
        discordProvider,        // OAuth adapter
    )

    return &Services{UserService: userSvc, AuthService: authSvc, ...}
}
```

**Never instantiate services manually** outside bootstrap. Always inject dependencies via constructors.

## Development Workflows

### Adding a New Feature to Existing Context

Example: Add profile picture to User

1. **Update domain entity** (`internal/users/domain/user.go`):
   ```go
   type User struct {
       avatarURL null.String  // Add field
   }

   func (u *User) UpdateAvatar(url string) error {
       // Add business logic + validation
   }
   ```

2. **Add migration** (`make migrate-create`):
   ```sql
   -- migrate:up
   ALTER TABLE users ADD COLUMN avatar_url_encrypted TEXT;
   ALTER TABLE users ADD COLUMN avatar_url_lookup_hash TEXT;
   CREATE INDEX idx_users_avatar_lookup ON users(avatar_url_lookup_hash);
   ```

3. **Update repository** (`internal/users/adapters/postgres/repository.go`):
   - Encrypt/decrypt `avatar_url_encrypted` in `Create` and `toDomain` mapper
   - Add HMAC for `avatar_url_lookup_hash` if searchable

4. **Update application service** (`internal/users/application/service.go`):
   ```go
   func (s *UserService) UpdateAvatar(ctx context.Context, userID int64, url string) error {
       user, _ := s.repo.FindByID(ctx, userID)
       if err := user.UpdateAvatar(url); err != nil {
           return err
       }
       return s.repo.Update(ctx, user)
   }
   ```

5. **Add HTTP endpoint** (`internal/users/adapters/http/handler.go`):
   - Add request DTO with validation tags
   - Add handler method with Swagger docs
   - Map response using `ToUserResponse` (update mapper)

6. **Register route** (`internal/bootstrap/server.go`):
   ```go
   usersGroup.PUT("/me/avatar", handler.UpdateAvatar)
   ```

7. **Run**: `make migrate-up && make dev`

### Adding a New Bounded Context

Follow the structure in `internal/users/`, `internal/identities/`, or `internal/auth/`. See README.md "Adding New Features" section for step-by-step guide.

**Critical steps**:
1. Create domain entity with factory function
2. Define repository interface in `domain/repository.go`
3. Implement Postgres adapter with encryption if needed
4. Create application service
5. Wire in `bootstrap/repositories.go` + `bootstrap/services.go`
6. Add HTTP handlers with DTOs and mappers
7. Register routes in `bootstrap/server.go`

### Database Migrations

```bash
make migrate-create       # Interactive prompt for migration name
# Edit db/migrations/YYYYMMDDHHMMSS_{name}.sql
make migrate-up          # Apply migration
make migrate-status      # Check applied migrations
```

**Migration format** (dbmate):
```sql
-- migrate:up
CREATE TABLE notifications (...);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);

-- migrate:down
DROP INDEX idx_notifications_user_id;
DROP TABLE notifications;
```

**Always test rollback** with `make migrate-down` to ensure idempotency.

## Key Conventions

### UUIDs
- **Internal IDs**: `int64` (database primary keys, never exposed)
- **External IDs**: `uuid.NewV7()` (time-ordered, exposed in APIs)
- Users have both `id` (int64, internal) and `uid` (UUID, external)

### Context Usage
- Always first parameter in functions
- Use `c.Request().Context()` in HTTP handlers
- Pass through transaction boundaries via context

### Null Values
Use `null.String`, `null.Time` from `internal/shared/types` for nullable database columns:
```go
type User struct {
    bio       null.String
    deletedAt null.Time
}
```

### Logging
Inject `ports.Logger` interface, NEVER `*zap.Logger` directly:
```go
func (s *UserService) CreateUser(ctx context.Context, ...) {
    s.logger.Info(ctx, "Creating user",
        ports.String("email", email),
        ports.String("username", username),
    )
}
```

Logger extracts trace/span IDs from context when OpenTelemetry is enabled.

### Testing
- **Unit tests**: Mock ports/interfaces (no DB required)
- **Integration tests**: Use real PostgreSQL/Redis via Docker
- Use `testify/require` for fatal assertions, `testify/assert` for soft checks
- Pass `context.Background()` in test contexts

## Common Pitfalls

1. **DON'T import domain packages across contexts**
   - Use ACL adapters (see `internal/auth/adapters/user_service_adapter.go`)

2. **DON'T expose domain entities in HTTP responses**
   - Always use request/response DTOs and mappers

3. **DON'T handle encryption in domain or service layers**
   - Encryption is ALWAYS transparent in repository adapters

4. **DON'T manage transactions manually in repositories**
   - Use `database.GetQuerier(ctx, r.db)` to respect transaction context

5. **DON'T create anemic domain models**
   - Add business logic methods, validation, and factory functions

6. **DON'T share encryption keys between domains**
   - Each context gets its own keys (`USER_ENCRYPTION_KEY`, `IDENTITY_ENCRYPTION_KEY`)

7. **DON'T use concrete logger types**
   - Always inject `ports.Logger` interface

8. **DON'T initialize OpenTelemetry after logger**
   - OTel MUST be initialized first for trace context propagation

## Essential Commands

```bash
# Development
make dev                 # Hot reload with Air
make test                # Run tests with race detector
make migrate-up          # Apply database migrations

# Docker infrastructure
make docker-up           # Start PostgreSQL + Redis
make docker-logs         # Follow container logs

# Documentation
make gen-docs            # Generate OpenAPI/Swagger docs
```

**First-time setup**:
```bash
cp .env.example .env
openssl rand -hex 32     # Run 5 times for 5 different keys
# Edit .env with generated keys
make docker-up
make migrate-up
make dev
```

## Directory Reference

- **`internal/{context}/domain/`** - Entities, value objects, repository interfaces
- **`internal/{context}/application/`** - Services, use cases, port interfaces
- **`internal/{context}/adapters/`** - HTTP handlers, Postgres repos, OAuth, tokens
- **`internal/shared/`** - Cross-cutting: config, crypto, database, errors, validation
- **`internal/adapters/http/`** - Global HTTP server setup and middleware
- **`internal/bootstrap/`** - Dependency injection and app initialization
- **`db/migrations/`** - SQL migrations (dbmate format)
- **`cmd/server/`** - Application entry point

## Key Dependencies

- **Web Framework**: Echo v4
- **Database**: PostgreSQL 18 + sqlx
- **Cache**: Redis 7 + go-redis
- **Tokens**: PASETO v4 (authenticated encryption)
- **Logging**: Zap (via port interface)
- **Observability**: OpenTelemetry (traces + logs)
- **Validation**: go-playground/validator
- **Encryption**: AES-256-GCM + HMAC-SHA256
- **Password Hashing**: Argon2id

## References

- **Architecture Overview**: This document
- **API Documentation**: http://localhost:8080/docs (after `make dev`)
- **Full Developer Guide**: README.md
- **Environment Config**: .env.example
