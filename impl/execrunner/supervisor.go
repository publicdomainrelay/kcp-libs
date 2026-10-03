package execrunner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type runState struct {
	PID int `json:"pid"`

	Started string `json:"started"`
}

type process struct {
	dir string

	cmd *exec.Cmd

	done chan struct{}

	waitErr error

	pid int

	started time.Time

	stopped bool

	env map[string]string
}

type supervisor struct {
	prefix string

	runsDir string

	mu sync.Mutex

	runs map[string]*process

	seq atomic.Int64
}

func newSupervisor(prefix, runsDir string) *supervisor {
	return &supervisor{prefix: prefix, runsDir: runsDir, runs: map[string]*process{}}
}

func (s *supervisor) nextID() string {
	return fmt.Sprintf("%s-%s-%d", s.prefix, time.Now().UTC().Format("20060102T150405"), s.seq.Add(1))
}

func (s *supervisor) start(entry processSpec) (string, error) {
	id := s.nextID()
	dir := filepath.Join(s.runsDir, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("execrunner: create %s directory: %w", s.prefix, err)
	}
	for name, body := range entry.files {
		if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
			return "", fmt.Errorf("execrunner: write %s: %w", name, err)
		}
	}
	stdout, err := os.Create(filepath.Join(dir, "stdout.txt"))
	if err != nil {
		return "", fmt.Errorf("execrunner: create stdout file: %w", err)
	}
	defer stdout.Close()
	stderr, err := os.Create(filepath.Join(dir, "stderr.txt"))
	if err != nil {
		return "", fmt.Errorf("execrunner: create stderr file: %w", err)
	}
	defer stderr.Close()

	cmd := exec.Command(entry.binary, entry.args...)
	cmd.Dir = entry.dir
	if entry.dir == "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), entry.env...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("execrunner: start %s process: %w", s.prefix, err)
	}

	run := &process{
		dir:     dir,
		cmd:     cmd,
		done:    make(chan struct{}),
		pid:     cmd.Process.Pid,
		started: time.Now(),
		env:     entry.envMap,
	}
	go func() {
		run.waitErr = cmd.Wait()
		close(run.done)
	}()
	s.writeState(run)

	s.mu.Lock()
	s.runs[id] = run
	s.mu.Unlock()
	return id, nil
}

type processSpec struct {
	binary string

	args []string

	dir string

	files map[string][]byte

	env []string

	envMap map[string]string
}

func (s *supervisor) writeState(run *process) {
	body, err := json.Marshal(runState{PID: run.pid, Started: run.started.Format(time.RFC3339Nano)})
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(run.dir, "state.json"), body, 0o644)
}

func (s *supervisor) lookup(id string) (*process, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[id]
	return run, ok
}

func (s *supervisor) dir(id string) (string, error) {
	if run, ok := s.lookup(id); ok {
		return run.dir, nil
	}
	recovered, err := s.recover(id)
	if err != nil {
		return "", err
	}
	return recovered.dir, nil
}

func (s *supervisor) recover(id string) (*process, error) {
	dir := filepath.Join(s.runsDir, id)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("execrunner: no %s run directory for %s", s.prefix, id)
	}
	var state runState
	if body, err := os.ReadFile(filepath.Join(dir, "state.json")); err == nil {
		_ = json.Unmarshal(body, &state)
	}
	started := time.Now()
	if state.Started != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, state.Started); err == nil {
			started = parsed
		}
	}
	return &process{dir: dir, pid: state.PID, started: started}, nil
}

func (s *supervisor) resolve(id string) (*process, error) {
	if run, ok := s.lookup(id); ok {
		return run, nil
	}
	return s.recover(id)
}

func (s *supervisor) finished(run *process) bool {
	if run.cmd != nil {
		select {
		case <-run.done:
			return true
		default:
			return false
		}
	}
	return !processAlive(run.pid)
}

func (s *supervisor) stop(id string) error {
	run, err := s.resolve(id)
	if err != nil {
		return nil
	}
	if s.finished(run) {
		return nil
	}
	run.stopped = true
	if run.pid > 0 {
		_ = syscall.Kill(-run.pid, syscall.SIGKILL)
	}
	return nil
}

func (s *supervisor) probe(ctx context.Context, id string, command []string, timeout time.Duration, withEnv bool) (bool, error) {
	if len(command) == 0 {
		return true, nil
	}
	dir, err := s.dir(id)
	if err != nil {
		return false, err
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, command[0], command[1:]...)
	cmd.Dir = dir
	if withEnv {
		if run, ok := s.lookup(id); ok {
			cmd.Env = append(os.Environ(), envPairs(run.env)...)
		}
	}
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func env(req map[string]string, extra []string) []string {
	out := append([]string{}, extra...)
	keys := make([]string, 0, len(req))
	for key := range req {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		out = append(out, key+"="+req[key])
	}
	return out
}

func envPairs(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for key, value := range env {
		out = append(out, key+"="+value)
	}
	return out
}

func withDefault(values []string, key, value string) []string {
	if containsKey(values, key) {
		return values
	}
	return append(values, key+"="+value)
}

func containsKey(env []string, key string) bool {
	prefix := key + "="
	for _, entry := range env {
		if len(entry) >= len(prefix) && entry[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
