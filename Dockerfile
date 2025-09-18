FROM golang:1.25.1 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ipam-api cmd/ipam-api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o ipam-cli ./cmd/cli

FROM alpine
RUN apk add --no-cache bash

WORKDIR /app
COPY config-docker-compose.json ./

COPY --from=builder /app/ipam-api .
COPY --from=builder /app/ipam-cli .
COPY nhn_internal_ca.crt /etc/ssl/certs/
CMD ["/app/ipam-api"]
