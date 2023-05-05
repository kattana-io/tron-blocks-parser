GO_PATH := $(shell go env GOPATH)

.PHONY : all
all : dep lint

lint: check-lint dep
	golangci-lint run --timeout=5m -c .golangci.yml

check-lint:
	@which golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_PATH)/bin v1.52.2

dep:
	@go mod tidy
	@go mod download
	@go mod vendor