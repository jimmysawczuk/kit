# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Overview

Kit is a modular toolkit for building Go services and programs. The design emphasizes modularity (packages can be used independently), minimal dependencies (only well-known, well-supported libraries), and composability over configuration.

## Build and Test Commands

```bash
# Run all tests with coverage and race detection
go test -cover -v -race ./...

# Run tests for a specific package
go test -v ./web/...

# Run a specific test
go test -v -run TestHandlerFunc ./web/...

# Run go vet for static analysis
go vet ./...

# Build all packages (library, no binary)
go build ./...
```

## Architecture

Kit is organized into independent packages that can be used separately or together:

### Package Structure

- **web**: HTTP service framework built on Chi router (primary package)
- **aws**: AWS integrations (Lambda API Gateway adapter, SSM Parameter Store, DynamoDB wrapper)
- **db/mysql**: MySQL connection management with TLS support
- **cryptorand**: Cryptographically secure random number generation
- **timestamp**: Nullable timestamp type for database operations

### Web Package Architecture

The `web` package is a complete HTTP service framework with the following key components:

#### App

The central orchestrator (`app.go`) that implements `http.Handler`:
- Uses a fluent builder pattern for configuration (all `With*` methods return `*App`)
- Manages router, logger (zerolog), health checkers, and shutdown handlers
- Automatically injects logger into request context

#### Handler Types

Two handler signatures for different use cases:

1. **Handler** - Full control over HTTP response:
   ```go
   Handler func(context.Context, *zerolog.Logger, http.ResponseWriter, *http.Request)
   ```

2. **SyncHandler** - Returns a response object that handles writing:
   ```go
   SyncHandler func(context.Context, *zerolog.Logger, *http.Request) respond.Response
   ```

#### Module Interface

Plugin architecture for composing features:
- Single method: `Route(router.Router)`
- Can optionally implement `HealthChecker` and `Shutdowner` interfaces
- Used with `App.WithModule()` to attach routes, health checks, and shutdown handlers together
- Supports middleware per module

#### Router Abstraction

Wrapper around Chi router providing:
- All HTTP methods (Get, Post, Put, Delete, etc.)
- Both handler and handler func variants (e.g., `Get()` and `Getf()`)
- `Group()` for middleware-scoped route grouping
- `Route()` for path-prefixed route grouping
- `Mount()` for sub-handler mounting

#### Respond Package

Pluggable response handling system (`web/respond/`):

- **Responder** interface allows custom response strategies (defaults to JSONResponder)
- **Response** interface represents an HTTP response
- JSONResponder provides:
  - `Error()` - HTTP error with error message
  - `CodedError()` - Error with optional error code (e.g., "INVALID_TOKEN")
  - `Success()` - Successful response with status code and body

Error response format:
```json
{
  "error": "error message",
  "errorID": "X-Request-Id header value",
  "code": "ERROR_CODE",
  "status": 400,
  "info": {}
}
```

**ErrorInfoer Pattern**: Errors can implement optional `Info() any` method to include structured data in error responses.

#### Request ID

Distributed tracing support (`web/requestid/`):
- Checks incoming `X-Request-Id` header first
- Generates sequential IDs with configurable prefix (hostname + random by default) if not provided
- Integrates with zerolog for automatic logging

#### Health Checks & Shutdown

- `HealthChecker` interface for named health check implementations
- `Shutdowner` interface for named graceful shutdown handlers
- `HealthCheckHandler()` aggregates all health checks into a single endpoint
- `Shutdown()` runs all shutdown handlers in parallel with deadline support

### Middleware

Available middleware in `web/middleware/`:

- **WithLogger**: Attaches zerolog logger to context (apply globally or per-route)
- **RequestID**: Generates/extracts and stores request IDs
- **RequestInfo**: Adds HTTP method, path, URL to log context
- **ProfileRequest**: Logs request duration, response size, status
- **Recoverer**: Catches panics and logs them as errors
- **WithTimeout**: Enforces request timeout limits (15s default; does NOT stop handler execution)
- **WithContext**: Injects custom context modifications
- **WithHeader**: Sets response headers (CORS, security headers, etc.)

Middleware uses standard `func(http.Handler) http.Handler` signature and is compatible with the Chi middleware ecosystem.

### AWS Package

- **apigateway**: Adapts AWS Lambda API Gateway events to standard `http.Handler` interfaces
  - `ProxyHandler()` for API Gateway V1
  - `HTTPHandler()` for API Gateway V2
  - Handles base64 encoding/decoding for binary responses

- **ssm**: AWS Systems Manager Parameter Store integration
  - `GetParametersFromPath()` with automatic pagination
  - `LoadIntoEnv()` loads parameters into environment variables

- **dtable**: DynamoDB table wrapper with convenience methods and automatic logging

## Testing

The codebase uses testify for assertions. Tests should:
- Use `testify/assert` or `testify/require` for assertions
- Test files are co-located with implementation (`*_test.go`)
- Run with race detection enabled (`-race` flag)

## Design Patterns

- **Builder Pattern**: `App.With*()` methods for fluent configuration
- **Strategy Pattern**: Responder and Module interfaces for pluggable behavior
- **Middleware Chain**: Standard Go HTTP middleware pattern
- **Dependency Injection**: Handler functions receive dependencies (context, logger) as parameters
- **Interface Segregation**: Minimal interfaces (Module, HealthChecker, Shutdowner)
