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

### Setup

```bash
# Start all services
docker compose up -d
```

### Connection Information

* **Container-to-container**: Use service names (e.g., `db:5432`, `isr:50051`)
* **Host-to-container**: Use `localhost` with fixed ports (e.g., `localhost:5432`)

## Documentation

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

```bash
# Lint markdown files
npm run lint:md

# Auto-fix markdown issues
npm run lint:md:fix
```

## License

This is a proof-of-concept project for educational purposes.
