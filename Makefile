.DEFAULT_GOAL := help
.PHONY : context clean dependencies build install test changelog

VERSION 		:= $(shell git describe --tags --abbrev=0 --exact-match 2> /dev/null)

LDFLAGS     := -w -s -X main.version=$(VERSION)
GO111MODULE := on

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

context: ## Print context information.
	@echo "Version:     ${VERSION}"
	@echo "Flags:       ${LDFLAGS}"

clean: ## Clean up working directory.
	rm -rf dist
	rm -f timesheets

dependencies: ## Install dependencies.
	go mod tidy

test: clean ## Run tests.
	go test

build: clean ## Build timesheets binary.
	go build -ldflags "$(LDFLAGS)"

install: clean ## Install timesheets binary.
	go install -ldflags "$(LDFLAGS)"

changelog: ## Generate changelog.
	git cliff -c .cliff.yaml -o CHANGELOG.md --tag "$(TAG)"
