FROM golang:1.20 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main main.go
FROM alpine:latest
ENV REDIS_PRICE_ADDRESS=redis-service:6379
COPY --from=builder main /app/main
CMD ["/app/main"]
