BINARY_DIR := bin
SERVER_BIN := $(BINARY_DIR)/server
CLIENT_BIN := $(BINARY_DIR)/client

SERVER_ADDR ?= 0.0.0.0:8285

.PHONY: all build build-server build-client build-client-windows run-server run-client clean

all: build

$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

build: build-server build-client

build-server: $(BINARY_DIR)
	go build -o $(SERVER_BIN) ./server/

build-client: $(BINARY_DIR)
	go build -o $(CLIENT_BIN) ./client/

# 64-bit Windows client, cross-compiled. Needs the mingw-w64 toolchain
# (brew install mingw-w64) because imgui uses cgo. Linked statically so it
# only depends on DLLs that ship with Windows; no console window.
build-client-windows: $(BINARY_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
	CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ \
	go build -trimpath -ldflags "-s -w -H windowsgui -extldflags=-static" \
	-o $(BINARY_DIR)/windows/Islanders.exe ./client/

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
