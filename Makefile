.PHONY: help proto-generate proto-lint clean test fmt lint lint-md ci

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

lint-md: ## Lint markdown files
	@echo "Linting markdown files..."
	npm run lint:md

ci: proto-lint lint fmt test lint-md ## Run all CI checks (proto, go lint, format, test, markdown)
