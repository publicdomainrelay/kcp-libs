package driver

import (
	"testing"
	"time"
)

func TestPolicyNextDefaultsTheInterval(t *testing.T) {
	policy := Policy{Interval: 2 * time.Second, MinTransitionPoll: 25 * time.Millisecond}
	after, requeue := policy.Next(Key{Kind: "job"}, 0, false)
	if !requeue || after != 2*time.Second {
		t.Fatalf("next = (%v, %v), want (2s, true)", after, requeue)
	}
}

func TestPolicyNextStopsOnTerminalWithoutRequeue(t *testing.T) {
	policy := Policy{Interval: 2 * time.Second, MinTransitionPoll: 25 * time.Millisecond}
	if _, requeue := policy.Next(Key{Kind: "job"}, 0, true); requeue {
		t.Fatal("a terminal key with no requeue must not be requeued")
	}
	after, requeue := policy.Next(Key{Kind: "job"}, time.Hour, true)
	if !requeue || after != time.Hour {
		t.Fatalf("next = (%v, %v), want (1h, true)", after, requeue)
	}
}

func TestPolicyNextClampsOnlyClampedKinds(t *testing.T) {
	policy := Policy{
		Interval:          2 * time.Second,
		MinTransitionPoll: 25 * time.Millisecond,
		ClampKinds:        map[string]bool{"denorun": true},
	}
	after, _ := policy.Next(Key{Kind: "denorun"}, 2*time.Second, false)
	if after != 25*time.Millisecond {
		t.Fatalf("clamped after = %v, want 25ms", after)
	}
	after, _ = policy.Next(Key{Kind: "denojob"}, 2*time.Second, false)
	if after != 2*time.Second {
		t.Fatalf("unclamped after = %v, want 2s", after)
	}
}

func TestPolicyZeroValueDefaults(t *testing.T) {
	var policy Policy
	if policy.Default() != DefaultRequeueAfter {
		t.Fatalf("default = %v", policy.Default())
	}
	if policy.Clamp() != DefaultMinTransitionPoll {
		t.Fatalf("clamp = %v", policy.Clamp())
	}
	if delay := policy.ConflictAfter(Key{Kind: "denorun"}); delay != DefaultRequeueAfter {
		t.Fatalf("conflict delay = %v", delay)
	}
}

func TestKeyString(t *testing.T) {
	key := Key{Kind: "denorun"}
	key.Ref.LogicalCluster = "root:alice"
	key.Ref.Namespace = "default"
	key.Ref.Name = "run-1"
	if key.String() != "denorun/root:alice/default/run-1" {
		t.Fatalf("key = %q", key.String())
	}
}
