package kcpstore

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/publicdomainrelay/kcp-libs/common/statuspatch"
)

var (
	mergePatch = types.MergePatchType

	jsonPatch = types.JSONPatchType
)

func IsNotFound(err error) bool {
	return apierrors.IsNotFound(err)
}

func IsConflict(err error) bool {
	return apierrors.IsConflict(err)
}

func removeFinalizerPatch(current []string, dropped string) ([]byte, error) {
	return statuspatch.FinalizerRemove(current, dropped)
}

func addFinalizerPatch(finalizers []string) ([]byte, error) {
	return statuspatch.FinalizerAdd(finalizers)
}
