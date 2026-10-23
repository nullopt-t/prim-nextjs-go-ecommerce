# Multi-stage production build for PRIM backend (context: ./backend)
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Compile statically linked binaries
RUN make build

# Final lightweight production image
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/server ./server
COPY --from=builder /app/bin/worker ./worker

EXPOSE 8080

CMD ["./server"]
