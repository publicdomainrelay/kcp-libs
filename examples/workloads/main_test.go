package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunExecutesAndSimulatesWorkloads(t *testing.T) {
	var out bytes.Buffer
	if err := Run(context.Background(), &out); err != nil {
		t.Fatalf("Run: %v\n%s", err, out.String())
	}
	printed := out.String()
	for _, want := range []string{
		"materialised the stand-in runtime at bin",
		"permissions became --allow-net=deno.land --deny-env",
		"the process started with argv run --allow-net=deno.land --deny-env main.ts",
		"finished succeeded with answer 42",
		"the readiness probe passed: true",
		"finished failed with exit code 7",
		"the liveness tracker asked for a restart after 3 failed probes",
		"the job allocated [batch-1-1 batch-1-2 batch-1-3] and still owes [batch-1-2 batch-1-3]",
		"the in-memory runner produced the same answer 42 without a process",
	} {
		if !strings.Contains(printed, want) {
			t.Fatalf("output must contain %q, got:\n%s", want, printed)
		}
	}
}
