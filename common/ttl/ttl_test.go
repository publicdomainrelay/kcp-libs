package ttl

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func int64Ptr(v int64) *int64 { return &v }

func TestEffectivePrefersOverrideAndDropsNegatives(t *testing.T) {
	if got := Effective(int64Ptr(30), int64Ptr(60)); got == nil || *got != 30 {
		t.Fatalf("override = %v", got)
	}
	if got := Effective(nil, int64Ptr(60)); got == nil || *got != 60 {
		t.Fatalf("fallback = %v", got)
	}
	if got := Effective(nil, int64Ptr(-1)); got != nil {
		t.Fatalf("a negative ttl disables retention, got %v", *got)
	}
	if got := Effective(nil, nil); got != nil {
		t.Fatalf("no value = %v", got)
	}
}

func TestExpired(t *testing.T) {
	completed := metav1.NewTime(time.Unix(1000, 0))
	now := time.Unix(1030, 0)
	deleteNow, after, known := Expired(&completed, int64Ptr(60), now)
	if !known || deleteNow || after != 30*time.Second {
		t.Fatalf("Expired = (%v, %v, %v)", deleteNow, after, known)
	}
	deleteNow, _, known = Expired(&completed, int64Ptr(10), now)
	if !known || !deleteNow {
		t.Fatal("a past expiry must delete")
	}
	if _, _, known := Expired(nil, int64Ptr(10), now); known {
		t.Fatal("no completion time means nothing to schedule")
	}
}

func TestDeadline(t *testing.T) {
	started := metav1.NewTime(time.Unix(1000, 0))
	if Deadline(&started, int64Ptr(60), time.Unix(1059, 0)) {
		t.Fatal("a deadline in the future is not exceeded")
	}
	if !Deadline(&started, int64Ptr(60), time.Unix(1060, 0)) {
		t.Fatal("a reached deadline is exceeded")
	}
	if Deadline(nil, int64Ptr(60), time.Unix(1060, 0)) {
		t.Fatal("no start time means no deadline")
	}
}
