.PHONY: build test docker-build run lint generate clean

APP_NAME := app
BIN_DIR := bin

build:
	@echo "Building application..."
	CGO_ENABLED=0 GOOS=linux go build -o $(BIN_DIR)/$(APP_NAME) cmd/app/main.go

test:
	@echo "Running tests..."
	go test -v -race ./...

docker-build:
	@echo "Building docker image..."
	docker build -t $(APP_NAME):latest .

run: build
	@echo "Starting application..."
	./$(BIN_DIR)/$(APP_NAME)

lint:
	@echo "Running linter..."
	golangci-lint run ./...

generate:
	@echo "Generating gRPC code..."
	protoc --proto_path=api/proto \
	       --go_out=pkg/api/rates/v1 --go_opt=paths=source_relative \
	       --go-grpc_out=pkg/api/rates/v1 --go-grpc_opt=paths=source_relative \
	       api/proto/rates.proto

clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)