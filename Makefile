GO ?= go

.PHONY: check test tidy fmt vet build race

check: fmt vet tidy test

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

tidy:
	$(GO) mod tidy

fmt:
	gofmt -l -w .

vet:
	$(GO) vet ./...

build:
	$(GO) build ./...
