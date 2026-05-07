# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY server/ ./server/
COPY resources/ ./resources/

RUN go build -o /islanders-server ./server/

FROM alpine:3.21

COPY --from=builder /islanders-server /islanders-server

EXPOSE 8285

CMD ["/islanders-server"]
