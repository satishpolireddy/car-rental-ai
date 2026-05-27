# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Cache dependency downloads separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# ── Runtime stage (distroless for minimal attack surface) ────────────────────
FROM gcr.io/distroless/static-debian12
COPY --from=builder /server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
