# Islanders

A game developed in Go.

## Requirements

This project currently depends on cgo, so you'll need to install GCC.

## Installation

To setup the server using docker:

```docker
docker build -t gameserver .
```

```bash
docker run -d gameserver
```

To run the server:

```bash
cd server
go run .
```

To run the client:

```bash
cd client
go run .
```
