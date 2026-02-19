# Placeholder for Backend service Dockerfile
# To be implemented in Milestone 3

FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go workspace files
COPY go.work go.work

# Copy pkg/gen/go module
COPY pkg/gen/go pkg/gen/go

# Copy service files
COPY services/be services/be

# Build the service
WORKDIR /app/services/be
RUN go mod download
RUN go build -o /app/be ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/be .

EXPOSE 50052
CMD ["./be"]
