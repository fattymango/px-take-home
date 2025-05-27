FROM golang:latest AS builder

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@latest

ENV PATH="/go/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init --parseDependency -g ./cmd/api/main.go -o ./api/swagger

# Enable CGO for SQLite support
ENV CGO_ENABLED=1
RUN go build -o server ./cmd/api/



FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    shellcheck \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app

COPY --from=builder /app/api/swagger/ ./api/swagger/

COPY --from=builder  /app/web/ ./web/

COPY --from=builder /app/server .

CMD ["/app/server"]
