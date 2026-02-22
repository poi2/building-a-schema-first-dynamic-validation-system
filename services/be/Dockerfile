FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go workspace files
COPY go.work go.work

# Copy pkg/gen/go module
COPY pkg/gen/go pkg/gen/go

# Copy service files
COPY services/be services/be

# Build the service
WORKDIR /app/services/be
ENV GOTOOLCHAIN=auto
ENV GOWORK=off
RUN go mod download && go mod tidy
RUN go build -o /app/be .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/be .

EXPOSE 50052
CMD ["./be"]
