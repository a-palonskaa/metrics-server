GOFLAGS :=
LDFLAGS :=

AGENT_SOURCES = ./cmd/agent/...
AGENT_EXECUTE := ./cmd/agent/agent

SERVER_SOURCES = ./cmd/server/...
SERVER_EXECUTE := ./cmd/server/server

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

clean:
	rm -f ./cmd/agent/agent ./cmd/server/server coverage.out

