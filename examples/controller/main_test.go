package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunDrivesWidgetsToCompletion(t *testing.T) {
	var out bytes.Buffer
	if err := Run(context.Background(), &out); err != nil {
		t.Fatalf("Run: %v\n%s", err, out.String())
	}
	printed := out.String()
	for _, want := range []string{
		"discovered denoruntime at ",
		"alpha reached Succeeded after 1 passes",
		"beta reached Succeeded after 1 passes",
		"gamma reached Succeeded from the watch, seeing 3 siblings in the cache",
		"queue depth 0",
		"example_reconcile_seconds_count",
	} {
		if !strings.Contains(printed, want) {
			t.Fatalf("output must contain %q, got:\n%s", want, printed)
		}
	}
}
