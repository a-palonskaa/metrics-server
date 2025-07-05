GOFLAGS :=
LDFLAGS :=


.PHONY: all deps server agent test lint

all: deps server agent test lint

server:
	go build $(GOFLAGS) $(LDFLAGS) -o ./cmd/server/server ./cmd/server/main.go

agent:
	go build $(GOFLAGS) $(LDFLAGS) -o ./cmd/agent/agent ./cmd/agent/main.go

deps:
	go mod download

test:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm -rf coverage.out

lint:
	golangci-lint run


