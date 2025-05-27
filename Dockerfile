# Stage 1: Build
FROM golang:latest AS builder

WORKDIR /app

# Install swag CLI
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Add Go bin to PATH (needed for swag)
ENV PATH="/go/bin:${PATH}"

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod tidy

# Copy the source code
COPY . .

# Generate Swagger docs
RUN swag init --parseDependency -g ./cmd/api/main.go -o ./api/swagger

# Enable CGO for SQLite support
ENV CGO_ENABLED=1
RUN go build -o server ./cmd/api/



# Stage 2: Runtime
FROM debian:bookworm-slim

# Install libc and sqlite dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    shellcheck \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app

# Copy the swagger files
COPY --from=builder /app/api/swagger/ ./api/swagger/

# Copy the web files
COPY --from=builder  /app/web/ ./web/

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Run the application
CMD ["/app/server"]
