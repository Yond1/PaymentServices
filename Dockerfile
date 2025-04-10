FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /paymentSystem ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /paymentSystem .
COPY config.yaml .
EXPOSE 9090
CMD ["./paymentSystem"]