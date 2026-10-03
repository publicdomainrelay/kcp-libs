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

	Name string
}

func TestLayerDependenciesFlowOneWay(t *testing.T) {
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = "../.."
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(out)))
	for decoder.More() {
		var entry listEntry
		if err := decoder.Decode(&entry); err != nil {
			t.Fatalf("decode go list output: %v", err)
		}
		if entry.Name == "main" || strings.HasSuffix(entry.ImportPath, "/internal/boundaries") {
			continue
		}
		layer := layerOf(entry.ImportPath)
		if layer == "" {
			t.Fatalf("%s is not in a recognised layer", entry.ImportPath)
		}
		for _, imported := range entry.Imports {
			if !strings.HasPrefix(imported, module) {
				continue
			}
			importedLayer := layerOf(imported)
			if importedLayer == "" {
				t.Fatalf("%s imports %s, which is not in a recognised layer", entry.ImportPath, imported)
			}
			if rank(importedLayer) >= rank(layer) {
				t.Fatalf("%s (%s) imports %s (%s); dependencies must flow common <- abc <- impl <- factory <- cmd",
					entry.ImportPath, layer, imported, importedLayer)
			}
		}
	}
}

func layerOf(path string) string {
	rest := strings.TrimPrefix(path, module)
	switch {
	case strings.HasPrefix(rest, "common/"), strings.HasPrefix(rest, "abc/"),
		strings.HasPrefix(rest, "impl/"), strings.HasPrefix(rest, "factory/"),
		strings.HasPrefix(rest, "cmd/"):
		return strings.SplitN(rest, "/", 2)[0]
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
	case "cmd":
		return 4
	}
	return -1
}

func TestCommonDoesNotImportProjectLocalPackages(t *testing.T) {
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = "../.."
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(out)))
	for decoder.More() {
		var entry listEntry
		if err := decoder.Decode(&entry); err != nil {
			t.Fatalf("decode go list output: %v", err)
		}
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
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = "../.."
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(out)))
	for decoder.More() {
		var entry listEntry
		if err := decoder.Decode(&entry); err != nil {
			t.Fatalf("decode go list output: %v", err)
		}
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
