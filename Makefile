desktop:
	go build -o gochip ./cmd/desktop

web:
	GOOS=js GOARCH=wasm go build -o server/main.wasm ./cmd/web -o server/main.wasm

fmt:
	go fmt ./...

run-desktop: desktop
	./gochip

run-web: web
	go run server/serve.go

