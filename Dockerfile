# Stage 1
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gophermart ./cmd/gophermart

# Stage 2
FROM alpine:3.21
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /home/appuser
COPY --from=builder --chown=appuser:appgroup /app/gophermart .
USER appuser
EXPOSE 8080

CMD ["./gophermart"]
