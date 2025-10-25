# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy everything (source + go.mod)
COPY . .

# Build the binary directly
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kv-server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/kv-server .

RUN mkdir -p /data

EXPOSE 8080

CMD ["./kv-server"]