# A Game

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

To run the client/server:

```bash
cd server
go run .
```
