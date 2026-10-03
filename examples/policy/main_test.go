package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunFollowsTasksToTheirVerdict(t *testing.T) {
	var out bytes.Buffer
	if err := Run(context.Background(), &out); err != nil {
		t.Fatalf("Run: %v\n%s", err, out.String())
	}
	printed := out.String()
	for _, want := range []string{
		"submitted task-1",
		"the verdict was allow=true with violations []",
		"the engine received 1 submissions",
		"a failed workflow reports failed: the workflow exited with status failure",
		"an unreachable engine leaves the task running rather than failing it",
	} {
		if !strings.Contains(printed, want) {
			t.Fatalf("output must contain %q, got:\n%s", want, printed)
		}
	}
}
