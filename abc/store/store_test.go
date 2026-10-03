package store

import (
	"testing"
	"time"
)

type status struct {
	Phase string `json:"phase"`

	Active int32 `json:"active"`

	Times []string `json:"times,omitempty"`
}

func TestSameComparesWireShape(t *testing.T) {
	first := status{Phase: "Running", Active: 1}
	second := status{Phase: "Running", Active: 1}
	if !Same(first, second) {
		t.Fatal("identical statuses must compare equal")
	}
	if Same(first, status{Phase: "Running", Active: 2}) {
		t.Fatal("a changed count is a change")
	}
	if Same(first, status{Phase: "Succeeded"}) {
		t.Fatal("a changed phase is a change")
	}
}

func TestSameIsNotFooledByOrderingOrNilSlices(t *testing.T) {
	if Same(status{Phase: "Running", Times: []string{"a"}}, status{Phase: "Running", Times: nil}) {
		t.Fatal("a cleared list is a change")
	}
	if !Same(map[string]any{"a": 1, "b": 2}, map[string]any{"b": 2, "a": 1}) {
		t.Fatal("map key order must not matter")
	}
	if Same(1, "1") {
		t.Fatal("different types are different")
	}
}

func TestSameHandlesUnencodableValues(t *testing.T) {
	if Same(make(chan int), make(chan int)) {
		t.Fatal("a value that cannot be encoded is never equal")
	}
}

func TestTimeIsComparedByValue(t *testing.T) {
	now := time.Unix(1000, 0).UTC()
	if !Same(map[string]any{"t": now}, map[string]any{"t": now}) {
		t.Fatal("the same instant must compare equal")
	}
}
