BINARY_DIR := bin
SERVER_BIN := $(BINARY_DIR)/server
CLIENT_BIN := $(BINARY_DIR)/client

SERVER_ADDR ?= 0.0.0.0:8285

.PHONY: all build build-server build-client run-server run-client clean

all: build

$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

build: build-server build-client

build-server: $(BINARY_DIR)
	go build -o $(SERVER_BIN) ./server/

build-client: $(BINARY_DIR)
	go build -o $(CLIENT_BIN) ./client/

run-server:
	go run ./server/ -addr "$(SERVER_ADDR)"

run-client:
	go run ./client/

server: build-server
	./$(SERVER_BIN) -addr "$(SERVER_ADDR)"

client: build-client
	./$(CLIENT_BIN)

clean:
	rm -rf $(BINARY_DIR)
