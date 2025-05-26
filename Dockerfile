FROM golang:latest


WORKDIR /app

# Copy source and go modules
COPY go.mod go.sum ./
RUN go mod download && go mod tidy

COPY . .

# Build the Go binary with CGO enabled
ENV CGO_ENABLED=1
RUN go build -o server ./cmd/api/


CMD ["/app/server"]
