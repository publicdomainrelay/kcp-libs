package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/publicdomainrelay/kcp-libs/internal/livekcp"
)

func TestRunKeepsTheCapAndDrainsTheQueue(t *testing.T) {
	livekcp.Require(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var out bytes.Buffer
	if err := Run(ctx, &out); err != nil {
		t.Fatalf("Run: %v\n%s", err, out.String())
	}
	printed := out.String()
	for _, want := range []string{
		"waits: waiting for a free slot (concurrencyPolicy=Forbid, maxConcurrent=1)",
		"started 5 of 5 items, peak running 2",
		"the duplicate-start guard refused 5 stale starts, leases held 0, succeeded 5",
	} {
		if !strings.Contains(printed, want) {
			t.Fatalf("output must contain %q, got:\n%s", want, printed)
		}
	}
}
