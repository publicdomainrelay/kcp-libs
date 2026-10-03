package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/publicdomainrelay/kcp-libs/internal/livekcp/livetest"
)

func TestRunKeepsTheCapAndDrainsTheQueue(t *testing.T) {
	livetest.Require(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var out bytes.Buffer
	if err := Run(ctx, &out); err != nil {
		t.Fatalf("Run: %v\n%s", err, out.String())
	}
	printed := out.String()
	for _, want := range []string{
		"item-2 waits: waiting for a free slot (concurrencyPolicy=Allow, maxConcurrent=2)",
		"the batch allows 2 at once and 5 items were created",
		"started 5 of 5 items, peak observed running 2",
		"succeeded 5",
		"the duplicate-start guard refused 5 stale copies and allowed 10 legitimate starts",
	} {
		if !strings.Contains(printed, want) {
			t.Fatalf("output must contain %q, got:\n%s", want, printed)
		}
	}
}
