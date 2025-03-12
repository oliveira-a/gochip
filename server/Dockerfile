FROM golang:1.23

WORKDIR /app
COPY . .

RUN GOOS=js GOARCH=wasm go build -o ./server/gochip.wasm github.com/oliveira-a/gochip

WORKDIR /app/server

RUN cp $(go env GOROOT)/misc/wasm/wasm_exec.js .

EXPOSE 8080

ENTRYPOINT ["go", "run", "server.go"]

