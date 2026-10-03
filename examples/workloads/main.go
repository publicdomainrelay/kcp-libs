package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/publicdomainrelay/kcp-libs/abc/joballoc"
	"github.com/publicdomainrelay/kcp-libs/abc/probe"
	"github.com/publicdomainrelay/kcp-libs/abc/runner"
	"github.com/publicdomainrelay/kcp-libs/common/denospec"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/common/ttl"
	"github.com/publicdomainrelay/kcp-libs/impl/assets"
	"github.com/publicdomainrelay/kcp-libs/impl/execrunner"
	"github.com/publicdomainrelay/kcp-libs/impl/memoryrunner"
)

const (
	greetingScript = `printf '{"answer":"42","greeting":"hello"}\n' > result.json`

	standin = "#!/bin/sh\nprintf '%s\\n' \"$@\" > argv.txt\nexec /bin/sh main.ts\n"
)

func Run(ctx context.Context, out io.Writer) error {
	dir, err := os.MkdirTemp("", "workloads-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	bin, err := (&assets.Set{Dir: filepath.Join(dir, "bin"), Perm: 0o755, Files: map[string][]byte{"deno": []byte(standin)}}).Path("deno")
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "materialised the stand-in runtime at %s\n", filepath.Base(filepath.Dir(bin)))

	pod, err := execrunner.NewPod(execrunner.PodOptions{DenoBin: bin, RunsDir: filepath.Join(dir, "runs"), Timeout: 5 * time.Second})
	if err != nil {
		return err
	}
	var _ runner.PodRunner = pod
	permissions := &denospec.Permissions{
		Net: &denospec.Permission{AllowList: []string{"deno.land"}},
		Env: &denospec.Permission{Deny: true},
	}
	args, err := denospec.Args(permissions)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "permissions became %s\n", strings.Join(args, " "))

	id, err := pod.Start(ctx, runner.PodRequest{
		Name:           "greeter",
		LogicalCluster: "root:alice",
		Script:         greetingScript,
		PermissionArgs: args,
		Env:            map[string]string{"GREETING": "hello"},
		Server:         "https://kcp.example/clusters/root:alice",
		Workspace:      "root:alice",
		Token:          "service-account-token",
	})
	if err != nil {
		return err
	}
	argv, err := readWhenWritten(filepath.Join(dir, "runs", id, "argv.txt"))
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "the process started with argv %s\n", strings.Join(strings.Fields(string(argv)), " "))

	status, err := wait(ctx, pod, id)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "run %s finished %s with answer %s\n", id, status.State, status.Outputs["answer"])

	ready, err := pod.Probe(ctx, id, []string{"/bin/sh", "-c", "test -f result.json"}, time.Second)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "the readiness probe passed: %v\n", ready)

	failing, err := pod.Start(ctx, runner.PodRequest{Name: "failing", Script: "exit 7"})
	if err != nil {
		return err
	}
	status, err = wait(ctx, pod, failing)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "run %s finished %s with exit code %d: %s\n", failing, status.State, status.ExitCode, status.Message)

	liveness := probe.NewTracker()
	failures := 0
	for i := range 3 {
		if liveness.Record("greeter", id, false, probe.DefaultFailureThreshold) {
			failures = i + 1
			break
		}
	}
	fmt.Fprintf(out, "the liveness tracker asked for a restart after %d failed probes\n", failures)

	completion := metav1.NewTime(time.Now().Add(-10 * time.Second))
	remove, after, known := ttl.Expired(&completion, ptr(int64(30)), time.Now())
	fmt.Fprintf(out, "retention: delete %v, %s left of the ttl (known %v)\n", remove, after.Round(time.Second), known)

	allocator := joballoc.New(time.Minute)
	job := ref.New("root:alice", "default", "batch-1")
	allocator.Allocate(job, []string{"batch-1-1", "batch-1-2", "batch-1-3"}, time.Now())
	now := time.Now()
	fmt.Fprintf(out, "the job allocated %v and still owes %v\n", allocator.Names(job, now), allocator.Pending(job, []string{"batch-1-1"}, now))

	simulated := memoryrunner.NewPod(memoryrunner.PodOptions{PollsBeforeDone: 1, Outcome: runner.PodStatus{State: runner.StateSucceeded, Outputs: map[string]string{"answer": "42"}}})
	sid, err := simulated.Start(ctx, runner.PodRequest{Name: "greeter", Script: greetingScript})
	if err != nil {
		return err
	}
	simulatedStatus, err := simulated.Observe(ctx, sid)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "the in-memory runner produced the same answer %s without a process\n", simulatedStatus.Outputs["answer"])
	return nil
}

func wait(ctx context.Context, pod *execrunner.Pod, id string) (runner.PodStatus, error) {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		status, err := pod.Observe(ctx, id)
		if err != nil {
			return runner.PodStatus{}, err
		}
		if status.State != runner.StateRunning {
			return status, nil
		}
		time.Sleep(5 * time.Millisecond)
	}
	return runner.PodStatus{}, fmt.Errorf("example: run %s never finished", id)
}

func readWhenWritten(path string) ([]byte, error) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		body, err := os.ReadFile(path)
		if err == nil {
			return body, nil
		}
		time.Sleep(5 * time.Millisecond)
	}
	return nil, fmt.Errorf("example: %s was never written", filepath.Base(path))
}

func ptr[T any](value T) *T {
	return &value
}

func main() {
	if err := Run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "example-workloads:", err)
		os.Exit(1)
	}
}
