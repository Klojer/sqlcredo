
GIT_VERSION := $(shell git describe --tags --always)

# Override version via: make release VERSION='my-custom-version'
VERSION ?= $(GIT_VERSION)

dependencies:
	go mod tidy

install-golangci-lint:
	if [ -f "$(shell go env GOPATH)/bin/golangci-lint" ]; then \
		echo 'golangci-lint already installed'; \
	else \
		echo 'golangci-lint binary not found, downloading'; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b $(shell go env GOPATH)/bin v1.54.1; \
	fi

install-git-hooks:
	pre-commit install

test: dependencies
	go test -v -race -shuffle=on ./...

test-cov: test
	go test -v -coverprofile=cover.out $(shell go list ./... | grep -v /internal/mocks)
	go tool cover -html cover.out -o cover.html
	xdg-open cover.html

lint: install-golangci-lint dependencies
	golangci-lint run ./...

version:
	echo $(VERSION)
