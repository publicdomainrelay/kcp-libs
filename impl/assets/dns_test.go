package assets

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTheShimAndProbeAreTheOnesTheConsumerShips(t *testing.T) {
	dir := consumerKcpdns(t)
	if dir == "" {
		return
	}
	for name, embedded := range map[string]string{ShimName: Shim, ProbeName: Probe} {
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != embedded {
			t.Fatalf("%s has drifted from the copy the consumer ships, so one of the two is now wrong", name)
		}
	}
}

func consumerKcpdns(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for {
		candidate := filepath.Join(root, "deno-kcp", "internal", "provider", "kcpdns")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}
	if os.Getenv("KCP_LIBS_REQUIRE_CONSUMER") == "1" {
		t.Fatal("the consumer's kcpdns directory was not found; this test reads it as the source of truth, so set KCP_LIBS_REQUIRE_CONSUMER only where the checkout is present")
	}
	t.Skip("the consumer checkout is not next to this module, so there is nothing to compare against")
	return ""
}

func TestDNSSetWritesTheShimAndTheProbe(t *testing.T) {
	dir := t.TempDir()
	paths, err := DNSSet(dir).Materialise()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(paths[ShimName]) != filepath.Join(dir, DNSDirName) {
		t.Fatalf("the shim landed in %s, not beside the runs", filepath.Dir(paths[ShimName]))
	}
	for name, want := range map[string]string{ShimName: Shim, ProbeName: Probe} {
		body, err := os.ReadFile(paths[name])
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != want {
			t.Fatalf("%s on disk does not match the embedded copy", name)
		}
	}
}

func TestTheShimNamesEveryVariableTheProbeIsGranted(t *testing.T) {
	for _, name := range DNSProbeEnv {
		if !strings.Contains(Shim, name) {
			t.Fatalf("the probe is granted %s but the shim never reads it, so the grant is a lie", name)
		}
	}
}

func TestTheShimResolvesANameFromTheTable(t *testing.T) {
	deno, err := exec.LookPath("deno")
	if err != nil {
		t.Skip("deno is not on PATH, so the shim cannot be run")
	}
	dir := t.TempDir()
	paths, err := DNSSet(dir).Materialise()
	if err != nil {
		t.Fatal(err)
	}
	driver := filepath.Join(dir, "driver.ts")
	body := `const [addr] = await Deno.resolveDns("pds.default.svc.kcp.local", "A");` + "\n" + `console.log(addr);` + "\n"
	if err := os.WriteFile(driver, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, deno, "run",
		"--allow-env="+strings.Join(DNSProbeEnv, ","),
		"--allow-net",
		"--preload", paths[ShimName],
		driver,
	)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"KCP_SERVICE_DOMAIN=kcp.local",
		`KCP_DNS_TABLE={"pds.default.svc.kcp.local":"127.0.0.1:8080"}`,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("deno run: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "127.0.0.1" {
		t.Fatalf("the shim resolved the name to %q, the table says 127.0.0.1:8080", got)
	}
}
