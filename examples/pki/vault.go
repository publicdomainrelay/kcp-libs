package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

type vault struct {
	mu sync.Mutex

	namespaces map[string]bool

	mounts map[string]map[string]string

	provisioned map[string]bool

	roles map[string]bool

	issued int

	calls []string

	listener net.Listener

	server *http.Server

	url string
}

func newVault() (*vault, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	fake := &vault{
		namespaces:  map[string]bool{"": true},
		mounts:      map[string]map[string]string{},
		provisioned: map[string]bool{},
		roles:       map[string]bool{},
		listener:    listener,
		url:         "http://" + listener.Addr().String(),
	}
	fake.server = &http.Server{Handler: fake}
	go func() {
		_ = fake.server.Serve(listener)
	}()
	return fake, nil
}

func (v *vault) URL() string {
	return v.url
}

func (v *vault) Close() error {
	return v.server.Close()
}

func (v *vault) CallCount() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return len(v.calls)
}

func (v *vault) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.mu.Lock()
	v.calls = append(v.calls, r.Method+" "+r.URL.Path)
	v.mu.Unlock()

	namespace := r.Header.Get("X-Vault-Namespace")
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	switch {
	case path == "sys/health":
		writeJSON(w, http.StatusOK, map[string]any{"initialized": true, "sealed": false, "standby": false, "version": "2.0.0"})
	case path == "sys/namespaces" || strings.HasPrefix(path, "sys/namespaces/"):
		v.serveNamespace(w, r, strings.TrimPrefix(path, "sys/namespaces/"))
	case path == "sys/mounts":
		v.serveMounts(w, namespace)
	case strings.HasPrefix(path, "sys/mounts/"):
		v.serveMountEnable(w, namespace, strings.TrimPrefix(path, "sys/mounts/"))
	case strings.HasSuffix(path, "/cert/ca"):
		v.serveCA(w, namespace)
	case strings.HasSuffix(path, "/ca_chain"):
		v.serveChain(w, namespace)
	case strings.HasSuffix(path, "/root/generate/internal"):
		v.serveRootGenerate(w, namespace)
	case strings.HasSuffix(path, "/root/sign-intermediate"):
		v.serveSignIntermediate(w, r)
	case strings.HasSuffix(path, "/intermediate/generate/internal"):
		v.data(w, map[string]any{"csr": "CSR", "private_key": "INTERMEDIATE-KEY"})
	case strings.HasSuffix(path, "/intermediate/set-signed"):
		v.serveSetSigned(w, namespace)
	case strings.Contains(path, "/roles/"):
		v.serveRole(w, namespace, path)
	case strings.Contains(path, "/issue/"):
		v.serveIssue(w, namespace, path)
	default:
		writeError(w, http.StatusNotFound, "no handler for "+path)
	}
}

func (v *vault) serveNamespace(w http.ResponseWriter, r *http.Request, path string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	switch r.Method {
	case http.MethodGet:
		if !v.namespaces[path] {
			writeError(w, http.StatusNotFound, "namespace not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPost:
		v.namespaces[path] = true
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		delete(v.namespaces, path)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (v *vault) serveMounts(w http.ResponseWriter, namespace string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	mounts := map[string]any{}
	for path, info := range v.mounts[namespace] {
		mounts[path] = map[string]any{"type": info}
	}
	v.data(w, mounts)
}

func (v *vault) serveMountEnable(w http.ResponseWriter, namespace, path string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.mounts[namespace] == nil {
		v.mounts[namespace] = map[string]string{}
	}
	v.mounts[namespace][path+"/"] = "pki"
	w.WriteHeader(http.StatusNoContent)
}

func (v *vault) serveCA(w http.ResponseWriter, namespace string) {
	v.mu.Lock()
	issued := v.provisioned[namespace]
	v.mu.Unlock()
	if !issued {
		writeError(w, http.StatusNotFound, "no authority in namespace "+namespace)
		return
	}
	authority, err := caFor(namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	v.data(w, map[string]any{"certificate": authority.pem})
}

func (v *vault) serveChain(w http.ResponseWriter, namespace string) {
	v.mu.Lock()
	issued := v.provisioned[namespace]
	v.mu.Unlock()
	if !issued {
		writeError(w, http.StatusNotFound, "no authority in namespace "+namespace)
		return
	}
	authority, err := caFor(namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	root, err := caFor("")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	chain := authority.pem
	if namespace != "" {
		chain += root.pem
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(chain))
}

func (v *vault) serveRootGenerate(w http.ResponseWriter, namespace string) {
	authority, err := caFor(namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	v.mu.Lock()
	v.provisioned[namespace] = true
	v.mu.Unlock()
	v.data(w, map[string]any{"certificate": authority.pem, "serial_number": authority.serial})
}

func (v *vault) serveSignIntermediate(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	var body struct {
		CommonName string `json:"common_name"`
	}
	_ = json.Unmarshal(raw, &body)
	authority, err := caFor(body.CommonName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	root, err := caFor("")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	v.data(w, map[string]any{"certificate": authority.pem, "ca_chain": []string{root.pem}})
}

func (v *vault) serveSetSigned(w http.ResponseWriter, namespace string) {
	v.mu.Lock()
	v.provisioned[namespace] = true
	v.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (v *vault) serveRole(w http.ResponseWriter, namespace, path string) {
	v.mu.Lock()
	v.roles[namespace+"/"+path] = true
	v.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (v *vault) serveIssue(w http.ResponseWriter, namespace, path string) {
	mount, role, _ := strings.Cut(path, "/issue/")
	v.mu.Lock()
	v.issued++
	serial := fmt.Sprintf("EE:%02X", v.issued)
	allowed := v.roles[namespace+"/"+mount+"/roles/"+role]
	v.mu.Unlock()
	if !allowed {
		writeError(w, http.StatusForbidden, "role "+role+" does not exist")
		return
	}
	authority, err := caFor(namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	root, err := caFor("")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	v.data(w, map[string]any{
		"certificate":   "LEAF " + serial,
		"private_key":   "LEAF-KEY",
		"issuing_ca":    authority.pem,
		"ca_chain":      []string{authority.pem, root.pem},
		"serial_number": serial,
	})
}

func (v *vault) data(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"errors": []string{message}})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
