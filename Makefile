GO ?= go

.PHONY: check test test-live tidy fmt vet build race examples

check: fmt vet tidy test

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

test-live:
	KCP_LIBS_REQUIRE_LIVE=1 $(GO) test -timeout 20m -count=1 -v -run 'TestLive' ./impl/kcpstore/ ./factory/controller/

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
