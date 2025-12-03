# syntax=docker/dockerfile:1.7

###############################################
# Build stage
###############################################
FROM golang:1.24-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOTOOLCHAIN=auto

RUN apk add --no-cache ca-certificates git

# Leverage build cache for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Ensure directories exist for static assets consumed at runtime
RUN mkdir -p web templates static tls

# Build the API binary
RUN GOVULN_DISABLE=1 go build -trimpath -ldflags="-s -w" -o /out/flight-api ./

###############################################
# Runtime stage
###############################################
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

# Copy CA bundle for outbound HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy application artifacts
COPY --from=builder /out/flight-api /app/flight-api
COPY --from=builder /src/web /app/web
COPY --from=builder /src/templates /app/templates
COPY --from=builder /src/static /app/static
COPY --from=builder /src/tls /app/tls

EXPOSE 8080
ENTRYPOINT ["/app/flight-api"]
