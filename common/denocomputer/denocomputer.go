package denocomputer

const (
	Group = "denocomputer.computer"

	Version = "v1alpha1"

	APIVersion = Group + "/" + Version
)

const (
	PolicyWorkflowPodLabel = "denocomputer.computer/policyworkflowpod"

	JobRunLabel = "denocomputer.computer/job"

	TriggerLabel = "denocomputer.computer/trigger"
)

const (
	FinalizerDenoRun = "denorun.denocomputer.computer/run"

	FinalizerDenoPod = "denopod.denocomputer.computer/run"

	FinalizerDenoJob = "denojob.denocomputer.computer/run"

	FinalizerPolicyEngine = "policyengine.denocomputer.computer/run"

	FinalizerPolicyWorkflowRun = "policyworkflowrun.denocomputer.computer/run"

	FinalizerPolicyWorkflowPod = "policyworkflowpod.denocomputer.computer/run"

	FinalizerRunTrigger = "runtrigger.denocomputer.computer/run"

	FinalizerOpenBao = "openbao.denocomputer.computer/namespace"
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
