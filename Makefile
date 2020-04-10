BUILD?= $(CURDIR)/bin
$(shell mkdir -p $(BUILD))
VERSION?= v$(shell git rev-list HEAD --count)
ARCH?= $(shell uname -m)
BRANCH_NAME ?= $(shell git rev-parse --abbrev-ref HEAD)
ifneq ($(BRANCH_NAME), master)
	BRANCH_NAME = dev
endif
export GO111MODULE=on
export GOPATH=$(go env GOPATH)

.PHONY: setup mod fmt test coverage lint vet checks build run gen gen-test clean help

setup: ## Install all the build and lint dependencies
	go get -u golang.org/x/tools
	go get -u golang.org/x/lint/golint

mod: ## Runs go mod on a project
	go mod verify
	go mod vendor
	go mod tidy

fmt: ## Runs goimports on all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do goimports -w "$$file"; done

test: ## Runs all unit tests
	echo 'mode: atomic' > coverage.txt && go test -covermode=atomic -coverprofile=coverage.txt -v -race \
	-timeout=30s ./log... ./internal/...

coverage: test ## Runs all the tests and opens the coverage report
	go tool cover -html=coverage.txt

lint: ## Runs all the linters
	golint ./internal/... ./cmd/... ./log/...
	golint ./test/integration-test/main_test.go

vet: ## Runs go vet
	go vet -atomic -bools -assign -copylocks -cgocall -asmdecl  ./...

checks: fmt lint vet ## Runs all checks for the project (go fmt, go lint, go vet)

build: ## Builds the project
	go build -o $(BUILD)/file-streamer $(CURDIR)

run: build ## Runs the project in production mode
	$(BUILD)/file-streamer -c ./configs/config.json

gen: ## Triggers code generation for the GRPC Server and Client API
	protoc -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:$(CURDIR)/internal/grpc ./api/*.proto

clean: ## Remove temporary files
	go clean $(CURDIR)
	rm -rf $(BUILD)
	rm -rf coverage.txt

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build
