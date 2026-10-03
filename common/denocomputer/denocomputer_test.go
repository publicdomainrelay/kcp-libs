package denocomputer

import "testing"

func TestTheVocabularyAddressesRealObjects(t *testing.T) {
	for _, value := range []struct {
		name string

		got string

		want string
	}{
		{"Group", Group, "deno.computer"},
		{"APIVersion", APIVersion, "deno.computer/v1alpha1"},
		{"PolicyWorkflowPodLabel", PolicyWorkflowPodLabel, "deno.computer/policyworkflowpod"},
		{"JobRunLabel", JobRunLabel, "deno.computer/job"},
		{"TriggerLabel", TriggerLabel, "deno.computer/trigger"},
		{"FinalizerDenoRun", FinalizerDenoRun, "denorun.deno.computer/run"},
		{"FinalizerDenoPod", FinalizerDenoPod, "denopod.deno.computer/run"},
		{"FinalizerOpenBao", FinalizerOpenBao, "openbao.deno.computer/namespace"},
		{"FinalizerPolicyWorkflowRun", FinalizerPolicyWorkflowRun, "policyworkflowrun.deno.computer/run"},
	} {
		if value.got != value.want {
			t.Fatalf("%s = %q, want %q", value.name, value.got, value.want)
		}
	}
}

func TestThePhasesAndConditionsAreTheOnesTheApiServerStores(t *testing.T) {
	if string(PhaseSucceeded) != "Succeeded" || string(PhaseCancelled) != "Cancelled" {
		t.Fatalf("phases = %q, %q", PhaseSucceeded, PhaseCancelled)
	}
	if ConditionReady != "Ready" || ConditionComplete != "Complete" {
		t.Fatalf("conditions = %q, %q", ConditionReady, ConditionComplete)
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
