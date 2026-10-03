package openbaoclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/publicdomainrelay/kcp-libs/abc/pki"
)

const namespaceHeader = "X-Vault-Namespace"

const tokenHeader = "X-Vault-Token"

var (
	ErrNotFound = errors.New("openbao: not configured (HTTP 404)")

	ErrForbidden = errors.New("openbao: denied by policy (HTTP 403)")

	ErrTransport = errors.New("openbao: transport failure")

	ErrNoCA = errors.New("openbao: the supplied CA carries no CERTIFICATE PEM block")
)

type ResponseError struct {
	Method string

	Path string

	Status int

	Body string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("openbao: %s /v1/%s -> HTTP %d: %s", e.Method, e.Path, e.Status, e.Body)
}

func (e *ResponseError) Is(target error) bool {
	switch e.Status {
	case http.StatusNotFound:
		return target == ErrNotFound
	case http.StatusForbidden:
		return target == ErrForbidden
	default:
		return false
	}
}

type Options struct {
	Address string

	Token string

	CACert []byte

	Timeout time.Duration
}

type Client struct {
	address string

	token string

	http *http.Client
}

var _ pki.Client = (*Client)(nil)

func New(opts Options) (*Client, error) {
	if opts.Address == "" {
		return nil, errors.New("openbao: an address is required")
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	transport := &http.Transport{}
	if len(opts.CACert) > 0 {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(opts.CACert) {
			return nil, ErrNoCA
		}
		transport.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	}
	return &Client{
		address: strings.TrimRight(opts.Address, "/"),
		token:   opts.Token,
		http:    &http.Client{Transport: transport, Timeout: timeout},
	}, nil
}

func (c *Client) Health(ctx context.Context) (pki.Health, error) {
	status, body, err := c.do(ctx, http.MethodGet, "", "sys/health", nil)
	if err != nil {
		return pki.Health{}, err
	}
	switch status {
	case http.StatusOK, http.StatusTooManyRequests, 472, 473, http.StatusNotImplemented, http.StatusServiceUnavailable:
	default:
		return pki.Health{}, &ResponseError{Method: http.MethodGet, Path: "sys/health", Status: status, Body: strings.TrimSpace(string(body))}
	}
	var health pki.Health
	if err := json.Unmarshal(body, &health); err != nil {
		return pki.Health{}, fmt.Errorf("openbao: sys/health answered with a body that is not JSON: %w", err)
	}
	return health, nil
}

func (c *Client) EnsureNamespace(ctx context.Context, path string) error {
	if path == "" {
		return nil
	}
	exists, err := c.NamespaceExists(ctx, path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return c.write(ctx, http.MethodPost, "", "sys/namespaces/"+path, map[string]any{})
}

func (c *Client) NamespaceExists(ctx context.Context, path string) (bool, error) {
	_, _, err := c.do(ctx, http.MethodGet, "", "sys/namespaces/"+path, nil)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (c *Client) DeleteNamespace(ctx context.Context, path string) error {
	err := c.write(ctx, http.MethodDelete, "", "sys/namespaces/"+path, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

type mountInfo struct {
	Type string `json:"type"`
}

func (c *Client) mounts(ctx context.Context, namespace string) (map[string]mountInfo, error) {
	raw, err := c.data(ctx, http.MethodGet, namespace, "sys/mounts", nil)
	if err != nil {
		return nil, err
	}
	var out map[string]mountInfo
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("openbao: sys/mounts answered with a body that is not JSON: %w", err)
	}
	return out, nil
}

func (c *Client) EnsureMount(ctx context.Context, namespace, path, kind string) error {
	mounts, err := c.mounts(ctx, namespace)
	if err != nil {
		return err
	}
	key := strings.Trim(path, "/") + "/"
	if info, ok := mounts[key]; ok {
		if info.Type != kind {
			return fmt.Errorf("openbao: %s in namespace %q is mounted as %q, not %q, and this will not replace a mount in use",
				key, namespace, info.Type, kind)
		}
		return nil
	}
	return c.write(ctx, http.MethodPost, namespace, "sys/mounts/"+strings.Trim(path, "/"), map[string]any{"type": kind})
}

func (c *Client) GenerateRoot(ctx context.Context, namespace, mount, commonName, ttl string) (pki.RootCA, error) {
	body := map[string]any{"common_name": commonName}
	if ttl != "" {
		body["ttl"] = ttl
	}
	raw, err := c.data(ctx, http.MethodPost, namespace, strings.Trim(mount, "/")+"/root/generate/internal", body)
	if err != nil {
		return pki.RootCA{}, err
	}
	var out pki.RootCA
	if err := json.Unmarshal(raw, &out); err != nil {
		return pki.RootCA{}, fmt.Errorf("openbao: generating a root CA answered with a body that is not JSON: %w", err)
	}
	if out.Serial == "" {
		out.Serial = serialOfPEM(out.Certificate)
	}
	return out, nil
}

func (c *Client) CASerial(ctx context.Context, namespace, mount string) (string, error) {
	raw, err := c.data(ctx, http.MethodGet, namespace, strings.Trim(mount, "/")+"/cert/ca", nil)
	if err != nil {
		return "", err
	}
	var out struct {
		Certificate string `json:"certificate"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("openbao: reading the CA certificate answered with a body that is not JSON: %w", err)
	}
	return serialOfPEM(out.Certificate), nil
}

func serialOfPEM(body string) string {
	block, _ := pem.Decode([]byte(body))
	if block == nil {
		return ""
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return ""
	}
	hex := strings.ToUpper(cert.SerialNumber.Text(16))
	if len(hex)%2 == 1 {
		hex = "0" + hex
	}
	pairs := make([]string, 0, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		pairs = append(pairs, hex[i:i+2])
	}
	return strings.Join(pairs, ":")
}

func (c *Client) CAChain(ctx context.Context, namespace, mount string) (string, error) {
	status, body, err := c.do(ctx, http.MethodGet, namespace, strings.Trim(mount, "/")+"/ca_chain", nil)
	if err != nil {
		return "", err
	}
	if status < 200 || status > 299 {
		return "", &ResponseError{Method: http.MethodGet, Path: mount + "/ca_chain", Status: status, Body: strings.TrimSpace(string(body))}
	}
	return strings.TrimSpace(string(body)), nil
}

func (c *Client) GenerateIntermediate(ctx context.Context, namespace, mount, commonName string) (pki.IntermediateCSR, error) {
	raw, err := c.data(ctx, http.MethodPost, namespace, strings.Trim(mount, "/")+"/intermediate/generate/internal", map[string]any{
		"common_name": commonName,
		"key_type":    "ec",
		"key_bits":    256,
	})
	if err != nil {
		return pki.IntermediateCSR{}, err
	}
	var out pki.IntermediateCSR
	if err := json.Unmarshal(raw, &out); err != nil {
		return pki.IntermediateCSR{}, fmt.Errorf("openbao: generating an intermediate answered with a body that is not JSON: %w", err)
	}
	return out, nil
}

func (c *Client) SignIntermediate(ctx context.Context, rootNamespace, mount, csr, commonName, ttl string) (string, error) {
	body := map[string]any{"csr": csr, "common_name": commonName}
	if ttl != "" {
		body["ttl"] = ttl
	}
	raw, err := c.data(ctx, http.MethodPost, rootNamespace, strings.Trim(mount, "/")+"/root/sign-intermediate", body)
	if err != nil {
		return "", err
	}
	var out struct {
		Certificate string   `json:"certificate"`
		CAChain     []string `json:"ca_chain"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("openbao: signing an intermediate answered with a body that is not JSON: %w", err)
	}
	chain := out.Certificate
	for _, entry := range out.CAChain {
		chain += "\n" + entry
	}
	return chain, nil
}

func (c *Client) SetSignedIntermediate(ctx context.Context, namespace, mount, chain string) error {
	return c.write(ctx, http.MethodPost, namespace, strings.Trim(mount, "/")+"/intermediate/set-signed", map[string]any{
		"certificate": chain,
	})
}

func (c *Client) WriteRole(ctx context.Context, namespace, mount, name string, role pki.Role) error {
	body := map[string]any{
		"allowed_domains":    role.AllowedDomains,
		"allow_subdomains":   role.AllowSubdomains,
		"allow_bare_domains": role.AllowBareDomains,
		"enforce_hostnames":  role.EnforceHostnames,
		"allow_ip_sans":      true,
	}
	if role.KeyType != "" {
		body["key_type"] = role.KeyType
	}
	if role.KeyBits > 0 {
		body["key_bits"] = role.KeyBits
	}
	if role.MaxTTL != "" {
		body["max_ttl"] = role.MaxTTL
	}
	return c.write(ctx, http.MethodPost, namespace, strings.Trim(mount, "/")+"/roles/"+strings.Trim(name, "/"), body)
}

func (c *Client) Issue(ctx context.Context, namespace, mount, role string, req pki.CertRequest) (pki.Cert, error) {
	body := map[string]any{"common_name": req.CommonName}
	if len(req.AltNames) > 0 {
		body["alt_names"] = strings.Join(req.AltNames, ",")
	}
	if len(req.IPSANs) > 0 {
		body["ip_sans"] = strings.Join(req.IPSANs, ",")
	}
	if req.TTL != "" {
		body["ttl"] = req.TTL
	}
	raw, err := c.data(ctx, http.MethodPost, namespace, strings.Trim(mount, "/")+"/issue/"+strings.Trim(role, "/"), body)
	if err != nil {
		return pki.Cert{}, err
	}
	var out pki.Cert
	if err := json.Unmarshal(raw, &out); err != nil {
		return pki.Cert{}, fmt.Errorf("openbao: issuing a certificate answered with a body that is not JSON: %w", err)
	}
	return out, nil
}

func (c *Client) data(ctx context.Context, method, namespace, path string, in any) (json.RawMessage, error) {
	status, body, err := c.do(ctx, method, namespace, path, in)
	if err != nil {
		return nil, err
	}
	if status < 200 || status > 299 {
		return nil, &ResponseError{Method: method, Path: path, Status: status, Body: strings.TrimSpace(string(body))}
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("openbao: %s /v1/%s answered with a body that is not JSON: %w", method, path, err)
	}
	return envelope.Data, nil
}

func (c *Client) do(ctx context.Context, method, namespace, path string, in any) (int, []byte, error) {
	status, body, err := c.request(ctx, method, namespace, path, in)
	if err != nil {
		return 0, nil, err
	}
	if status < 200 || status > 299 {
		return status, body, responseError(method, path, status, body)
	}
	return status, body, nil
}

func (c *Client) write(ctx context.Context, method, namespace, path string, in any) error {
	_, _, err := c.do(ctx, method, namespace, path, in)
	return err
}

func responseError(method, path string, status int, body []byte) error {
	message := strings.TrimSpace(string(body))
	var envelope struct {
		Errors []string `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && len(envelope.Errors) > 0 {
		message = strings.Join(envelope.Errors, "; ")
	}
	return &ResponseError{Method: method, Path: path, Status: status, Body: message}
}

func (c *Client) request(ctx context.Context, method, namespace, path string, in any) (int, []byte, error) {
	var body io.Reader
	if in != nil {
		encoded, err := json.Marshal(in)
		if err != nil {
			return 0, nil, fmt.Errorf("%w: encode %s /v1/%s: %v", ErrTransport, method, path, err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.address+"/v1/"+strings.TrimPrefix(path, "/"), body)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: build %s /v1/%s: %v", ErrTransport, method, path, err)
	}
	if c.token != "" {
		req.Header.Set(tokenHeader, c.token)
	}
	if namespace != "" {
		req.Header.Set(namespaceHeader, strings.Trim(namespace, "/"))
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %s /v1/%s: %v", ErrTransport, method, path, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: read %s /v1/%s: %v", ErrTransport, method, path, err)
	}
	return resp.StatusCode, data, nil
}
