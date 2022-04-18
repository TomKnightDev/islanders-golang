# syntax=docker/dockerfile:1

FROM golang:1.18

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY /server ./server
COPY /resources ./resources

RUN go build -o /islanders-server ./server

EXPOSE 8285

CMD [ "/islanders-server" ]