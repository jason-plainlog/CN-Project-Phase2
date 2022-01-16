all: server client

server: ./src/server/* ./src/server/models/*
	go env -w GO111MODULE="auto"
	go mod download
	go build -o server src/server/server.go

client: ./src/client/* ./src/client/api/* ./src/client/http/*
	go env -w GO111MODULE="auto"
	go mod download
	go build -o client src/client/client.go

.PHONY=clean
clean:
	rm -rf server client server_dir client_dir