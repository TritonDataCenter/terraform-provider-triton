SWEEP?=us-central-1
PKG_NAME=triton

default: fmt lint install generate

build: fmtcheck ## Build the provider
	go build -v ./...

install: build
	go install -v ./...

test: fmtcheck ## Test the provider
	go test -v -cover -timeout=120s -parallel=10 ./...

generate:
	cd tools; go generate ./tools.go

testacc: fmtcheck ## Test acceptance of the provider
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

sweep:
	@echo "WARNING: This will destroy acceptance test infrastructure. Use only in development accounts."
	@echo "   10 seconds to hit ^C."
	sleep 10; TF_LOG=DEBUG go test $(TEST) -v -sweep=$(SWEEP) -sweep-run=$(SWEEPARGS) -timeout 60m

fmt: ## Run gofmt across all go files
	gofmt -s -w -e .

fmtcheck: ## Check that code complies with gofmt requirements
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck: ## Check for unchecked errors
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

.PHONY: build test testacc fmt fmtcheck errcheck test-compile generate

help:
	@echo "Valid targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
