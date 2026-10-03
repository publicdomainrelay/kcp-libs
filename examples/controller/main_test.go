package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/publicdomainrelay/kcp-libs/internal/livekcp"
)

func TestRunDrivesWidgetsToCompletion(t *testing.T) {
	livekcp.Require(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var out bytes.Buffer
	if err := Run(ctx, &out); err != nil {
		t.Fatalf("Run: %v\n%s", err, out.String())
	}
	printed := out.String()
	for _, want := range []string{
		"kcp published widgets at https://",
		"alpha reached Succeeded after 1 passes",
		"beta reached Succeeded after 1 passes",
		"gamma reached Succeeded from the watch, having seen 3 of the group in the cache",
		"reconciles ",
	} {
		if !strings.Contains(printed, want) {
			t.Fatalf("output must contain %q, got:\n%s", want, printed)
		}
	}
}
