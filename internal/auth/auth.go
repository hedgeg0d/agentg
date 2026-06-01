package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hedgeg0d/agentg/internal/config"
)

type Decision int

const (
	Denied Decision = iota
	Allowed
	NeedPassword
)

type Principal struct {
	ID       int64
	Username string
}

type Authorizer struct {
	mu       sync.RWMutex
	path     string
	password string
	ttl      time.Duration

	seedAdmins    map[int64]bool
	seedUsers     map[int64]bool
	seedUsernames map[string]bool

	st persisted
}

type persisted struct {
	Admins    []int64         `json:"admins"`
	Users     []int64         `json:"users"`
	Usernames []string        `json:"usernames"`
	Sessions  map[int64]int64 `json:"sessions"`
}

func New(dir string, cfg config.Access) (*Authorizer, error) {
	a := &Authorizer{
		path:          filepath.Join(dir, "access.json"),
		password:      cfg.Password,
		ttl:           cfg.SessionTTL(),
		seedAdmins:    toIntSet(cfg.Admins),
		seedUsers:     toIntSet(cfg.AllowedUsers),
		seedUsernames: toNameSet(cfg.AllowedUsernames),
	}
	a.st.Sessions = map[int64]int64{}
	raw, err := os.ReadFile(a.path)
	if err == nil {
		if err := json.Unmarshal(raw, &a.st); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if a.st.Sessions == nil {
		a.st.Sessions = map[int64]int64{}
	}
	return a, nil
}

func (a *Authorizer) Check(p Principal) Decision {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.isAdmin(p.ID) || a.isAllowed(p) || a.hasSession(p.ID) {
		return Allowed
	}
	if a.password != "" {
		return NeedPassword
	}
	return Denied
}

func (a *Authorizer) Authenticate(id int64, password string) (time.Time, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.password == "" || password != a.password {
		return time.Time{}, false
	}
	var exp time.Time
	var unix int64
	if a.ttl > 0 {
		exp = time.Now().Add(a.ttl)
		unix = exp.Unix()
	}
	a.st.Sessions[id] = unix
	a.flush()
	return exp, true
}

func (a *Authorizer) IsAdmin(id int64) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isAdmin(id)
}

func (a *Authorizer) BootstrapAdmin(id int64) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.seedAdmins) > 0 || len(a.st.Admins) > 0 {
		return false
	}
	a.st.Admins = append(a.st.Admins, id)
	a.flush()
	return true
}

func (a *Authorizer) AddUser(id int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !slices.Contains(a.st.Users, id) {
		a.st.Users = append(a.st.Users, id)
		a.flush()
	}
}

func (a *Authorizer) AddUsername(name string) {
	name = normalize(name)
	a.mu.Lock()
	defer a.mu.Unlock()
	if !slices.Contains(a.st.Usernames, name) {
		a.st.Usernames = append(a.st.Usernames, name)
		a.flush()
	}
}

func (a *Authorizer) RemoveUser(id int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.st.Users = removeInt(a.st.Users, id)
	delete(a.st.Sessions, id)
	a.flush()
}

func (a *Authorizer) RemoveUsername(name string) {
	name = normalize(name)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.st.Usernames = removeStr(a.st.Usernames, name)
	a.flush()
}

type Entry struct {
	Display string
	Token   string
}

func (a *Authorizer) Entries() []Entry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	var out []Entry
	for _, id := range a.st.Users {
		out = append(out, Entry{Display: itoa(id), Token: "id:" + itoa(id)})
	}
	for _, n := range a.st.Usernames {
		out = append(out, Entry{Display: "@" + n, Token: "name:" + n})
	}
	return out
}

func (a *Authorizer) isAdmin(id int64) bool {
	return a.seedAdmins[id] || slices.Contains(a.st.Admins, id)
}

func (a *Authorizer) isAllowed(p Principal) bool {
	if a.seedUsers[p.ID] || slices.Contains(a.st.Users, p.ID) {
		return true
	}
	name := normalize(p.Username)
	return name != "" && (a.seedUsernames[name] || slices.Contains(a.st.Usernames, name))
}

func (a *Authorizer) hasSession(id int64) bool {
	exp, ok := a.st.Sessions[id]
	if !ok {
		return false
	}
	if exp != 0 && time.Now().Unix() >= exp {
		delete(a.st.Sessions, id)
		a.flush()
		return false
	}
	return true
}

func (a *Authorizer) flush() {
	raw, err := json.MarshalIndent(a.st, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(a.path, raw, 0o600)
}

func normalize(name string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(name), "@"))
}
