.PHONY: help proto-generate proto-lint clean docker-up docker-down docker-logs docker-clean

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
	rm -rf pkg/gen/go/**/*.pb.go pkg/gen/go/**/*connect.go pkg/gen/ts/

docker-up: ## Start all services with docker compose
	docker compose up -d

docker-down: ## Stop all services
	docker compose down

docker-logs: ## Show logs from all services
	docker compose logs -f

docker-clean: ## Stop services and remove volumes
	docker compose down -v
