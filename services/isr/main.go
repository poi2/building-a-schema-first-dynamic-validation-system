package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1/isrv1connect"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr/internal/handler"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr/internal/repository"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// Get configuration from environment
	dbURL := os.Getenv("CELO_DB_URL")
	if dbURL == "" {
		return fmt.Errorf("CELO_DB_URL environment variable is required")
	}

	port := os.Getenv("CELO_PORT")
	if port == "" {
		port = "50051"
	}

	// Connect to database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Ping database to verify connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	log.Println("Successfully connected to database")

	// Run migrations
	if err := runMigrations(ctx, pool); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize repository and handler
	repo := repository.NewSchemaRepository(pool)
	schemaHandler := handler.NewSchemaHandler(repo)

	// Create HTTP server with Connect
	mux := http.NewServeMux()
	path, handler := isrv1connect.NewSchemaRegistryServiceHandler(schemaHandler)
	mux.Handle(path, handler)

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("ISR service listening on %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migration := `
		CREATE TABLE IF NOT EXISTS schemas (
			id VARCHAR(36) PRIMARY KEY,
			version VARCHAR(20) UNIQUE NOT NULL,
			schema_binary BYTEA NOT NULL,
			size_bytes INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_schemas_version ON schemas(version);
	`

	_, err := pool.Exec(ctx, migration)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}
