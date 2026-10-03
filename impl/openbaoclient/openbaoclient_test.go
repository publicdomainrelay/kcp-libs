package openbaoclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openbao "github.com/openbao/openbao/api/v2"

	"github.com/publicdomainrelay/kcp-libs/abc/pki"
)

type request struct {
	Method string

	Path string

	Namespace string

	Token string

	Body string
}

func newServer(t *testing.T, respond func(w http.ResponseWriter, r *http.Request)) (*Client, *[]request) {
	t.Helper()
	seen := &[]request{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		*seen = append(*seen, request{
			Method:    r.Method,
			Path:      r.URL.Path,
			Namespace: r.Header.Get(openbao.NamespaceHeaderName),
			Token:     r.Header.Get("X-Vault-Token"),
			Body:      string(body),
		})
		respond(w, r)
	}))
	t.Cleanup(server.Close)
	client, err := New(Options{Address: server.URL, Token: "vault-token"})
	if err != nil {
		t.Fatal(err)
	}
	return client, seen
}

func envelope(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func TestNewRequiresAnAddressAndCA(t *testing.T) {
	if _, err := New(Options{}); err == nil {
		t.Fatal("an address is required")
	}
	if _, err := New(Options{Address: "https://bao.invalid", CACert: []byte("not pem")}); !errors.Is(err, ErrNoCA) {
		t.Fatalf("err = %v", err)
	}
}

func TestRequestsCarryTheTokenAndNamespace(t *testing.T) {
	client, seen := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/mounts" {
			envelope(w, map[string]any{})
			return
		}
		w.WriteHeader(204)
	})
	if err := client.EnsureMount(context.Background(), "alice.default", "pki", "pki"); err != nil {
		t.Fatal(err)
	}
	last := (*seen)[len(*seen)-1]
	if last.Token != "vault-token" {
		t.Fatalf("token header = %q", last.Token)
	}
	if last.Namespace != "alice.default" {
		t.Fatalf("namespace header = %q", last.Namespace)
	}
}

func TestEnsureMountReportsATypeMismatch(t *testing.T) {
	client, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		envelope(w, map[string]any{"pki/": map[string]any{"type": "kv"}})
	})
	err := client.EnsureMount(context.Background(), "alice.default", "pki", "pki")
	if err == nil || !strings.Contains(err.Error(), "mounted as") {
		t.Fatalf("err = %v", err)
	}
}

func TestEnsureMountCreatesAMissingMount(t *testing.T) {
	client, seen := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/mounts" {
			envelope(w, map[string]any{})
			return
		}
		w.WriteHeader(204)
	})
	if err := client.EnsureMount(context.Background(), "alice.default", "pki", "pki"); err != nil {
		t.Fatal(err)
	}
	last := (*seen)[len(*seen)-1]
	if last.Path != "/v1/sys/mounts/pki" || !strings.Contains(last.Body, `"type":"pki"`) {
		t.Fatalf("request = %+v", last)
	}
}

func TestNotFoundMapsToTheSentinel(t *testing.T) {
	client, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
	})
	exists, err := client.NamespaceExists(context.Background(), "alice.default")
	if err != nil || exists {
		t.Fatalf("exists = (%v, %v)", exists, err)
	}
	if err := client.DeleteNamespace(context.Background(), "alice.default"); err != nil {
		t.Fatalf("deleting a missing namespace is not an error: %v", err)
	}
}

func TestForbiddenMapsToTheSentinel(t *testing.T) {
	client, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"errors":["permission denied"]}`))
	})
	_, err := client.CASerial(context.Background(), "root", "pki")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
	var response *ResponseError
	if !errors.As(err, &response) || response.Body != "permission denied" {
		t.Fatalf("the error body must be unwrapped from the envelope: %v", err)
	}
}

func TestGenerateRootReadsTheSerialFromTheCertificate(t *testing.T) {
	certificate := testCertificate(t)
	client, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		envelope(w, map[string]any{"certificate": certificate})
	})
	root, err := client.GenerateRoot(context.Background(), "root", "pki", "kcp-mesh-root", "87600h")
	if err != nil {
		t.Fatal(err)
	}
	if root.Serial == "" {
		t.Fatal("the serial must be read out of the certificate")
	}
	if !strings.Contains(root.Serial, ":") {
		t.Fatalf("serial = %q, want colon-separated hexadecimal", root.Serial)
	}
}

func TestCAChainIsNotEnveloped(t *testing.T) {
	client, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("-----BEGIN CERTIFICATE-----\nAAA\n-----END CERTIFICATE-----\n"))
	})
	chain, err := client.CAChain(context.Background(), "alice.default", "pki")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(chain, "-----BEGIN CERTIFICATE-----") {
		t.Fatalf("chain = %q", chain)
	}
}

func TestSignIntermediateConcatenatesTheChain(t *testing.T) {
	client, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		envelope(w, map[string]any{"certificate": "SIGNED", "ca_chain": []string{"ROOT"}})
	})
	chain, err := client.SignIntermediate(context.Background(), "root", "pki", "CSR", "alice.default.intermediate", "43800h")
	if err != nil {
		t.Fatal(err)
	}
	if chain != "SIGNED\nROOT" {
		t.Fatalf("chain = %q", chain)
	}
}

func TestIssueSeparatesIPsFromNames(t *testing.T) {
	client, seen := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		envelope(w, map[string]any{
			"certificate":   "LEAF",
			"private_key":   "KEY",
			"ca_chain":      []string{"INTERMEDIATE"},
			"serial_number": "EE:FF",
		})
	})
	cert, err := client.Issue(context.Background(), "alice.default", "pki", "denopod", pki.CertRequest{
		CommonName: "pds.default.alice.svc.kcp.local",
		AltNames:   []string{"pds.default.alice.svc.kcp.local"},
		IPSANs:     []string{"127.0.0.1", "::1"},
		TTL:        "720h",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cert.Certificate != "LEAF" || cert.PrivateKey != "KEY" || cert.Serial != "EE:FF" {
		t.Fatalf("cert = %+v", cert)
	}
	body := (*seen)[0].Body
	if !strings.Contains(body, `"alt_names"`) || !strings.Contains(body, `"ip_sans"`) {
		t.Fatalf("body = %s", body)
	}
}

func TestWriteRoleCarriesDomainsAndKeyType(t *testing.T) {
	client, seen := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(204)
	})
	err := client.WriteRole(context.Background(), "alice.default", "pki", "denopod", pki.Role{
		AllowedDomains:   []string{"kcp.local"},
		AllowSubdomains:  true,
		EnforceHostnames: true,
		KeyType:          "ec",
		KeyBits:          256,
		MaxTTL:           "720h",
	})
	if err != nil {
		t.Fatal(err)
	}
	body := (*seen)[0].Body
	for _, want := range []string{`"allowed_domains":["kcp.local"]`, `"allow_subdomains":true`, `"key_type":"ec"`, `"key_bits":256`, `"allow_ip_sans":true`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body %s must contain %s", body, want)
		}
	}
}

func TestHealthAcceptsTheStandbyStatuses(t *testing.T) {
	client, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"initialized":true,"sealed":false,"standby":true,"version":"2.0.0"}`))
	})
	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !health.Initialized || health.Sealed || health.Version != "2.0.0" {
		t.Fatalf("health = %+v", health)
	}
}
