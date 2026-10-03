package fakekcp

import "time"

func Object(apiVersion, kind, namespace, name string) map[string]any {
	metadata := map[string]any{"name": name}
	if namespace != "" {
		metadata["namespace"] = namespace
	}
	return map[string]any{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata":   metadata,
		"spec":       map[string]any{},
		"status":     map[string]any{},
	}
}

func WithLabel(body map[string]any, key, value string) map[string]any {
	ensureMap(ensureMap(body, "metadata"), "labels")[key] = value
	return body
}

func WithAnnotation(body map[string]any, key, value string) map[string]any {
	ensureMap(ensureMap(body, "metadata"), "annotations")[key] = value
	return body
}

func WithSpec(body map[string]any, values map[string]any) map[string]any {
	spec := ensureMap(body, "spec")
	for key, value := range values {
		spec[key] = value
	}
	return body
}

func WithCreated(body map[string]any, at time.Time) map[string]any {
	ensureMap(body, "metadata")["creationTimestamp"] = at.UTC().Format(time.RFC3339Nano)
	return body
}

func NameOf(body map[string]any) string {
	metadata, ok := body["metadata"].(map[string]any)
	if !ok {
		return ""
	}
	name, _ := metadata["name"].(string)
	return name
}

func PhaseOf(body map[string]any) string {
	status, ok := body["status"].(map[string]any)
	if !ok {
		return ""
	}
	phase, _ := status["phase"].(string)
	return phase
}

func StatusOf(body map[string]any) map[string]any {
	status, ok := body["status"].(map[string]any)
	if !ok {
		return nil
	}
	return status
}

func watchFrame(event string, object map[string]any) map[string]any {
	return map[string]any{"type": event, "object": object}
}

func bookmark(apiVersion, kind, resourceVersion string) map[string]any {
	return map[string]any{
		"type": "BOOKMARK",
		"object": map[string]any{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]any{
				"resourceVersion": resourceVersion,
				"annotations":     map[string]any{"k8s.io/initial-events-end": "true"},
			},
		},
	}
}

func ensureMap(body map[string]any, key string) map[string]any {
	if existing, ok := body[key].(map[string]any); ok {
		return existing
	}
	created := map[string]any{}
	body[key] = created
	return created
}

func clone(body map[string]any) map[string]any {
	out := make(map[string]any, len(body))
	for key, value := range body {
		out[key] = cloneValue(value)
	}
	return out
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return clone(typed)
	case []any:
		out := make([]any, len(typed))
		for i, entry := range typed {
			out[i] = cloneValue(entry)
		}
		return out
	default:
		return value
	}
}

func stripAnnotations(body map[string]any) map[string]any {
	out := clone(body)
	metadata, ok := out["metadata"].(map[string]any)
	if !ok {
		return out
	}
	if annotations, ok := metadata["annotations"].(map[string]any); ok {
		delete(annotations, ClusterAnnotation)
		if len(annotations) == 0 {
			delete(metadata, "annotations")
		}
	}
	return out
}
