package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunBuildsTheTableAndTheTokens(t *testing.T) {
	var out bytes.Buffer
	if err := Run(context.Background(), &out); err != nil {
		t.Fatalf("Run: %v\n%s", err, out.String())
	}
	printed := out.String()
	for _, want := range []string{
		"api.default.alice.svc.kcp.local resolves to 127.0.0.1:3000",
		"pds.default.alice.svc.kcp.local resolves to 127.0.0.1:8080",
		"pds.default.bob.svc.kcp.local resolves to 127.0.0.1:9000",
		"the table covers 2 workspaces",
		"the pod carries 3 names and a token for [2j35eh7jjhsc8ny9 8fj2hd8jjdhs09ab]",
		"a service that binds 0.0.0.0 advertises 127.0.0.1:3000",
		"the workspace path reads as labels: root:alice becomes alice",
	} {
		if !strings.Contains(printed, want) {
			t.Fatalf("output must contain %q, got:\n%s", want, printed)
		}
	}
}
