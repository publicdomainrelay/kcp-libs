package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunProvisionsAndIssues(t *testing.T) {
	var out bytes.Buffer
	if err := Run(context.Background(), &out); err != nil {
		t.Fatalf("Run: %v\n%s", err, out.String())
	}
	printed := out.String()
	for _, want := range []string{
		"vault reachable, sealed false",
		"intermediate alice.default.intermediate serial ",
		"issued pds.default.alice.svc.kcp.local serial EE:01",
		"the workload serves 3 certificates: its leaf, the namespace intermediate, and the root",
		"a second leaf cost 1 calls, the authority was cached",
		"provisioning again after a delete cost 6 calls, the namespace was gone",
	} {
		if !strings.Contains(printed, want) {
			t.Fatalf("output must contain %q, got:\n%s", want, printed)
		}
	}
	if strings.Contains(printed, "serial \n") {
		t.Fatalf("the serial must be read out of the certificate, got:\n%s", printed)
	}
}
