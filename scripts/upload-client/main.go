package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	isrv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1/isrv1connect"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: upload-client <version> [schema-file]")
	}

	version := os.Args[1]
	schemaFile := "/tmp/schema-descriptor.bin"
	if len(os.Args) >= 3 {
		schemaFile = os.Args[2]
	}

	isrURL := os.Getenv("CELO_ISR_URL")
	if isrURL == "" {
		isrURL = "http://localhost:50051"
	}

	// Read schema binary
	schemaBinary, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}

	fmt.Printf("📊 Schema size: %d bytes\n", len(schemaBinary))
	fmt.Printf("🚀 Uploading schema version %s to ISR (%s)...\n", version, isrURL)

	// Create Connect client
	client := isrv1connect.NewSchemaRegistryServiceClient(
		http.DefaultClient,
		isrURL,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Upload schema
	req := connect.NewRequest(&isrv1.UploadSchemaRequest{
		Version:      version,
		SchemaBinary: schemaBinary,
	})

	resp, err := client.UploadSchema(ctx, req)
	if err != nil {
		log.Fatalf("❌ Upload failed: %v", err)
	}

	fmt.Println("✅ Schema uploaded successfully!")
	fmt.Printf("  Schema ID: %s\n", resp.Msg.Metadata.Id)
	fmt.Printf("  Version: %s\n", resp.Msg.Metadata.Version)
	fmt.Printf("  Size: %d bytes\n", resp.Msg.Metadata.SizeBytes)
	fmt.Printf("  Created At: %s\n", resp.Msg.Metadata.CreatedAt.AsTime().Format(time.RFC3339))
}
