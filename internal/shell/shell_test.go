package shell

import (
	"strings"
	"testing"
	"time"
)

func drain(t *testing.T, m *Manager, cmd string) (string, []string, error) {
	t.Helper()
	ch, err := m.Run(1, cmd)
	if err != nil {
		t.Fatal(err)
	}
	var snapshots []string
	var last Update
	for u := range ch {
		if !u.Done {
			snapshots = append(snapshots, u.Output)
		}
		last = u
	}
	return last.Output, snapshots, last.Err
}

func TestRunStreamsAndExitZero(t *testing.T) {
	m := NewManager(5 * time.Second)
	out, _, err := drain(t, m, "printf 'a\\nb\\nc\\n'")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(out, "a") || !strings.Contains(out, "c") {
		t.Fatalf("missing output: %q", out)
	}
	if strings.Contains(out, "exit") {
		t.Fatalf("exit-zero should not annotate: %q", out)
	}
}

func TestRunReportsExitCode(t *testing.T) {
	m := NewManager(5 * time.Second)
	out, _, err := drain(t, m, "false")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(out, "[exit 1]") {
		t.Fatalf("want exit annotation, got %q", out)
	}
}

func TestInactivityTimeout(t *testing.T) {
	m := NewManager(300 * time.Millisecond)
	_, _, err := drain(t, m, "sleep 5")
	if err == nil {
		t.Fatal("want timeout error")
	}
}

func TestSessionPersistsState(t *testing.T) {
	m := NewManager(5 * time.Second)
	drain(t, m, "export FOO=bar")
	out, _, _ := drain(t, m, "echo $FOO")
	if !strings.Contains(out, "bar") {
		t.Fatalf("session state lost: %q", out)
	}
}
