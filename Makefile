GO ?= go

.PHONY: check test tidy fmt vet build race examples

check: fmt vet tidy test

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

examples:
	$(GO) run ./examples/controller
	$(GO) run ./examples/admission
	$(GO) run ./examples/workloads
	$(GO) run ./examples/pki
	$(GO) run ./examples/policy
	$(GO) run ./examples/dns

tidy:
	$(GO) mod tidy

fmt:
	gofmt -l -w .

vet:
	$(GO) vet ./...

build:
	$(GO) build ./...
