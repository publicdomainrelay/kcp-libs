package livekcp

import (
	"context"
	"testing"
	"time"
)

func TestLiveStartupBreakdown(t *testing.T) {
	Require(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	started := time.Now()
	cluster, err := Start(ctx)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	defer cluster.Stop()
	t.Logf("\n%s%8.2fs  Start() wall clock\n", cluster.Trace(), time.Since(started).Seconds())
}
