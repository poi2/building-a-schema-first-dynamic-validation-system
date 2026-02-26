package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	userv1connect "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/user/v1/userv1connect"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/handler"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/repository"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func run() error {
	// Get configuration from environment
	port := os.Getenv("CELO_PORT")
	if port == "" {
		port = "50052"
	}

	// YAML file path for user data
	dataDir := os.Getenv("CELO_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	userYAMLPath := filepath.Join(dataDir, "user.yaml")

	// Initialize YAML repository
	userRepo, err := repository.NewYAMLUserRepository(userYAMLPath)
	if err != nil {
		return fmt.Errorf("failed to initialize user repository: %w", err)
	}
	log.Printf("Initialized YAML user repository at %s", userYAMLPath)

	// Initialize handler with YAML repository
	userHandler := handler.NewUserHandler(userRepo)

	// Create HTTP server with Connect
	mux := http.NewServeMux()
	// Use standard protovalidate interceptor
	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
	)
	path, connectHandler := userv1connect.NewUserServiceHandler(userHandler, interceptors)
	mux.Handle(path, connectHandler)

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

	log.Printf("BE service listening on %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}
