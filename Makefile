GOFLAGS :=
LDFLAGS :=

AGENT_SRC_DIR := ./cmd/agent
AGENT_SRC := main.go cmd_args.go flags.go
AGENT_SOURCES = $(addprefix $(AGENT_SRC_DIR)/, $(AGENT_SRC))
AGENT_EXECUTE := ./cmd/agent/agent

SERVER_SRC_DIR := ./cmd/server
SERVER_SRC := main.go cmd_args.go flags.go
SERVER_SOURCES = $(addprefix $(SERVER_SRC_DIR)/, $(SERVER_SRC))
SERVER_EXECUTE := ./cmd/server/server

.PHONY: all deps server agent test lint

all: deps server agent test lint

server:
	go build $(GOFLAGS) $(LDFLAGS) -o $(SERVER_EXECUTE) $(SERVER_SOURCES)

agent:
	go build $(GOFLAGS) $(LDFLAGS) -o $(AGENT_EXECUTE) $(AGENT_SOURCES)

deps:
	go mod download
	go mod vendor

test:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm -rf coverage.out

lint:
	go mod download
	golangci-lint run

clean:
	rm -f ./cmd/agent/agent ./cmd/server/server coverage.out

