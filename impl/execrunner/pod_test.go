package execrunner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/publicdomainrelay/kcp-libs/abc/runner"
)

func writeStub(t *testing.T, dir, body string) string {
	t.Helper()
	path := filepath.Join(dir, "stub.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func observeUntilDone(t *testing.T, pod *Pod, id string) runner.PodStatus {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		status, err := pod.Observe(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		if status.State != runner.StateRunning {
			return status
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the stub never finished")
	return runner.PodStatus{}
}

func TestPodWritesOutputsAndExitCode(t *testing.T) {
	dir := t.TempDir()
	stub := writeStub(t, dir, `printf '{"answer":"42"}' > result.json`)
	pod, err := NewPod(PodOptions{DenoBin: stub, RunsDir: filepath.Join(dir, "runs")})
	if err != nil {
		t.Fatal(err)
	}
	id, err := pod.Start(context.Background(), runner.PodRequest{Name: "pds", Script: "ignored"})
	if err != nil {
		t.Fatal(err)
	}
	status := observeUntilDone(t, pod, id)
	if status.State != runner.StateSucceeded || status.ExitCode != 0 {
		t.Fatalf("status = %+v", status)
	}
	if status.Outputs["answer"] != "42" {
		t.Fatalf("outputs = %v", status.Outputs)
	}
}

func TestPodReportsANonZeroExit(t *testing.T) {
	dir := t.TempDir()
	stub := writeStub(t, dir, `exit 3`)
	pod, err := NewPod(PodOptions{DenoBin: stub, RunsDir: filepath.Join(dir, "runs")})
	if err != nil {
		t.Fatal(err)
	}
	id, err := pod.Start(context.Background(), runner.PodRequest{Name: "pds"})
	if err != nil {
		t.Fatal(err)
	}
	status := observeUntilDone(t, pod, id)
	if status.State != runner.StateFailed || status.ExitCode != 3 {
		t.Fatalf("status = %+v", status)
	}
	if status.Message == "" {
		t.Fatal("a failure must carry a message")
	}
}

func TestPodStopIsReportedAsFailure(t *testing.T) {
	dir := t.TempDir()
	stub := writeStub(t, dir, `sleep 30`)
	pod, err := NewPod(PodOptions{DenoBin: stub, RunsDir: filepath.Join(dir, "runs")})
	if err != nil {
		t.Fatal(err)
	}
	id, err := pod.Start(context.Background(), runner.PodRequest{Name: "pds"})
	if err != nil {
		t.Fatal(err)
	}
	if err := pod.Stop(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	status := observeUntilDone(t, pod, id)
	if status.State != runner.StateFailed || status.Message != "the deno process was stopped" {
		t.Fatalf("status = %+v", status)
	}
}

func TestPodRecoversARunFromItsDirectory(t *testing.T) {
	dir := t.TempDir()
	runsDir := filepath.Join(dir, "runs")
	stub := writeStub(t, dir, `printf '{"answer":"42"}' > result.json`)
	pod, err := NewPod(PodOptions{DenoBin: stub, RunsDir: runsDir})
	if err != nil {
		t.Fatal(err)
	}
	id, err := pod.Start(context.Background(), runner.PodRequest{Name: "pds"})
	if err != nil {
		t.Fatal(err)
	}
	if status := observeUntilDone(t, pod, id); status.State != runner.StateSucceeded {
		t.Fatalf("status = %+v", status)
	}

	restarted, err := NewPod(PodOptions{DenoBin: stub, RunsDir: runsDir})
	if err != nil {
		t.Fatal(err)
	}
	status, err := restarted.Observe(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if status.State != runner.StateSucceeded || status.ExitCode != 0 || status.Outputs["answer"] != "42" {
		t.Fatalf("recovered status = %+v", status)
	}
}

func TestPodProbeRunsInTheRunDirectory(t *testing.T) {
	dir := t.TempDir()
	stub := writeStub(t, dir, `exit 0`)
	pod, err := NewPod(PodOptions{DenoBin: stub, RunsDir: filepath.Join(dir, "runs")})
	if err != nil {
		t.Fatal(err)
	}
	id, err := pod.Start(context.Background(), runner.PodRequest{Name: "pds"})
	if err != nil {
		t.Fatal(err)
	}
	passed, err := pod.Probe(context.Background(), id, []string{"/bin/sh", "-c", "test -f main.ts"}, time.Second)
	if err != nil || !passed {
		t.Fatalf("probe = (%v, %v)", passed, err)
	}
	passed, err = pod.Probe(context.Background(), id, []string{"/bin/sh", "-c", "test -f absent"}, time.Second)
	if err != nil || passed {
		t.Fatalf("failing probe = (%v, %v)", passed, err)
	}
	if passed, err := pod.Probe(context.Background(), id, nil, time.Second); err != nil || !passed {
		t.Fatalf("an empty probe passes = (%v, %v)", passed, err)
	}
}

func TestPodRunnerTimeoutFailsTheRun(t *testing.T) {
	dir := t.TempDir()
	stub := writeStub(t, dir, `sleep 30`)
	pod, err := NewPod(PodOptions{DenoBin: stub, RunsDir: filepath.Join(dir, "runs"), Timeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	id, err := pod.Start(context.Background(), runner.PodRequest{Name: "pds"})
	if err != nil {
		t.Fatal(err)
	}
	status := observeUntilDone(t, pod, id)
	if status.State != runner.StateFailed || status.Message != "the deno process exceeded the runner timeout" {
		t.Fatalf("status = %+v", status)
	}
}

func TestEngineStartAndObserve(t *testing.T) {
	dir := t.TempDir()
	stub := writeStub(t, dir, `sleep 30`)
	engine, err := NewEngine(EngineOptions{DenoBin: stub, ServerDir: dir, RunsDir: filepath.Join(dir, "runs")})
	if err != nil {
		t.Fatal(err)
	}
	id, err := engine.Start(context.Background(), runner.EngineRequest{Name: "gha-lite", Port: 8787})
	if err != nil {
		t.Fatal(err)
	}
	status, err := engine.Observe(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if status.State != runner.StateRunning {
		t.Fatalf("status = %+v", status)
	}
	if err := engine.Stop(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	status, err = engine.Observe(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if status.State != runner.StateFailed {
		t.Fatalf("status = %+v", status)
	}
}

func TestNewPodRequiresARunsDirectory(t *testing.T) {
	if _, err := NewPod(PodOptions{}); err == nil {
		t.Fatal("RunsDir is required")
	}
	if _, err := NewEngine(EngineOptions{RunsDir: t.TempDir()}); err == nil {
		t.Fatal("ServerDir is required")
	}
}
