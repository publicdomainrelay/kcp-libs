package outputs

import "testing"

func TestStringify(t *testing.T) {
	if got := Stringify(nil); got != nil {
		t.Fatalf("empty input = %v", got)
	}
	got := Stringify(map[string]any{
		"text":   "plain",
		"number": 3,
		"flag":   true,
		"list":   []any{"a", "b"},
	})
	if got["text"] != "plain" {
		t.Fatalf("text = %q", got["text"])
	}
	if got["number"] != "3" {
		t.Fatalf("number = %q", got["number"])
	}
	if got["flag"] != "true" {
		t.Fatalf("flag = %q", got["flag"])
	}
	if got["list"] != `["a","b"]` {
		t.Fatalf("list = %q", got["list"])
	}
}

func TestMerge(t *testing.T) {
	merged := Merge(map[string]string{"allow": "true"}, map[string]string{"violations": "[]"})
	if len(merged) != 2 || merged["allow"] != "true" || merged["violations"] != "[]" {
		t.Fatalf("merged = %v", merged)
	}
	if Merge(nil, nil) != nil {
		t.Fatal("no inputs must merge to nothing")
	}
}
