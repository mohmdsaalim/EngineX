FROM golang:1.25.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o engine ./cmd/engine

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/engine .

EXPOSE 9093

CMD ["./engine"]