package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/publicdomainrelay/kcp-libs/abc/policy"
	"github.com/publicdomainrelay/kcp-libs/impl/policyclient"
)

var workflow = []byte(`{"jobs":{"check":{"runs-on":"ubuntu-latest","steps":[{"run":"echo hi"}]}}}`)

const followTimeout = 2 * time.Second

func Run(ctx context.Context, out io.Writer) error {
	client := policyclient.New()
	var _ policy.Client = client

	engine, err := newEngine(2, "success")
	if err != nil {
		return err
	}
	defer engine.Close()

	task, err := follow(ctx, client, engine.URL(), map[string]string{"perspective": "alice"}, out)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "the verdict was allow=%s with violations %s\n", task.Outputs["allow"], task.Outputs["violations"])
	fmt.Fprintf(out, "the engine received %d submissions\n", engine.Submits())

	failing, err := newEngine(0, "failure")
	if err != nil {
		return err
	}
	defer failing.Close()
	status, err := follow(ctx, client, failing.URL(), nil, out)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "a failed workflow reports %s: %s\n", status.State, status.Message)

	status, err = client.Status(ctx, "http://127.0.0.1:1", "task-1")
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "an unreachable engine leaves the task %s rather than failing it\n", status.State)
	return nil
}

func follow(ctx context.Context, client policy.Client, endpoint string, inputs map[string]string, out io.Writer) (policy.Task, error) {
	id, err := client.Submit(ctx, endpoint, workflow, inputs)
	if err != nil {
		return policy.Task{}, err
	}
	fmt.Fprintf(out, "submitted %s\n", id)
	deadline := time.Now().Add(followTimeout)
	for time.Now().Before(deadline) {
		task, err := client.Status(ctx, endpoint, id)
		if err != nil {
			return policy.Task{}, err
		}
		if task.State != policy.StateRunning {
			return task, nil
		}
		time.Sleep(time.Millisecond)
	}
	return policy.Task{}, fmt.Errorf("example: %s never finished", id)
}

func main() {
	if err := Run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "example-policy:", err)
		os.Exit(1)
	}
}
