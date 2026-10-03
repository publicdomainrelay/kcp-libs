package denospec

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var (
	structOpen = regexp.MustCompile(`^type (\w+) struct \{`)

	jsonTag = regexp.MustCompile(`json:"([^",]+)`)
)

func TestTheShapesAreTheOnesTheConsumerDeclares(t *testing.T) {
	dir := consumerAPIDir(t)
	if dir == "" {
		return
	}
	declared := consumerStructTags(t, dir)
	for _, pair := range []struct {
		shape any
		name  string
	}{
		{ServiceAccountRef{}, "ServiceAccountRef"},
		{Permission{}, "DenoPermission"},
		{Permissions{}, "DenoPermissions"},
		{PodTemplate{}, "DenoPodTemplate"},
		{ExecProbe{}, "ExecProbe"},
	} {
		want, ok := declared[pair.name]
		if !ok {
			t.Fatalf("the consumer declares no %s, so the shape here is invented", pair.name)
		}
		got := ownTags(pair.shape)
		if !slices.Equal(got, want) {
			t.Fatalf("%s carries %v, the consumer declares %v, so a patch built from one would not decode as the other", pair.name, got, want)
		}
	}
}

func ownTags(shape any) []string {
	out := []string{}
	kind := reflect.TypeOf(shape)
	for i := 0; i < kind.NumField(); i++ {
		tag, ok := kind.Field(i).Tag.Lookup("json")
		if !ok {
			continue
		}
		out = append(out, strings.Split(tag, ",")[0])
	}
	return out
}

func consumerStructTags(t *testing.T, dir string) map[string][]string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]string{}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		current := ""
		for _, line := range strings.Split(string(body), "\n") {
			line = strings.TrimSpace(line)
			if match := structOpen.FindStringSubmatch(line); match != nil {
				current = match[1]
				out[current] = []string{}
				continue
			}
			if current == "" {
				continue
			}
			if line == "}" {
				current = ""
				continue
			}
			if strings.HasPrefix(line, "//") {
				continue
			}
			if match := jsonTag.FindStringSubmatch(line); match != nil {
				out[current] = append(out[current], match[1])
			}
		}
	}
	return out
}

func consumerAPIDir(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for {
		candidate := filepath.Join(root, "deno-kcp", "api", "v1alpha1")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}
	if os.Getenv("KCP_LIBS_REQUIRE_CONSUMER") == "1" {
		t.Fatal("the consumer's api/v1alpha1 was not found; this test reads it as the source of truth, so set KCP_LIBS_REQUIRE_CONSUMER only where the checkout is present")
	}
	t.Skip("the consumer checkout is not next to this module, so there is nothing to compare against")
	return ""
}
