# ====== Builder ======
FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache ca-certificates git

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Static build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -v -trimpath -tags netgo -ldflags="-s -w" -o /server ./main.go

# ====== Runtime (scratch) ======
FROM scratch

# CA bundle for HTTPS RPC calls
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# App binary
COPY --from=builder /server /server

# Non-root
USER 10001:10001

# Expose app port (container side)
EXPOSE 8080

# No envs baked in; pass them at runtime
ENTRYPOINT ["/server"]
