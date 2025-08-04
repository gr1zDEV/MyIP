# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myip .

# Stage 2: Create the final, minimal image
FROM scratch

COPY --from=builder /app/myip /myip
EXPOSE 8000
CMD ["/myip"]
