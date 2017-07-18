TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

default: build

build: fmtcheck ## Build the provider
	go install

test: fmtcheck ## Test the provider
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: fmtcheck ## Test acceptance of the provider
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

vet: ## Run go vet across the provider
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt: ## Run gofmt across all go files
	gofmt -w $(GOFMT_FILES)

fmtcheck: ## Check that code complies with gofmt requirements
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck: ## Check for unchecked errors
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

vendor-status: ## Run govendor status over the provider
	@govendor status

vendor-triton: ## Update triton specific vendored packages
	govendor update github.com/joyent/triton-go
	govendor update github.com/joyent/triton-go/account
	govendor update github.com/joyent/triton-go/authentication
	govendor update github.com/joyent/triton-go/client
	govendor update github.com/joyent/triton-go/compute
	govendor update github.com/joyent/triton-go/identity
	govendor update github.com/joyent/triton-go/network

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./aws"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

.PHONY: build test testacc vet fmt fmtcheck errcheck vendor-status test-compile

help:
	@echo "Valid targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
