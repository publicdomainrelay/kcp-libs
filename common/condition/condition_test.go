package condition

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetAndIs(t *testing.T) {
	var list []metav1.Condition
	SetTrue(&list, 3, "Ready", "Ready", "all good")
	if !Is(list, "Ready", metav1.ConditionTrue) {
		t.Fatalf("list = %v", list)
	}
	if list[0].ObservedGeneration != 3 {
		t.Fatal("the observed generation must be recorded")
	}
	SetFalse(&list, 4, "Ready", "NotReady", "waiting")
	if Is(list, "Ready", metav1.ConditionTrue) {
		t.Fatal("a set must replace the previous condition of the same type")
	}
	if len(list) != 1 {
		t.Fatalf("list = %v", list)
	}
}

func TestRemove(t *testing.T) {
	var list []metav1.Condition
	SetTrue(&list, 1, "Suspended", "Suspended", "suspended")
	Remove(&list, "Suspended")
	if Of(list, "Suspended") != nil {
		t.Fatal("the condition must be gone")
	}
}

func TestCopyIsIndependent(t *testing.T) {
	var list []metav1.Condition
	SetTrue(&list, 1, "Ready", "Ready", "ok")
	clone := Copy(list)
	SetFalse(&clone, 2, "Ready", "NotReady", "changed")
	if !Is(list, "Ready", metav1.ConditionTrue) {
		t.Fatal("the copy must not alias the original")
	}
}

func TestSetPreservesTheTransitionTimeOfAnUnchangedCondition(t *testing.T) {
	var list []metav1.Condition
	SetTrue(&list, 1, "Ready", "Ready", "ok")
	first := list[0].LastTransitionTime
	SetTrue(&list, 2, "Ready", "Ready", "still ok")
	if len(list) != 1 {
		t.Fatalf("list = %v", list)
	}
	if !list[0].LastTransitionTime.Equal(&first) {
		t.Fatalf("transition time = %v, want the original %v", list[0].LastTransitionTime, first)
	}
	if list[0].ObservedGeneration != 2 {
		t.Fatal("the generation must still move")
	}
}
