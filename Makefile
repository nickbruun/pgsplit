all: check test

check:
	golangci-lint run ./...

fmt:
	go fmt ./...

test:
	go test -v ./...

.PHONY: check fmt test
