GO=go
BINARY=gochip
OUTPUT_DIR=.

fmt:
	go fmt ./...

build-desktop:
	go build -o $(OUTPUT_DIR)/$(BINARY) ./cmd/desktop 

build-web:
	rm -rf $(OUTPUT_DIR)/server; mkdir $(OUTPUT_DIR)/server
	cp ./cmd/web/server/index.html $(OUTPUT_DIR)/server/index.html 
	cp ./cmd/web/server/wasm_exec.js $(OUTPUT_DIR)/server/wasm_exec.js 
	cp ./cmd/web/server/serve.go $(OUTPUT_DIR)/server/serve.go 
	GOOS=js GOARCH=wasm $(GO) build -o $(OUTPUT_DIR)/server/main.wasm ./cmd/web

all: build-desktop build-web

run-web: build-web
	go run $(OUTPUT_DIR)/server/serve.go

