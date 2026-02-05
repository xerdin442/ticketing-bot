# --- Stage 1: Build ---
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/api_bin ./cmd/api
RUN go build -o /app/worker_bin ./cmd/worker

# --- Stage 2: Final Runtime ---
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/api_bin .
COPY --from=builder /app/worker_bin .

CMD ["./api_bin"]