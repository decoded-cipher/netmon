# --- Build stage ---
FROM golang:1.25-alpine AS builder

# gcc + musl-dev for CGO (go-sqlite3); sqlite-dev for headers
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o netmon ./cmd/netmon

# --- Runtime stage ---
FROM alpine:3.21

# ca-certificates for HTTPS speed tests; sqlite-libs for the CGO runtime
RUN apk add --no-cache ca-certificates sqlite-libs

# Store database in /data so it can be mounted as a volume
WORKDIR /data

COPY --from=builder /app/netmon /usr/local/bin/netmon

EXPOSE 8080
VOLUME ["/data"]

CMD ["netmon"]
