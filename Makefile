.PHONY: proto install-grpc-tools run-data run-search tidy build

## install-grpc-tools: install protoc-gen-go and protoc-gen-go-grpc (run once)
install-grpc-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

## proto: regenerate Go code from proto/fund.proto
proto:
	protoc \
		--proto_path=proto \
		--go_out=gen \
		--go_opt=paths=source_relative \
		--go-grpc_out=gen \
		--go-grpc_opt=paths=source_relative \
		fund.proto

## tidy: tidy go modules
tidy:
	go mod tidy

## build: build all packages
build:
	go build ./...

## run-data: start the CRUD data-service on :9090
run-data:
	go run ./data-service/cmd/main.go

## run-search: start the search-service on :9091
run-search:
	go run ./search-service/cmd/main.go