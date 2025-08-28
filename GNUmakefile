SWEEP?=us-central-1
PKG_NAME=triton

default: fmt lint install generate

build: lint ## Build the provider
	go build -v ./...

install: build
	go install -v ./...

test: lint ## Test the provider
	go test -v -cover -timeout=120s -parallel=10 ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./tools.go

testacc: lint ## Test acceptance of the provider
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

sweep:
	@echo "WARNING: This will destroy acceptance test infrastructure in $(SWEEP). Use only in development accounts."
	@echo "   10 seconds to hit ^C."
	TF_LOG=DEBUG go test ./... -v -sweep=$(SWEEP) -sweep-run=$(SWEEPARGS) -timeout 60m

fmt: ## Run gofmt across all go files
	gofmt -s -w -e .

.PHONY:  fmt lint test testacc build install generate

help:
	@echo "Valid targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
