FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api-gateway ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /api-gateway /api-gateway
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api-gateway"]
