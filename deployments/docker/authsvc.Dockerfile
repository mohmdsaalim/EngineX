FROM golang:1.25.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o authsvc ./cmd/authsvc

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/authsvc .

EXPOSE 8082
EXPOSE 9091

CMD ["./authsvc"]