package denospec

import (
	"reflect"
	"strings"
	"testing"
)

func TestArgsAll(t *testing.T) {
	args, err := Args(&Permissions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []string{"-A"}) {
		t.Fatalf("args = %v", args)
	}
}

func TestArgsCapabilities(t *testing.T) {
	args, err := Args(&Permissions{
		Read:         &Permission{Allow: true},
		Net:          &Permission{AllowList: []string{"example.com", "other.example"}},
		Env:          &Permission{DenyList: []string{"SECRET"}},
		Sys:          &Permission{Deny: true},
		HRTime:       true,
		IgnoreEnv:    &Permission{AllowList: []string{"PUBLIC"}},
		AllowScripts: []string{"npm"},
		NoPrompt:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, want := range []string{
		"--allow-hrtime",
		"--allow-read",
		"--allow-net=example.com,other.example",
		"--deny-env=SECRET",
		"--deny-sys",
		"--ignore-env=PUBLIC",
		"--allow-scripts=npm",
		"--no-prompt",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args %q must contain %q", joined, want)
		}
	}
}

func TestValidateRejectsCommasAndEmptyValues(t *testing.T) {
	if err := Validate(&Permissions{Net: &Permission{AllowList: []string{"a,b"}}}); err == nil {
		t.Fatal("a comma in a value must be refused")
	}
	if err := Validate(&Permissions{Net: &Permission{AllowList: []string{""}}}); err == nil {
		t.Fatal("an empty value must be refused")
	}
	if err := Validate(&Permissions{IgnoreEnv: &Permission{Allow: true}}); err == nil {
		t.Fatal("ignoreEnv must take only an allowList")
	}
	if err := Validate(nil); err != nil {
		t.Fatal(err)
	}
}

func TestEffectiveRestartPolicy(t *testing.T) {
	if EffectiveRestartPolicy("") != RestartAlways {
		t.Fatal("an unset restart policy is Always")
	}
	if EffectiveRestartPolicy(RestartNever) != RestartNever {
		t.Fatal("an explicit policy is preserved")
	}
}
