GO=go

PACKAGE=github.com/m110/cort
EXAMPLE_CMD=cmd/example
CLIENT_CMD=cmd/client

all: build

build: example client

example:
	@$(GO) vet ./$(EXAMPLE_CMD)/main.go
	@$(GO) build -o ./bin/example $(PACKAGE)/$(EXAMPLE_CMD)

client:
	@$(GO) vet ./$(CLIENT_CMD)/main.go
	@$(GO) build -o ./bin/client $(PACKAGE)/$(CLIENT_CMD)

test:
	@$(GO) test ./...

clean:
	@rm -f ./bin/*
