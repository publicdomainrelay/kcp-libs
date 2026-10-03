package assets

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/denospec"
)

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

// The probe runs the shim in its own process under its own permissions. A
// variable the shim reads and the argv does not allow is a probe that dies with
// NotCapable, so the two lists are pinned to each other here rather than
// discovered as a service that never reports ready.
func TestTheProbeAllowsEveryVariableTheShimReads(t *testing.T) {
	reads := shimEnvReads(Shim)
	if len(reads) == 0 {
		t.Fatal("the shim reads no environment, so this test would pass without checking anything")
	}
	command := strings.Join(denospec.ProbeCommand("deno", "/runs/.kcpdns/shim.ts", "/runs/.kcpdns/probe.ts", "pds.default.alice.svc.kcp.local", "/health", DNSProbeEnv), " ")
	for _, name := range reads {
		if !knownEnv(DNSProbeEnv, name) {
			t.Fatalf("the shim reads %s and the probe command does not allow it: %s", name, command)
		}
	}
	if !strings.Contains(command, "--allow-net") {
		t.Fatalf("the probe resolves a name to an address and looks one up at the kcp API, so it needs the network: %s", command)
	}
}

// The shim is only useful if it patches the entry points Deno resolves in Rust;
// a shim that patched Deno.resolveDns alone would do nothing, which is the
// mistake its comment records. connect and connectTls share one loop.
func TestTheShimPatchesEveryEntryPointDenoUses(t *testing.T) {
	for _, want := range []string{"globalThis.fetch", "globalThis.WebSocket", `["connect", "connectTls"]`, "Deno.resolveDns"} {
		if !strings.Contains(Shim, want) {
			t.Fatalf("the shim does not patch %s", want)
		}
	}
}

func shimEnvReads(source string) []string {
	seen := map[string]bool{}
	var out []string
	for _, match := range regexp.MustCompile(`Deno\.env\.get\("([A-Z0-9_]+)"\)`).FindAllStringSubmatch(source, -1) {
		if seen[match[1]] {
			continue
		}
		seen[match[1]] = true
		out = append(out, match[1])
	}
	return out
}

func knownEnv(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
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
