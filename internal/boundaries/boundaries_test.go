package boundaries

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

const module = "github.com/publicdomainrelay/kcp-libs/"

type listEntry struct {
	ImportPath string

	Imports []string

	TestImports []string

	Name string
}

func packages(t *testing.T) []listEntry {
	t.Helper()
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = "../.."
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	var entries []listEntry
	decoder := json.NewDecoder(strings.NewReader(string(out)))
	for decoder.More() {
		var entry listEntry
		if err := decoder.Decode(&entry); err != nil {
			t.Fatalf("decode go list output: %v", err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func TestTestSupportIsNotImportedByProductionCode(t *testing.T) {
	for _, entry := range packages(t) {
		if strings.HasPrefix(entry.ImportPath, module+"examples/") ||
			strings.HasPrefix(entry.ImportPath, module+"internal/") {
			continue
		}
		for _, imported := range entry.Imports {
			if strings.HasPrefix(imported, module+"internal/") {
				t.Fatalf("%s imports %s; test support belongs in tests", entry.ImportPath, imported)
			}
		}
	}
}

func TestLayerDependenciesFlowOneWay(t *testing.T) {
	for _, entry := range packages(t) {
		if strings.HasSuffix(entry.ImportPath, "/internal/boundaries") {
			continue
		}
		layer := layerOf(entry.ImportPath)
		if layer == "" {
			t.Fatalf("%s is not in a recognised layer", entry.ImportPath)
		}
		for _, imported := range append(append([]string(nil), entry.Imports...), entry.TestImports...) {
			if !strings.HasPrefix(imported, module) {
				continue
			}
			importedLayer := layerOf(imported)
			if importedLayer == "" {
				t.Fatalf("%s imports %s, which is not in a recognised layer", entry.ImportPath, imported)
			}
			if layer == "internal" || importedLayer == "internal" {
				continue
			}
			if rank(importedLayer) >= rank(layer) {
				t.Fatalf("%s (%s) imports %s (%s); dependencies must flow common <- abc <- impl <- factory <- examples",
					entry.ImportPath, layer, imported, importedLayer)
			}
		}
	}
}

func layerOf(path string) string {
	rest := strings.TrimPrefix(path, module)
	if rest == path || rest == "" {
		return ""
	}
	segment := rest
	if i := strings.Index(rest, "/"); i >= 0 {
		segment = rest[:i]
	}
	switch segment {
	case "common", "abc", "impl", "factory", "examples", "cmd", "internal":
		return segment
	}
	return ""
}

func rank(layer string) int {
	switch layer {
	case "common":
		return 0
	case "abc":
		return 1
	case "impl":
		return 2
	case "factory":
		return 3
	case "examples":
		return 4
	case "cmd":
		return 4
	case "internal":
		return 5
	}
	return -1
}

func TestCommonDoesNotImportProjectLocalPackages(t *testing.T) {
	for _, entry := range packages(t) {
		if !strings.HasPrefix(entry.ImportPath, module+"common/") {
			continue
		}
		for _, imported := range entry.Imports {
			if strings.HasPrefix(imported, module) {
				t.Fatalf("%s imports %s; the common layer must be leaf packages", entry.ImportPath, imported)
			}
		}
	}
}

func TestAbcImportsOnlyCommon(t *testing.T) {
	for _, entry := range packages(t) {
		if !strings.HasPrefix(entry.ImportPath, module+"abc/") {
			continue
		}
		for _, imported := range entry.Imports {
			if strings.HasPrefix(imported, module) && !strings.HasPrefix(imported, module+"common/") {
				t.Fatalf("%s imports %s; the abc layer may import common only", entry.ImportPath, imported)
			}
		}
	}
}
