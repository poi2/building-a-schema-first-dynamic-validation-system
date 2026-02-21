# Building a Schema-First Dynamic Validation System

A proof-of-concept for a schema-first dynamic validation system using ConnectRPC and protovalidate, ensuring seamless contract synchronization across FE, BFF, and BE.

## Overview

This project demonstrates a dynamic validation system where:

* `.proto` files serve as the single source of truth for validation rules
* Validation schemas can be updated at runtime without service restarts
* Business logic (e.g., user plan-based restrictions) is enforced consistently across all layers using CEL expressions

## Architecture

* **Frontend**: TypeScript, React, `@connectrpc/connect-web`
* **BFF**: Node.js, TypeScript, `@connectrpc/connect-node`
* **Backend**: Go, `connect-go`, `protovalidate-go`
* **ISR (Internal Schema Registry)**: Go service for schema distribution
* **Database**: PostgreSQL (with separate databases for ISR and backend)

## Getting Started

### Prerequisites

* Docker and Docker Compose
* Node.js 20+ (for local development)
* Go 1.21+ (for local development)
* [Buf CLI](https://docs.buf.build/installation) (for proto code generation)

### Setup

```bash
# 1. Generate code from proto files
make proto-generate

# 2. Start services
docker compose up -d

# 3. Check service status
docker compose ps

# 4. Upload initial schema
./scripts/upload-schema.sh 1.0.0
```

### Connection Information

* **Container-to-container**: Use service names (e.g., `db:5432`, `isr:50051`)
* **Host-to-container**: Use `localhost` with fixed ports (e.g., `localhost:5432`)

## Documentation

### For Developers

* **[Developer Guide](docs/DEVELOPER_GUIDE.md)** - Setup, Git hooks, coding conventions, and contribution workflow

### Design Documentation

Detailed design documentation is available in the `docs/` directory:

* [Requirements](docs/000.requirement.md) - Project goals and PoC scenarios
* [High-Level Design](docs/001-DD.001.high-level-design.md) - Overall architecture
* [Schema Management](docs/001-DD.002.schema-management-and-distribution.md) - SemVer strategy and distribution
* [Validation Strategy](docs/001-DD.003.validation-strategy.md) - Context enrichment pattern
* [Data Model](docs/001-DD.004.data-model-and-api-interface.md) - Database schema and API interface
* [Implementation Guidelines](docs/001-DD.005.other.md) - Error handling and observability
* [Monorepo Structure](docs/001-DD.006.monorepo.md) - Project layout and dependencies

## Key Features

1. **Hot Reload**: Schema updates propagate to all services within minutes without restarts
2. **Multi-Layer Validation**: Optimistic validation in FE, authoritative validation in BE
3. **Context Enrichment**: Business rules (user plans) injected into proto messages for CEL-based validation
4. **Fail-Safe**: Services fallback to bundled schemas if ISR is unavailable

## Development

### Code Generation

```bash
# Generate code from proto files
make proto-generate

# Lint proto files
make proto-lint
```

### Testing

```bash
# Run all tests
make test

# Format code
make fmt

# Lint code
make lint

# Run all CI checks (lint, format, test)
make ci
```

### Docker Commands

```bash
# Start all services
make docker-up

# Stop all services
make docker-down

# View logs
make docker-logs

# Clean up (remove volumes)
make docker-clean
```

### Schema Management

```bash
# Upload schema to ISR
./scripts/upload-schema.sh 1.0.0

# Pull latest schema from ISR
./scripts/pull-schema.sh 1 0

# Or use make targets
make schema-upload VERSION=1.0.0
make schema-pull MAJOR=1 MINOR=0
```

### Linting

```bash
# Lint markdown files
npm run lint:md

# Auto-fix markdown issues
npm run lint:md:fix
```

### Project Structure

```text
celo/
├── go.work              # Go Workspaces configuration
├── package.json         # Node.js Workspaces configuration
├── buf.yaml             # Buf configuration
├── buf.gen.yaml         # Code generation settings
├── Makefile             # Common development tasks
├── docker-compose.yml   # Service orchestration
├── proto/               # Proto definitions (single source of truth)
│   ├── common/v1/
│   ├── user/v1/
│   ├── post/v1/
│   └── isr/v1/
├── pkg/
│   └── gen/             # Generated code (shared module)
│       ├── go/
│       └── ts/
├── services/
│   ├── isr/             # Internal Schema Registry
│   ├── be/              # Backend service
│   ├── bff/             # Backend for Frontend
│   └── fe/              # Frontend
├── docker/init-db/      # Database initialization scripts
└── tests/
    └── e2e/             # End-to-end tests
```

## License

This is a proof-of-concept project for educational purposes.
