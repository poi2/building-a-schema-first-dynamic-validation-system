FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go workspace files
COPY go.work go.work

# Copy pkg/gen/go module
COPY pkg/gen/go pkg/gen/go

# Copy service files
COPY services/isr services/isr

# Build the service
WORKDIR /app/services/isr
ENV GOTOOLCHAIN=auto
ENV GOWORK=off
RUN go mod download
RUN go build -o /app/isr .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/isr .

EXPOSE 50051
CMD ["./isr"]
