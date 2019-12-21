# Go parameters
PROJECT_NAME := $(shell echo $${PWD\#\#*/})
PKG_LIST := $(shell go list ./...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

all: lint vet install docs

install: ## Run install
	@go install && echo Installed `date` && echo

dep:
	@go get -u ./...

lintAll: # run golangci-lint
	@golangci-lint run

lint: ## Run lint
	@golint -set_exit_status ${PKG_LIST}

vet: ## Run go vet
	@go vet ./...

vetclose: ## Run go vet with bodyclose
	go vet -vettool=$$(which bodyclose) ./...

check: ## Run gosimple and staticcheck
	@gosimple && staticcheck

test: ## Run unittests
	@go test -short ${PKG_LIST}

race: ## Run data race detector
	@go test -race -short ${PKG_LIST}

msan: ## Run memory sanitizer
	@go test -msan -short ${PKG_LIST}

build: ## Build the binary file
	@go build -i -v

clean: ## Remove previous build
	@go clean ./...

upgrade: ## Get latest libs
	@go get -u

linux:
	@env GOOS=linux GOARCH=amd64 go build -v -o ./build/$(PROJECT_NAME)

watch:
	@echo Watching for changes...
	@fswatch -or . -e ".*" -i "\\.go$$" | xargs -n1 -I{} make lintAll install tags

watchrun:
	@echo Watching for changes...
	@fswatch -or . -e ".*" -i "\\.go$$" | xargs -n1 -I{} make stop all start tags

start: # Start the server
	@$(PROJECT_NAME) &

stop: ## Stop the server
	@if pgrep $(PROJECT_NAME); then `pkill $(PROJECT_NAME)`; fi

tags:
	@gotags -R *.go > tags

docs: install
	@$(PROJECT_NAME) --docs=true > README.md
	@godocdown domain > domain/README.md
	@godocdown handlers > handlers/README.md

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all install lint test race msan build clean upgrade deployTest watch start stop tags help
