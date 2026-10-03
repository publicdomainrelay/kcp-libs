package livekcp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/publicdomainrelay/kcp-libs/common/kcp"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

func init() {
	os.Setenv("KUBE_FEATURE_WatchListClient", "false")
}

const (
	EnvRequire = "KCP_LIBS_REQUIRE_LIVE"

	ProviderWorkspace = "provider"

	ConsumerWorkspace = "consumer"

	Export = "probes"

	Group = "example.computer"

	Version = "v1alpha1"

	Resource = "probes"

	Kind = "Probe"
)

func Resolve(name string) string {
	switch name {
	case "kcp":
		if value := os.Getenv("KCP_BIN"); value != "" {
			return value
		}
	case "kine":
		if value := os.Getenv("KINE_BIN"); value != "" {
			return value
		}
	case "kubectl":
		if value := os.Getenv("KUBECTL"); value != "" {
			return value
		}
	}
	return name
}

func Missing() []string {
	var missing []string
	for _, name := range []string{"kcp", "kine", "kubectl"} {
		if _, err := exec.LookPath(Resolve(name)); err != nil {
			missing = append(missing, name)
		}
	}
	return missing
}

func Require(t *testing.T) {
	t.Helper()
	missing := Missing()
	if os.Getenv(EnvRequire) == "1" {
		if len(missing) > 0 {
			t.Fatalf("live tests were required but %s are not on PATH", strings.Join(missing, ", "))
		}
		return
	}
	if len(missing) > 0 {
		t.Skipf("skipping the live tier: %s are not on PATH", strings.Join(missing, ", "))
	}
	t.Skipf("skipping the live tier: set %s=1 to run it", EnvRequire)
}

type Cluster struct {
	Root string

	Kubeconfig string

	Server string

	Config *rest.Config

	Provider string

	Consumer string

	processes []*exec.Cmd

	logs []string
}

func (c *Cluster) Cluster() ref.Ref {
	return ref.Ref{LogicalCluster: c.Consumer}
}

func (c *Cluster) Stop() {
	for _, process := range c.processes {
		if process.Process != nil {
			_ = process.Process.Kill()
		}
	}
	for _, process := range c.processes {
		_, _ = process.Process.Wait()
	}
}

func (c *Cluster) Logs() string {
	var builder strings.Builder
	for _, path := range c.logs {
		body, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		builder.WriteString("== " + path + "\n")
		lines := strings.Split(strings.TrimSpace(string(body)), "\n")
		if len(lines) > 40 {
			lines = lines[len(lines)-40:]
		}
		builder.WriteString(strings.Join(lines, "\n"))
		builder.WriteString("\n")
	}
	return builder.String()
}

func Start(ctx context.Context) (*Cluster, error) {
	var lastErr error
	for attempt := range 3 {
		cluster, err := start(ctx)
		if err == nil {
			return cluster, nil
		}
		lastErr = fmt.Errorf("attempt %d: %w", attempt+1, err)
	}
	return nil, lastErr
}

func start(ctx context.Context) (*Cluster, error) {
	root, err := os.MkdirTemp("", "kcp-libs-live-")
	if err != nil {
		return nil, err
	}
	kinePort, err := freePort()
	if err != nil {
		return nil, err
	}
	kcpPort, err := freePort()
	if err != nil {
		return nil, err
	}
	cluster := &Cluster{
		Root:       root,
		Kubeconfig: filepath.Join(root, "admin.kubeconfig"),
		Server:     fmt.Sprintf("https://127.0.0.1:%d", kcpPort),
		Provider:   kcp.RootWorkspace + ":" + ProviderWorkspace,
		Consumer:   kcp.RootWorkspace + ":" + ConsumerWorkspace,
	}
	ok := false
	defer func() {
		if !ok {
			cluster.Stop()
		}
	}()

	kineLog := filepath.Join(root, "kine.log")
	cluster.logs = append(cluster.logs, kineLog)
	kine, err := spawn(kineLog, Resolve("kine"),
		"--endpoint", "sqlite://"+filepath.Join(root, "kine.db"),
		"--listen-address", fmt.Sprintf("127.0.0.1:%d", kinePort),
		"--metrics-bind-address=0",
	)
	if err != nil {
		return nil, err
	}
	cluster.processes = append(cluster.processes, kine)
	if err := waitTCP(ctx, fmt.Sprintf("127.0.0.1:%d", kinePort), 60*time.Second); err != nil {
		return nil, fmt.Errorf("kine did not listen: %w\n%s", err, cluster.Logs())
	}

	kcpLog := filepath.Join(root, "kcp.log")
	cluster.logs = append(cluster.logs, kcpLog)
	kcp, err := spawn(kcpLog, Resolve("kcp"), "start",
		"--root-directory="+root,
		"--etcd-servers="+fmt.Sprintf("http://127.0.0.1:%d", kinePort),
		"--bind-address=127.0.0.1",
		"--secure-port="+fmt.Sprint(kcpPort),
		"--feature-gates=WorkspaceMounts=true",
	)
	if err != nil {
		return nil, err
	}
	cluster.processes = append(cluster.processes, kcp)
	if err := waitReady(ctx, cluster.Server, cluster.Kubeconfig, 180*time.Second); err != nil {
		return nil, fmt.Errorf("kcp did not become ready: %w\n%s", err, cluster.Logs())
	}

	config, err := clientcmd.BuildConfigFromFlags("", cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}
	cluster.Config = config

	if err := cluster.provision(ctx); err != nil {
		return nil, fmt.Errorf("%w\n%s", err, cluster.Logs())
	}
	ok = true
	return cluster, nil
}

