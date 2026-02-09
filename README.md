# Just Another Go Backend Boilerplate

[![Go Version](https://img.shields.io/badge/Go-1.25.0-00ADD8?logo=go)](https://go.dev/)

Well it's just another Go backend boilerplate, but this one is designed with **Hexagonal Architecture** and **Domain-Driven Design** (DDD) principles in mind. It features a complete authentication system, structured logging, OpenTelemetry tracing.

Is it over-egineered for a simple app? It is. It's meant to be a learning resource and a solid foundation for building more complex applications.
Feel free to use and adapt it as you see fit!

For detailed architecture documentation, see [.github/copilot-instructions.md](.github/copilot-instructions.md).

---

## Quick Start

### Prerequisites

- **Go**: 1.25.0 or higher
- **Docker** & **Docker Compose**: For PostgreSQL and Redis (optional if you have your own instances)
- **OpenSSL**: For generating encryption keys (usually pre-installed)

### Setup

```bash
# 1. Copy environment template
cp .env.example .env

# 2. Generate encryption keys (run 4 times for different keys)
openssl rand -hex 32

# 3. Edit .env and replace placeholder keys:
#    PASETO_SYMMETRIC_KEY, USER_ENCRYPTION_KEY, USER_HMAC_SECRET,
#    IDENTITY_ENCRYPTION_KEY, IDENTITY_HMAC_SECRET

# 4. Start infrastructure (PostgreSQL + Redis)
make docker-up

# 5. Run migrations and start dev server
make migrate-up && make dev
```

The server starts on `http://localhost:8080` with API docs at `http://localhost:8080/docs`.

**Test it works:**
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy"}
```

---

## Development Workflow

### Essential Commands

| Command | Description |
|---------|-------------|
| **Development** | |
| `make dev` | Start server with hot reload (Air) |
| `make run` | Run server without hot reload |
| `make build` | Build binary to `bin/server` |
| **Testing** | |
| `make test` | Run tests with race detector |
| `make test-coverage` | Generate HTML coverage report |
| **Database** | |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-create` | Create new migration (interactive) |
| `make migrate-status` | Show migration status |
| **Documentation** | |
| `make gen-docs` | Generate OpenAPI/Swagger docs |
| **Docker** | |
| `make docker-up` | Start PostgreSQL + Redis |
| `make docker-down` | Stop containers |
| `make docker-logs` | Follow container logs |
| **Utilities** | |
| `make deps` | Download and tidy dependencies |
| `make clean` | Remove build artifacts |

### Development Cycle

1. Start infrastructure: `make docker-up`
2. Run dev server: `make dev` (auto-reloads on file changes)
3. Make code changes
4. Run tests: `make test`
5. Update docs if needed: `make gen-docs`

### Hot Reload

Hot reload is configured via Air (see `configs/.air.toml`). It watches:
- All `*.go` files
- `.env` file changes
- Excludes: `tmp/`, `vendor/`, `docs/`

---

## API Endpoints

**Base URL:** `/v1`

### System Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | None | Health check |
| GET | `/docs` | None | Interactive API documentation (Scalar UI) |

### Authentication Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | None | Register with email/password |
| POST | `/auth/login` | None | Login with credentials |
| POST | `/auth/refresh` | None | Exchange refresh token for new pair |
| POST | `/auth/logout` | None | Revoke a refresh token (single device) |
| POST | `/auth/logout-all` | Bearer | Revoke all refresh tokens (all devices) |
| GET | `/auth/discord` | None | Redirect to Discord OAuth |
| GET | `/auth/callback/discord` | None | Discord OAuth callback |

### User Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/users/me` | Bearer | Get current user profile |

### Quick Examples

**Register:**
```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "name": "John Doe",
    "username": "johndoe"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

**Get Current User:**
```bash
curl http://localhost:8080/v1/users/me \
  -H "Authorization: Bearer <your-access-token>"
```

**Refresh Token:**
```bash
curl -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<your-refresh-token>"}'
```

---

## Configuration Reference

Edit `.env` to configure the application. Required settings are marked with ⚠️.

### Server

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SERVER_HOST` | No | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | No | `8080` | HTTP port |
| `SERVER_ENV` | No | `development` | Environment (`development`, `production`) |
| `LOG_LEVEL` | No | `debug` | Log level: `debug`, `info`, `warn`, `error` |

### Database

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DB_HOST` | No | `localhost` | PostgreSQL host |
| `DB_PORT` | No | `5432` | PostgreSQL port |
| `DB_USER` | No | `postgres` | Database user |
| `DB_PASSWORD` | No | `postgres` | Database password |
| `DB_NAME` | No | `marshal` | Database name |
| `DB_SSLMODE` | No | `disable` | SSL mode (`disable` for dev, `require` for prod) |
| `DB_MAX_OPEN_CONNS` | No | `25` | Maximum open connections |
| `DB_MAX_IDLE_CONNS` | No | `5` | Maximum idle connections |

### Redis

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REDIS_HOST` | No | `localhost` | Redis server host |
| `REDIS_PORT` | No | `6379` | Redis server port |
| `REDIS_PASSWORD` | No | (empty) | Redis password |
| `REDIS_DB` | No | `0` | Redis database number (0-15) |

### Security (⚠️ Required for Auth)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PASETO_SYMMETRIC_KEY` | ⚠️ Yes | - | Token signing key (32-byte hex, 64 chars). Generate: `openssl rand -hex 32` |
| `PASETO_ACCESS_TOKEN_DURATION` | No | `15m` | Access token lifetime |
| `PASETO_REFRESH_TOKEN_DURATION` | No | `7d` | Refresh token lifetime |
| `USER_ENCRYPTION_KEY` | ⚠️ Yes | - | User domain AES key (32-byte hex, 64 chars) |
| `USER_HMAC_SECRET` | ⚠️ Yes | - | User domain HMAC key (32-byte hex, 64 chars) |
| `IDENTITY_ENCRYPTION_KEY` | ⚠️ Yes | - | Identity domain AES key (32-byte hex, 64 chars) |
| `IDENTITY_HMAC_SECRET` | ⚠️ Yes | - | Identity domain HMAC key (32-byte hex, 64 chars) |

**Security Note**: Use different keys for each domain and environment. Never commit real keys to git.

### OAuth (Discord)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DISCORD_CLIENT_ID` | For OAuth | - | Discord application client ID |
| `DISCORD_CLIENT_SECRET` | For OAuth | - | Discord application client secret |
| `DISCORD_REDIRECT_URL` | For OAuth | - | OAuth callback URL (e.g., `http://localhost:8080/v1/auth/callback/discord`) |

Get credentials from [Discord Developer Portal](https://discord.com/developers/applications).

### CORS

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | No | `http://localhost:3000` | Comma-separated allowed origins |

### OpenTelemetry (Optional)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OTEL_ENABLED` | No | `false` | Enable/disable OpenTelemetry tracing |
| `OTEL_SERVICE_NAME` | No | `marshal` | Service name in traces |
| `OTEL_EXPORTER_TYPE` | No | `console` | Exporter: `otlp` (prod) or `console` (dev) |
| `OTEL_OTLP_ENDPOINT` | No | `localhost:4317` | OTLP gRPC endpoint (Jaeger, Tempo, etc.) |
| `OTEL_SAMPLE_RATE` | No | `1.0` | Trace sampling rate (0.0 to 1.0) |

---

## Project Structure

```
marshal/
├── cmd/server/             # Application entry point
│   └── main.go            # Server startup + Swagger config
├── internal/              # Private application code
│   ├── auth/              # Authentication bounded context
│   │   ├── domain/        # Entities, value objects (RefreshToken)
│   │   ├── application/   # Auth service, use cases (register, login, OAuth)
│   │   └── adapters/      # HTTP handlers, PASETO tokens, OAuth, Postgres, Redis
│   ├── users/             # User management bounded context
│   │   ├── domain/        # User entity, Email/Username value objects
│   │   ├── application/   # User service
│   │   └── adapters/      # HTTP handlers, Postgres repository
│   ├── identities/        # Identity/credentials bounded context
│   │   ├── domain/        # Identity entity, Password value object
│   │   ├── application/   # Identity service
│   │   └── adapters/      # Postgres repository
│   ├── shared/            # Cross-cutting concerns
│   │   ├── config/        # Environment configuration
│   │   ├── crypto/        # AES-256-GCM encryption, HMAC, Argon2
│   │   ├── database/      # PostgreSQL + Redis clients
│   │   ├── errors/        # Shared error types (AppError)
│   │   └── adapters/      # Logger, transaction manager, telemetry
│   ├── adapters/http/     # Global HTTP infrastructure
│   │   ├── server.go      # Echo server setup
│   │   └── middleware/    # Auth, error handling, logging, telemetry
│   └── bootstrap/         # Dependency injection
│       ├── app.go         # App initialization orchestrator
│       ├── repositories.go # Repository wiring
│       ├── services.go    # Service wiring
│       └── server.go      # Route registration
├── db/
│   ├── migrations/        # SQL migration files (dbmate format)
│   └── schema.sql         # Auto-generated schema
├── docs/                  # Generated OpenAPI documentation
├── configs/
│   └── .air.toml          # Hot reload configuration
├── Makefile               # Development commands
├── docker-compose.yml     # PostgreSQL + Redis services
└── .env.example           # Environment template
```

---

## Adding New Features

### Adding a New Domain Module

Follow these steps to add a new bounded context (e.g., `notifications`):

1. **Create domain entity**: `internal/notifications/domain/notification.go`
   - Define entity struct with factory function (`NewNotification()`)
   - Add business logic methods
   - Use private fields with public getters

2. **Define repository interface**: `internal/notifications/domain/repository.go`
   - Define methods like `Create()`, `FindByID()`, `FindByUserID()`
   - Interface only - no implementation

3. **Implement Postgres adapter**: `internal/notifications/adapters/postgres/repository.go`
   - Implement the repository interface
   - Handle encryption/decryption if needed
   - Use `context.Context` for all methods

4. **Create application service**: `internal/notifications/application/service.go`
   - Orchestrate domain logic
   - Depend on repository interface (not concrete implementation)
   - Add business validation

5. **Wire in bootstrap**: `internal/bootstrap/repositories.go` + `services.go`
   - Instantiate repository and service
   - Inject dependencies

6. **Add HTTP handlers**: `internal/notifications/adapters/http/handler.go`
   - Create request/response DTOs
   - Add mappers (entity ↔ DTO)
   - Never expose domain entities directly

7. **Register routes**: `internal/bootstrap/server.go`
   - Add route group
   - Apply middleware (auth, etc.)

**Reference**: See `internal/auth/`, `internal/users/`, or `internal/identities/` for complete examples.

### Key Patterns

- **Repository Pattern**: Domain defines interface, adapter implements
- **DTO Pattern**: Use request/response DTOs, never expose domain entities in HTTP
- **Error Handling**: Use `shared/errors` types (NotFoundError, BadRequestError, etc.)
- **Transaction Boundaries**: Use `TransactionManager` for atomic multi-repo operations
- **Value Objects**: Encapsulate validation (Email, Password, Username)
- **Anti-Corruption Layer**: Use adapters between bounded contexts (see `auth/adapters/user_service_adapter.go`)

### Code Conventions

- **UUIDs**: Use `uuid.NewV7()` (time-ordered) for all entity IDs
- **Context**: Always first parameter in functions
- **Null Values**: Use `null.String`, `null.Time` from `shared/types`
- **Logging**: Use injected `ports.Logger` interface, not `*zap.Logger` directly
- **Encryption**: Transparent in repositories (AES-256-GCM + HMAC for lookups)

---

## Database Management

### Creating Migrations

```bash
# Create new migration (interactive prompt)
make migrate-create

# Example: Enter migration name: add_notifications_table
```

Generates: `db/migrations/YYYYMMDDHHMMSS_add_notifications_table.sql`

### Migration Format

Migrations use dbmate format:

```sql
-- migrate:up
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);

-- migrate:down
DROP INDEX idx_notifications_user_id;
DROP TABLE notifications;
```

### Migration Commands

```bash
make migrate-up          # Apply all pending migrations
make migrate-down        # Rollback last migration
make migrate-status      # Show applied/pending migrations
```

### Schema

- **Location**: `db/migrations/`
- **Auto-generated schema**: `db/schema.sql` (updated by dbmate)
- **Existing tables**: `users`, `identities`, `refresh_tokens`

---

## Testing

### Running Tests

```bash
# Run all tests with race detector
make test

# Generate HTML coverage report
make test-coverage
# Opens coverage report in browser
```

### Testing Patterns

- **Integration tests**: Use real PostgreSQL/Redis via Docker
- **Setup helper**: Create test clients that connect to localhost services
- **Assertions**: Use `testify/require` for fatal errors, `testify/assert` for soft checks
- **Context**: Pass `context.Background()` or `context.TODO()` in tests

**Example test structure:**
```go
func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    repo := postgres.NewUserRepository(db, cryptoService)
    user := domain.NewUser("test@example.com", "Test User")

    err := repo.Create(context.Background(), user)
    require.NoError(t, err)

    found, err := repo.FindByID(context.Background(), user.ID())
    require.NoError(t, err)
    assert.Equal(t, user.Email(), found.Email())
}
```

---

## Troubleshooting

### "Invalid PASETO key" error

**Cause**: `PASETO_SYMMETRIC_KEY` is not 64 hex characters (32 bytes).

**Fix**: Generate a new key and update `.env`:
```bash
openssl rand -hex 32
# Copy the output to PASETO_SYMMETRIC_KEY in .env
```

### Database connection errors

**Symptoms**: `connection refused`, `database does not exist`

**Fix**:
```bash
# 1. Ensure PostgreSQL is running
make docker-up

# 2. Check database exists
docker exec -it marshal-postgres psql -U postgres -c '\l'

# 3. Run migrations
make migrate-up

# 4. Check connection settings in .env
# Ensure DB_HOST=localhost, DB_PORT=5432, DB_NAME=marshal
```

### Redis cache not working

**Symptoms**: Tokens still work but logs show Redis errors

**Fix**:
```bash
# 1. Verify Redis is running
docker ps | grep redis

# 2. Test connection
docker exec -it marshal-redis redis-cli ping
# Expected: PONG

# 3. Check .env settings
# REDIS_HOST=localhost, REDIS_PORT=6379
```

**Note**: Application gracefully degrades if Redis is unavailable (uses Postgres only).

### Migration fails

**Symptoms**: `migration failed`, `duplicate key`, `column already exists`

**Common causes**:
- Migration already applied (check `make migrate-status`)
- Database schema inconsistent

**Fix**:
```bash
# 1. Check migration status
make migrate-status

# 2. If migration is partially applied, rollback
make migrate-down

# 3. Fix the migration file in db/migrations/
# 4. Re-apply
make migrate-up
```

### Trace IDs not appearing in logs

**Symptoms**: Logs missing `trace_id` and `span_id` fields

**Fix**:
1. Ensure `OTEL_ENABLED=true` in `.env`
2. Verify OpenTelemetry is initialized **before** logger (see `internal/bootstrap/app.go`)
3. Check `OTEL_EXPORTER_TYPE` is set (default: `console`)
4. Restart server: `make dev`

---

## Additional Resources

- **Detailed Architecture**: [.github/copilot-instructions.md](.github/copilot-instructions.md) - Comprehensive architecture guide, patterns, and conventions
- **API Documentation**: http://localhost:8080/docs - Interactive Scalar UI (after `make dev`)
- **OpenAPI Spec**: `docs/swagger.json` - Generated API specification
- **Make Commands**: Run `make help` to see all available commands

### Key Dependencies

- **HTTP Framework**: [Echo v4](https://echo.labstack.com/)
- **Database**: PostgreSQL 18 + [sqlx](https://github.com/jmoiron/sqlx)
- **Cache**: Redis 7 + [go-redis](https://github.com/redis/go-redis)
- **Tokens**: [PASETO v4](https://github.com/aidanwoods/go-paseto)
- **Logging**: [Zap](https://github.com/uber-go/zap) via port interface
- **Observability**: [OpenTelemetry](https://opentelemetry.io/)
- **Migrations**: [dbmate](https://github.com/amacneil/dbmate)
- **Hot Reload**: [Air](https://github.com/air-verse/air)