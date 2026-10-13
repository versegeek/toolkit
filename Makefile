include Makefile.def.mk

.PHONY: all
all: lint test build

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test -race ./...

.PHONY: lint
lint: fmt vet golangci-lint

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: golangci-lint
golangci-lint:
	golangci-lint run ./...

.PHONY: tidy
tidy:
	GOPROXY=$(GOPROXY) go mod tidy

# Regenerate wire_gen.go for every injector package.
.PHONY: wire
wire:
	$(WIRE) gen $(WIRE_PKGS)

# Fail if any wire_gen.go is out of date.
.PHONY: wire-check
wire-check:
	$(WIRE) diff $(WIRE_PKGS)
