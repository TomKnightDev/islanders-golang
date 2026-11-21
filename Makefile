BINARY_DIR := bin
SERVER_BIN := $(BINARY_DIR)/server
CLIENT_BIN := $(BINARY_DIR)/client

SERVER_ADDR ?= 0.0.0.0:8285

.PHONY: all build build-server build-client build-client-windows package-windows build-client-mac package-mac run-server run-client clean

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

# The Windows client zipped for a release.
package-windows: build-client-windows
	cd $(BINARY_DIR)/windows && rm -f ../Islanders-windows.zip && zip -q ../Islanders-windows.zip Islanders.exe

# Universal macOS client (Apple Silicon and Intel), built on a Mac: each
# architecture with clang, then joined with lipo.
MAC_APP := $(BINARY_DIR)/mac/Islanders.app
MAC_MIN := 12.0
build-client-mac: $(BINARY_DIR)
	mkdir -p $(BINARY_DIR)/mac
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC="clang -arch arm64" CXX="clang++ -arch arm64" 	MACOSX_DEPLOYMENT_TARGET=$(MAC_MIN) go build -trimpath -ldflags "-s -w" -o $(BINARY_DIR)/mac/client-arm64 ./client/
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC="clang -arch x86_64" CXX="clang++ -arch x86_64" 	MACOSX_DEPLOYMENT_TARGET=$(MAC_MIN) go build -trimpath -ldflags "-s -w" -o $(BINARY_DIR)/mac/client-amd64 ./client/
	lipo -create -output $(BINARY_DIR)/mac/islanders-client $(BINARY_DIR)/mac/client-arm64 $(BINARY_DIR)/mac/client-amd64

# Islanders.app, ad-hoc signed and zipped for a release. The launcher script
# runs the client from ~/Library/Application Support/Islanders so it can
# save its window layout.
package-mac: build-client-mac
	rm -rf $(MAC_APP) && mkdir -p $(MAC_APP)/Contents/MacOS
	cp packaging/mac/Info.plist $(MAC_APP)/Contents/Info.plist
	cp packaging/mac/Islanders $(MAC_APP)/Contents/MacOS/Islanders
	cp $(BINARY_DIR)/mac/islanders-client $(MAC_APP)/Contents/MacOS/islanders-client
	chmod +x $(MAC_APP)/Contents/MacOS/*
	codesign --force --deep -s - $(MAC_APP)
	rm -f $(BINARY_DIR)/Islanders-mac.zip && ditto -c -k --keepParent $(MAC_APP) $(BINARY_DIR)/Islanders-mac.zip

# Local servers start with dev tools on (game speed, from the client's F12
# window). The production deploy doesn't use -dev.
run-server:
	go run ./server/ -addr "$(SERVER_ADDR)" -dev

run-client:
	go run ./client/

server: build-server
	./$(SERVER_BIN) -addr "$(SERVER_ADDR)" -dev

client: build-client
	./$(CLIENT_BIN)

clean:
	rm -rf $(BINARY_DIR)
