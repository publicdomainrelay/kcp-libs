package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/publicdomainrelay/kcp-libs/internal/livekcp/livetest"
)

func TestRunBuildsTheTableAndTheTokens(t *testing.T) {
	livetest.Require(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var out bytes.Buffer
	if err := Run(ctx, &out); err != nil {
		t.Fatalf("Run: %v\n%s", err, out.String())
	}
	printed := out.String()
	for _, want := range []string{
		"api.default.consumer.svc.kcp.local resolves to 127.0.0.1:3000",
		"pds.default.consumer-two.svc.kcp.local resolves to 127.0.0.1:9000",
		"pds.default.consumer.svc.kcp.local resolves to 127.0.0.1:8080",
		"the table covers 2 workspaces",
		"the pod carries 3 names and a token for 2 workspaces",
		"its own name is in the table from its first moment: 127.0.0.1:8080",
		"kcp minted a token for root:consumer: true",
		"the shim and probe are written under .kcpdns: shim.ts and probe.ts",
		"a readiness probe preloads shim.ts and is granted KCP_SERVICE_DOMAIN,KCP_DNS_TABLE,KCP_TOKENS,KCP_SERVER",
		"a service that binds 0.0.0.0 advertises 127.0.0.1:3000",
		"the workspace path reads as labels: root:consumer becomes consumer",
	} {
		if !strings.Contains(printed, want) {
			t.Fatalf("output must contain %q, got:\n%s", want, printed)
		}
	}
}
