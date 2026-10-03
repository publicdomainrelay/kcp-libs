package kcpstore

import (
	"encoding/json"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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

func IsAlreadyExists(err error) bool {
	return apierrors.IsAlreadyExists(err)
}

func StatusPatch(status map[string]any) ([]byte, error) {
	body, err := json.Marshal(map[string]any{"status": status})
	if err != nil {
		return nil, fmt.Errorf("kcpstore: encode status: %w", err)
	}
	return body, nil
}

func removeFinalizerPatch(current []string, dropped string) ([]byte, error) {
	remaining := make([]string, 0, len(current))
	for _, finalizer := range current {
		if finalizer != dropped {
			remaining = append(remaining, finalizer)
		}
	}
	if len(remaining) == len(current) {
		return nil, nil
	}
	return json.Marshal([]map[string]any{
		{"op": "test", "path": "/metadata/finalizers", "value": current},
		{"op": "add", "path": "/metadata/finalizers", "value": remaining},
	})
}

func addFinalizerPatch(finalizers []string) ([]byte, error) {
	return json.Marshal([]map[string]any{
		{"op": "add", "path": "/metadata/finalizers", "value": finalizers},
	})
}
