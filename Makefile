GOFLAGS :=
LDFLAGS :=

AGENT_SOURCES = ./cmd/agent/...
AGENT_EXECUTE := ./cmd/agent/agent

SERVER_SOURCES = ./cmd/server/...
SERVER_EXECUTE := ./cmd/server/server

SERVER_MAIN := ./cmd/server/main.go
SWAGGER_DIR := ./api/openapi

.PHONY: all deps server agent test lint

all: deps server agent test lint

server: deps
	go build $(GOFLAGS) $(LDFLAGS) -o $(SERVER_EXECUTE) $(SERVER_SOURCES)

agent: deps
	go build $(GOFLAGS) $(LDFLAGS) -o $(AGENT_EXECUTE) $(AGENT_SOURCES)

deps:
	go mod download
	go mod vendor

test:
	go test ./... -v -coverprofile=coverage.out

test_results: test
	go tool cover -html=coverage.out
	rm -rf coverage.out

lint: deps
	golangci-lint run

swagger:
	swag init --generalInfo $(SERVER_MAIN) --output $(SWAGGER_DIR)

proto:
	protoc -I./googleapis \
	--go_out=./gen/proto --go_opt=paths=source_relative \
    --go-grpc_out=./gen/proto --go-grpc_opt=paths=source_relative \
    --proto_path=api/proto \
    api/proto/metricservice.proto

clean:
	rm -f $(AGENT_EXECUTE) $(SERVER_EXECUTE) coverage.out

