GO = go
BINARY_NAME = main.wasm
BUILD_DIR = server
SOURCE_DIR = .

build:
		GOOS=js GOARCH=wasm $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(SOURCE_DIR)

fmt: 
		$(GO) fmt ./...

run: build
		$(GO) run server/serve.go
