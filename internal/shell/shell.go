package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const marker = "__AGENTG_DONE__"

type session struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
}

type Manager struct {
	mu       sync.Mutex
	sessions map[int64]*session
	timeout  time.Duration
}

func NewManager(timeout time.Duration) *Manager {
	return &Manager{sessions: map[int64]*session{}, timeout: timeout}
}

func (m *Manager) get(id int64) (*session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		return s, nil
	}
	s, err := newSession()
	if err != nil {
		return nil, err
	}
	m.sessions[id] = s
	return s, nil
}

func newSession() (*session, error) {
	cmd := exec.Command("bash")
	cmd.Env = append([]string{"PS1=", "TERM=dumb"}, os.Environ()...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &session{cmd: cmd, stdin: stdin, stdout: bufio.NewReader(stdout)}, nil
}

func (m *Manager) Run(id int64, command string) (string, error) {
	s, err := m.get(id)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := fmt.Fprintf(s.stdin, "%s\nprintf '\\n%s%%d\\n' \"$?\"\n", command, marker); err != nil {
		return "", err
	}

	type result struct {
		out string
		err error
	}
	done := make(chan result, 1)
	go func() {
		var b strings.Builder
		for {
			line, err := s.stdout.ReadString('\n')
			if err != nil {
				done <- result{b.String(), err}
				return
			}
			if strings.HasPrefix(line, marker) {
				code := strings.TrimSpace(strings.TrimPrefix(line, marker))
				out := b.String()
				if code != "0" && code != "" {
					out = fmt.Sprintf("%s[exit %s]", out, code)
				}
				done <- result{out, nil}
				return
			}
			b.WriteString(line)
		}
	}()

	select {
	case r := <-done:
		return r.out, r.err
	case <-time.After(m.timeout):
		return "", fmt.Errorf("command timed out after %s", m.timeout)
	}
}

func (m *Manager) Reset(id int64) {
	m.mu.Lock()
	s, ok := m.sessions[id]
	delete(m.sessions, id)
	m.mu.Unlock()
	if ok {
		s.close()
	}
}

func (s *session) close() {
	_ = s.stdin.Close()
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	_ = s.cmd.Wait()
}
