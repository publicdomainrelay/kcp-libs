package condition

import (
	"slices"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Copy(in []metav1.Condition) []metav1.Condition {
	return slices.Clone(in)
}

func Set(list *[]metav1.Condition, generation int64, status metav1.ConditionStatus, kind, reason, message string) {
	meta.SetStatusCondition(list, metav1.Condition{
		Type:               kind,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: generation,
	})
}

func SetTrue(list *[]metav1.Condition, generation int64, kind, reason, message string) {
	Set(list, generation, metav1.ConditionTrue, kind, reason, message)
}

func SetFalse(list *[]metav1.Condition, generation int64, kind, reason, message string) {
	Set(list, generation, metav1.ConditionFalse, kind, reason, message)
}

func Remove(list *[]metav1.Condition, kind string) {
	meta.RemoveStatusCondition(list, kind)
}

func Of(list []metav1.Condition, kind string) *metav1.Condition {
	return meta.FindStatusCondition(list, kind)
}

func Is(list []metav1.Condition, kind string, status metav1.ConditionStatus) bool {
	found := Of(list, kind)
	return found != nil && found.Status == status
}
