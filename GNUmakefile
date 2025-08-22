SWEEP?=us-central-1
PKG_NAME=triton

default: fmt lint install generate

build: fmtcheck ## Build the provider
	go build -v ./...

install: build
	go install -v ./...

test: fmtcheck ## Test the provider
	go test -v -cover -timeout=120s -parallel=10 ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./tools.go

testacc: fmtcheck ## Test acceptance of the provider
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

sweep:
	@echo "WARNING: This will destroy acceptance test infrastructure. Use only in development accounts."
	@echo "   10 seconds to hit ^C."
	sleep 10; TF_LOG=DEBUG go test ./... -v -sweep=$(SWEEP) -sweep-run=$(SWEEPARGS) -timeout 60m

fmt: ## Run gofmt across all go files
	gofmt -s -w -e .

fmtcheck: ## Check that code complies with gofmt requirements
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck: ## Check for unchecked errors
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

.PHONY:  fmt lint test testacc build install generate fmtcheck errcheck

help:
	@echo "Valid targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
