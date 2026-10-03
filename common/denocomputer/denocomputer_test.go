package denocomputer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var assignment = regexp.MustCompile(`^\s*(\w+)(?:\s+\w+)?\s*=\s*"([^"]+)"`)

func consumerConstants(t *testing.T) map[string]string {
	t.Helper()
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for {
		candidate := filepath.Join(root, "deno-kcp", "api", "v1alpha1")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return readConstants(t, candidate)
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
	return nil
}

func readConstants(t *testing.T, dir string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(body), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			match := assignment.FindStringSubmatch(strings.TrimPrefix(line, "const "))
			if match == nil {
				continue
			}
			out[match[1]] = match[2]
		}
	}
	return out
}

func TestTheGroupIsTheOneTheConsumerDeclares(t *testing.T) {
	declared := consumerConstants(t)
	want, ok := declared["GroupName"]
	if !ok {
		t.Fatal("the consumer declares no GroupName, so the group this module uses cannot be checked against it")
	}
	if Group != want {
		t.Fatalf("Group = %q, the consumer declares %q", Group, want)
	}
	version, ok := declared["Version"]
	if !ok {
		t.Fatal("the consumer declares no Version")
	}
	if Version != version {
		t.Fatalf("Version = %q, the consumer declares %q", Version, version)
	}
}

func TestEveryFinalizerIsOneTheConsumerDeclares(t *testing.T) {
	declared := consumerConstants(t)
	for name, value := range map[string]string{
		"FinalizerDenoRun":           FinalizerDenoRun,
		"FinalizerDenoPod":           FinalizerDenoPod,
		"FinalizerPolicyEngine":      FinalizerPolicyEngine,
		"FinalizerPolicyWorkflowRun": FinalizerPolicyWorkflowRun,
		"FinalizerOpenBao":           FinalizerOpenBao,
	} {
		want, ok := declared[name]
		if !ok {
			t.Fatalf("the consumer declares no %s, so this one is invented", name)
		}
		if value != want {
			t.Fatalf("%s = %q, the consumer declares %q", name, value, want)
		}
	}
}

func TestTheLabelIsTheOneTheConsumerDeclares(t *testing.T) {
	declared := consumerConstants(t)
	if want := declared["PolicyWorkflowPodLabel"]; want != "" && PolicyWorkflowPodLabel != want {
		t.Fatalf("PolicyWorkflowPodLabel = %q, the consumer declares %q", PolicyWorkflowPodLabel, want)
	}
	for _, value := range []string{JobRunLabel, TriggerLabel} {
		if !strings.HasPrefix(value, Group+"/") {
			t.Fatalf("label %q is not in the %s group", value, Group)
		}
	}
}

func TestThePhasesAndConditionsAreTheOnesTheConsumerStores(t *testing.T) {
	declared := consumerConstants(t)
	for name, value := range map[string]string{
		"DenoRunPending":          string(PhasePending),
		"DenoRunRunning":          string(PhaseRunning),
		"DenoRunSucceeded":        string(PhaseSucceeded),
		"DenoRunFailed":           string(PhaseFailed),
		"PolicyWorkflowCancelled": string(PhaseCancelled),
		"RunTriggerTriggered":     string(PhaseTriggered),
		"ConditionReady":          ConditionReady,
		"ConditionComplete":       ConditionComplete,
		"ConditionFailed":         ConditionFailed,
		"ConditionSuspended":      ConditionSuspended,
		"ConditionCancelled":      ConditionCancelled,
	} {
		want, ok := declared[name]
		if !ok {
			t.Fatalf("the consumer declares no %s, so this module's value for it is invented", name)
		}
		if value != want {
			t.Fatalf("%s = %q, the consumer declares %q", name, value, want)
		}
	}
}

func TestTheTerminalPredicatesSeparateTheKinds(t *testing.T) {
	if !TerminalDenoRun(string(PhaseSucceeded)) || TerminalDenoRun(string(PhaseRunning)) {
		t.Fatal("a deno run is terminal only when it has finished")
	}
	if !TerminalPolicyWorkflow(string(PhaseCancelled)) {
		t.Fatal("a cancelled policy run is terminal")
	}
	if TerminalPolicyEngine(string(PhaseRunning)) || !TerminalPolicyEngine(string(PhaseFailed)) {
		t.Fatal("a policy engine is terminal only when it has failed")
	}
}
