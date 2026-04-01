FROM golang:1.25.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o risksvc ./cmd/risksvc

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/risksvc .

EXPOSE 9092

CMD ["./risksvc"]