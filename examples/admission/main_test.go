package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunKeepsTheCapAndDrainsTheQueue(t *testing.T) {
	var out bytes.Buffer
	if err := Run(context.Background(), &out); err != nil {
		t.Fatalf("Run: %v\n%s", err, out.String())
	}
	printed := out.String()
	for _, want := range []string{
		`item-2 waits: waiting for a free slot (concurrencyPolicy=Forbid, maxConcurrent=1)`,
		"started 5 of 5 items, peak running 2",
		"the duplicate-start guard refused 5 stale starts, leases held 0, succeeded 5",
	} {
		if !strings.Contains(printed, want) {
			t.Fatalf("output must contain %q, got:\n%s", want, printed)
		}
	}
}
