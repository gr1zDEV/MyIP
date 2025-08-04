# Stage 1: Build the Go app
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod ./
RUN go mod download

# Copy source and add TLS support
COPY main.go ./
RUN apk add --no-cache ca-certificates

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myip .

# Stage 2: Minimal container
FROM scratch

COPY --from=builder /app/myip /myip
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8000
CMD ["/myip"]
