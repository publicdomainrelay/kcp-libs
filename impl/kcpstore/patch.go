package kcpstore

import (
	"fmt"

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

func StatusPatch(status map[string]any) ([]byte, error) {
	body, err := statuspatch.Merge(status)
	if err != nil {
		return nil, fmt.Errorf("kcpstore: %w", err)
	}
	return body, nil
}

func statusBody(patch []byte, resourceVersion string) ([]byte, error) {
	return statuspatch.WithResourceVersion(patch, resourceVersion)
}

func removeFinalizerPatch(current []string, dropped string) ([]byte, error) {
	return statuspatch.FinalizerRemove(current, dropped)
}

func addFinalizerPatch(finalizers []string) ([]byte, error) {
	return statuspatch.FinalizerAdd(finalizers)
}
