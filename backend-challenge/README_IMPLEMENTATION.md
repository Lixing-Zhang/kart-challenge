# Food Ordering API - Go Implementation

A production-ready Go API server implementing the [OpenAPI specification](../api/openapi.yaml) for a food ordering system.

## Features

- 🚀 Built with [chi/v5](https://github.com/go-chi/chi) router
- 📋 Full OpenAPI 3.1 specification compliance
- 🔐 API key authentication
- ✅ Promo code validation with multi-file lookup
- 🏗️ Clean architecture with dependency injection
- 📊 Structured JSON logging
- ⚙️ 12-factor app configuration
- 🛡️ Graceful shutdown
- 🔄 CORS support

## Project Structure

```
backend-challenge/
├── cmd/
│   └── server/          # Application entry point
│       └── main.go
├── internal/            # Private application code
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Domain models
│   └── store/           # Data persistence layer
├── pkg/                 # Public libraries
│   └── logger/          # Structured logging
├── .env.example         # Example environment variables
├── .gitignore          # Git ignore rules
├── go.mod              # Go module definition
└── README.md           # This file
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/Lixing-Zhang/kart-challenge.git
cd kart-challenge/backend-challenge
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Install dependencies:
```bash
go mod download
```

4. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`.

### Verify Installation

Check the health endpoint:
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2025-12-14T10:30:00Z",
  "version": "1.0.0"
}
```

## Configuration

All configuration is managed through environment variables following [12-factor app](https://12factor.net/) principles.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `HOST` | Server host | `0.0.0.0` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `API_KEYS` | Comma-separated valid API keys | `apitest` |
| `READ_TIMEOUT` | HTTP read timeout (seconds) | `15` |
| `WRITE_TIMEOUT` | HTTP write timeout (seconds) | `15` |
| `SHUTDOWN_TIMEOUT` | Graceful shutdown timeout (seconds) | `30` |
| `COUPON_FILE1_URL` | First coupon database URL | AWS S3 URL |
| `COUPON_FILE2_URL` | Second coupon database URL | AWS S3 URL |
| `COUPON_FILE3_URL` | Third coupon database URL | AWS S3 URL |

## API Endpoints

### Products

- `GET /api/product` - List all products
- `GET /api/product/{productId}` - Get product by ID

### Orders

- `POST /api/order` - Create new order (requires authentication)

### Health

- `GET /health` - Health check endpoint

For detailed API documentation, see the [OpenAPI specification](../api/openapi.yaml) or visit the [API docs](https://orderfoodonline.deno.dev/public/openapi.html).

## Development

### Running Tests

```bash
go test ./...
```

### Running with Custom Configuration

```bash
PORT=3000 LOG_LEVEL=debug go run cmd/server/main.go
```

### Building for Production

```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

## Architecture Decisions

### 12-Factor App Principles

1. **Codebase**: Single repository tracked in Git
2. **Dependencies**: Explicitly declared in `go.mod`
3. **Config**: All configuration via environment variables
4. **Backing Services**: Attached resources (coupon files)
5. **Build, Release, Run**: Strict separation of stages
6. **Processes**: Stateless server processes
7. **Port Binding**: Self-contained HTTP server
8. **Concurrency**: Horizontal scaling via process model
9. **Disposability**: Fast startup and graceful shutdown
10. **Dev/Prod Parity**: Keep environments similar
11. **Logs**: Structured JSON logs to stdout
12. **Admin Processes**: One-off tasks as separate commands

### Design Patterns

- **Dependency Injection**: Handlers receive dependencies explicitly
- **Middleware Chain**: Composable request processing
- **Repository Pattern**: Abstract data access in store layer
- **Handler Pattern**: Dedicated handlers for each endpoint

## Contributing

1. Create a feature branch from `main`
2. Make your changes
3. Run tests and ensure they pass
4. Submit a pull request

## License

MIT
