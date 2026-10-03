GO ?= go

.PHONY: check test test-live tidy fmt vet build race examples

check: fmt vet tidy test

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

test-live:
	@bash scripts/live.sh '$(GO) test -timeout 20m -count=1 -v -run TestLive ./impl/kcpstore/ ./factory/controller/ ./internal/livekcp/'

examples:
	@bash scripts/live.sh 'for e in controller admission workloads pki policy dns; do $(GO) run ./examples/$$e; done' 


tidy:
	$(GO) mod tidy

fmt:
	gofmt -l -w .

vet:
	$(GO) vet ./...

build:
	$(GO) build ./...
