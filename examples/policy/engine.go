package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type taskState struct {
	polls int

	pollsBeforeDone int

	exitStatus string
}

type engine struct {
	mu sync.Mutex

	tasks map[string]*taskState

	pollsBeforeDone int

	exitStatus string

	submitCount int

	listener net.Listener

	server *http.Server

	url string
}

func newEngine(pollsBeforeDone int, exitStatus string) (*engine, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	fake := &engine{
		tasks:           map[string]*taskState{},
		pollsBeforeDone: pollsBeforeDone,
		exitStatus:      exitStatus,
		listener:        listener,
		url:             "http://" + listener.Addr().String(),
	}
	fake.server = &http.Server{Handler: fake}
	go func() {
		_ = fake.server.Serve(listener)
	}()
	return fake, nil
}

func (e *engine) URL() string {
	return e.url
}

func (e *engine) Close() error {
	return e.server.Close()
}

func (e *engine) Submits() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.submitCount
}

func (e *engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/request/create":
		e.serveCreate(w, r)
	case strings.HasPrefix(r.URL.Path, "/request/status/"):
		e.serveStatus(w, r)
	default:
		writeJSON(w, http.StatusNotFound, map[string]any{"status": "error", "detail": "no route"})
	}
}

func (e *engine) serveCreate(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	var body struct {
		Workflow json.RawMessage `json:"workflow"`

		Inputs map[string]string `json:"inputs"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "error"})
		return
	}
	e.mu.Lock()
	e.submitCount++
	id := "task-" + strconv.Itoa(e.submitCount)
	e.tasks[id] = &taskState{pollsBeforeDone: e.pollsBeforeDone, exitStatus: e.exitStatus}
	e.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "detail": map[string]any{"id": id}})
}

func (e *engine) serveStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/request/status/")
	e.mu.Lock()
	task := e.tasks[id]
	if task != nil {
		task.polls++
	}
	done := task != nil && task.polls > task.pollsBeforeDone
	exitStatus := e.exitStatus
	e.mu.Unlock()
	if task == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"status": "error"})
		return
	}
	if !done {
		writeJSON(w, http.StatusOK, map[string]any{"status": "in_progress", "detail": map[string]any{}})
		return
	}
	detail := map[string]any{"exit_status": exitStatus, "outputs": map[string]any{"perspective": "alice"}}
	if exitStatus == "success" {
		detail["cache"] = map[string]any{
			"policy/allow": map[string]any{
				"result.json": map[string]any{
					"data":     `{"allow":true,"violations":[]}`,
					"encoding": "utf-8",
				},
			},
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "completed", "detail": detail})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
