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

type Update struct {
	Output string
	Done   bool
	Err    error
}

func (m *Manager) Run(id int64, command string) (<-chan Update, error) {
	s, err := m.get(id)
	if err != nil {
		return nil, err
	}
	out := make(chan Update, 32)
	go s.run(command, m.timeout, out, func() { m.Reset(id) })
	return out, nil
}

func (s *session) run(command string, timeout time.Duration, out chan<- Update, reset func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer close(out)

	if _, err := fmt.Fprintf(s.stdin, "%s\nprintf '\\n%s%%d\\n' \"$?\"\n", command, marker); err != nil {
		out <- Update{Done: true, Err: err}
		return
	}

	lines := make(chan string)
	codes := make(chan string, 1)
	fail := make(chan error, 1)
	quit := make(chan struct{})
	defer close(quit)

	go func() {
		for {
			line, err := s.stdout.ReadString('\n')
			if err != nil {
				select {
				case fail <- err:
				case <-quit:
				}
				return
			}
			if strings.HasPrefix(line, marker) {
				select {
				case codes <- strings.TrimSpace(strings.TrimPrefix(line, marker)):
				case <-quit:
				}
				return
			}
			select {
			case lines <- line:
			case <-quit:
				return
			}
		}
	}()

	var b strings.Builder
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case line := <-lines:
			b.WriteString(line)
			out <- Update{Output: b.String()}
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(timeout)
		case code := <-codes:
			text := b.String()
			if code != "0" && code != "" {
				text += fmt.Sprintf("\n[exit %s]", code)
			}
			out <- Update{Output: text, Done: true}
			return
		case err := <-fail:
			out <- Update{Output: b.String(), Done: true, Err: err}
			return
		case <-timer.C:
			out <- Update{Output: b.String(), Done: true, Err: fmt.Errorf("no output for %s, session restarted", timeout)}
			go reset()
			return
		}
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
