package deno

const (
	Group = "deno.computer"

	Version = "v1alpha1"

	APIVersion = Group + "/" + Version
)

const (
	PolicyWorkflowPodLabel = "deno.computer/policyworkflowpod"

	JobRunLabel = "deno.computer/job"

	TriggerLabel = "deno.computer/trigger"
)

const (
	FinalizerDenoRun = "denorun.deno.computer/run"

	FinalizerDenoPod = "denopod.deno.computer/run"

	FinalizerDenoJob = "denojob.deno.computer/run"

	FinalizerPolicyEngine = "policyengine.deno.computer/run"

	FinalizerPolicyWorkflowRun = "policyworkflowrun.deno.computer/run"

	FinalizerPolicyWorkflowPod = "policyworkflowpod.deno.computer/run"

	FinalizerRunTrigger = "runtrigger.deno.computer/run"

	FinalizerOpenBao = "openbao.deno.computer/namespace"
)

const (
	ConditionReady = "Ready"

	ConditionComplete = "Complete"

	ConditionFailed = "Failed"

	ConditionSuspended = "Suspended"

	ConditionCancelled = "Cancelled"

	ConditionOpenBaoReady = "Ready"

	ConditionOpenBaoAmbiguous = "Ambiguous"
)

const (
	ReasonAtCapacity = "AtCapacity"

	ReasonSuperseded = "Superseded"

	ReasonPolicyWorkflowPodMissing = "PolicyWorkflowPodMissing"

	ReasonEngineNotReady = "EngineNotReady"

	ReasonQueued = "Queued"
)

type ConcurrencyPolicy string

const (
	ConcurrencyAllow ConcurrencyPolicy = "Allow"

	ConcurrencyForbid ConcurrencyPolicy = "Forbid"

	ConcurrencyReplace ConcurrencyPolicy = "Replace"
)

func EffectiveConcurrencyPolicy(policy ConcurrencyPolicy) ConcurrencyPolicy {
	if policy == "" {
		return ConcurrencyForbid
	}
	return policy
}

type Phase string

const (
	PhasePending Phase = "Pending"

	PhaseRunning Phase = "Running"

	PhaseSucceeded Phase = "Succeeded"

	PhaseFailed Phase = "Failed"

	PhaseCancelled Phase = "Cancelled"

	PhaseTriggered Phase = "Triggered"

	PhaseSkipped Phase = "Skipped"

	PhaseReady Phase = "Ready"
)

func TerminalDenoRun(phase string) bool {
	return phase == string(PhaseSucceeded) || phase == string(PhaseFailed)
}

func TerminalDenoPod(phase string) bool {
	return phase == string(PhaseSucceeded) || phase == string(PhaseFailed)
}

func TerminalDenoJob(phase string) bool {
	return phase == string(PhaseSucceeded) || phase == string(PhaseFailed)
}

func TerminalPolicyWorkflow(phase string) bool {
	return phase == string(PhaseSucceeded) || phase == string(PhaseFailed) || phase == string(PhaseCancelled)
}

func TerminalPolicyEngine(phase string) bool {
	return phase == string(PhaseFailed)
}

func RunningPhase(phase string) bool {
	return phase == string(PhaseRunning)
}