func (c *Cluster) provision(ctx context.Context) error {
	for _, workspace := range []string{ProviderWorkspace, ConsumerWorkspace} {
		if _, err := c.Kubectl(ctx, "", workspaceYAML(workspace)); err != nil {
			return err
		}
		if err := c.waitWorkspace(ctx, workspace); err != nil {
			return err
		}
	}
	for _, manifest := range []string{SchemaYAML, ExportYAML} {
		if _, err := c.Kubectl(ctx, c.Provider, manifest); err != nil {
			return err
		}
	}
	if err := c.waitExport(ctx); err != nil {
		return err
	}
	if _, err := c.Kubectl(ctx, c.Consumer, BindingYAML); err != nil {
		return err
	}
	return c.waitBinding(ctx)
}

func (c *Cluster) Kubectl(ctx context.Context, logicalCluster, manifest string) (string, error) {
	args := []string{"--kubeconfig=" + c.Kubeconfig}
	if logicalCluster != "" {
		args = append(args, "--server="+ref.ClusterURL(c.Server, logicalCluster))
	}
	args = append(args, "apply", "--validate=false", "-f", "-")
	return run(ctx, manifest, Resolve("kubectl"), args...)
}

func (c *Cluster) Get(ctx context.Context, logicalCluster string, args ...string) (string, error) {
	all := append([]string{"--kubeconfig=" + c.Kubeconfig}, "--server="+ref.ClusterURL(c.Server, logicalCluster))
	all = append(all, "get")
	all = append(all, args...)
	return run(ctx, "", Resolve("kubectl"), all...)
}

func (c *Cluster) waitWorkspace(ctx context.Context, name string) error {
	return waitFor(ctx, 120*time.Second, func() bool {
		out, err := c.Get(ctx, kcp.RootWorkspace, "workspace", name, "-o", "jsonpath={.status.phase}")
		return err == nil && strings.TrimSpace(out) == "Ready"
	}, "workspace "+name+" to reach Ready")
}

func (c *Cluster) waitExport(ctx context.Context) error {
	return waitFor(ctx, 120*time.Second, func() bool {
		out, err := c.Get(ctx, c.Provider, "apiexport", Export, "-o",
			`jsonpath={.status.conditions[?(@.type=="IdentityValid")].status}`)
		return err == nil && strings.TrimSpace(out) == "True"
	}, "apiexport "+Export+" to report IdentityValid")
}

func (c *Cluster) waitBinding(ctx context.Context) error {
	return waitFor(ctx, 120*time.Second, func() bool {
		out, err := c.Get(ctx, c.Consumer, "apibinding", Export, "-o", "jsonpath={.status.phase}")
		return err == nil && strings.TrimSpace(out) == "Bound"
	}, "apibinding "+Export+" to bind")
}

func spawn(logPath string, name string, args ...string) (*exec.Cmd, error) {
	log, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = log
	cmd.Stderr = log
	cmd.SysProcAttr = procAttr()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("livekcp: start %s: %w", name, err)
	}
	return cmd, nil
}

func run(ctx context.Context, stdin string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("livekcp: %s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return out.String(), nil
}

func freePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func waitTCP(ctx context.Context, address string, timeout time.Duration) error {
	return waitFor(ctx, timeout, func() bool {
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, "a listener on "+address)
}

func waitReady(ctx context.Context, server, kubeconfig string, timeout time.Duration) error {
	client := &http.Client{
		Timeout:   2 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	return waitFor(ctx, timeout, func() bool {
		if _, err := os.Stat(kubeconfig); err != nil {
			return false
		}
		resp, err := client.Get(server + "/readyz")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, "kcp to answer /readyz")
}

func waitFor(ctx context.Context, timeout time.Duration, condition func() bool, what string) error {
	deadline := time.Now().Add(timeout)
	for {
		if condition() {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("livekcp: timed out waiting for %s", what)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
}
