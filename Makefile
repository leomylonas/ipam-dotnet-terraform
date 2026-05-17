BINARY_NAME ?= terraform-provider-dotnet-ipam
DEV_DIR ?= .terraform-dev

.PHONY: help build test testacc dev-install lint tidy

help:
	@echo "Available targets:"
	@echo "  make help        - Show this help"
	@echo "  make tidy        - Run go mod tidy"
	@echo "  make build       - Build provider binary ($(BINARY_NAME))"
	@echo "  make test        - Run go test ./..."
	@echo "  make testacc     - Run acceptance tests (requires IPAM_* env vars)"
	@echo "  make dev-install - Build and place binary in $(DEV_DIR)/"
	@echo "  make lint        - Alias for test"

build:
	go build -o $(BINARY_NAME) .

test:
	go test ./...

testacc:
	IPAM_ACC=1 go test ./internal/provider -v -run 'TestAcc' -count=1

dev-install: build
	mkdir -p $(DEV_DIR)
	mv $(BINARY_NAME) $(DEV_DIR)/$(BINARY_NAME)

lint:
	go test ./...

tidy:
	go mod tidy
