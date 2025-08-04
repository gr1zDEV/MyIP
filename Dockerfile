# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum if you have them
COPY go.mod ./


# Download dependencies
RUN go mod download

# Now copy the rest of the source code
COPY main.go .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myip .

# Stage 2: Create the final, minimal image
FROM scratch

COPY --from=builder /app/myip /myip
EXPOSE 8000
CMD ["/myip"]
