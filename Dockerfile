FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o relay .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/relay /usr/local/bin/relay

EXPOSE 3000

ENTRYPOINT ["relay"]
