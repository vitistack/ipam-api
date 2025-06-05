# Build stage
FROM golang:1.24.2 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ipam-api cmd/ipam-api/main.go

# Final minimal image
# FROM gcr.io/distroless/base-debian12
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY config.json ./

COPY --from=builder /app/ipam-api .
COPY nhn_internal_ca.crt /etc/ssl/certs/
CMD ["/app/ipam-api"]
