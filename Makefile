.PHONY: help proto-generate proto-lint clean docker-up docker-down docker-logs docker-clean schema-upload schema-pull test fmt lint ci

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

proto-generate: ## Generate code from proto files using buf
	buf generate

proto-lint: ## Lint proto files using buf
	buf lint

clean: ## Clean generated code
	find pkg/gen/go -name '*.pb.go' -delete 2>/dev/null || true
	find pkg/gen/go -name '*connect.go' -delete 2>/dev/null || true
	rm -rf pkg/gen/ts/

docker-up: ## Start all services with docker compose
	docker compose up -d

docker-down: ## Stop all services
	docker compose down

docker-logs: ## Show logs from all services
	docker compose logs -f

docker-clean: ## Stop services and remove volumes
	docker compose down -v

schema-upload: ## Upload schema to ISR (usage: make schema-upload VERSION=1.0.0)
	@./scripts/upload-schema.sh $(VERSION)

schema-pull: ## Pull latest schema from ISR (usage: make schema-pull MAJOR=1 MINOR=0)
	@./scripts/pull-schema.sh $(MAJOR) $(MINOR)

test: ## Run tests for all services
	@echo "Running ISR tests..."
	cd services/isr && go test -v -race -coverprofile=coverage.out ./...

fmt: ## Format Go code
	@echo "Formatting Go code..."
	gofmt -s -w services/isr
	cd services/isr && go mod tidy

lint: proto-lint ## Lint proto files and Go code
	@echo "Linting ISR service..."
	cd services/isr && go vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then \
		cd services/isr && staticcheck ./...; \
	else \
		echo "staticcheck not installed, skipping (install: go install honnef.co/go/tools/cmd/staticcheck@latest)"; \
	fi

ci: proto-lint fmt test ## Run all CI checks (lint, format, test)
