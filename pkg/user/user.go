package user

import (
	"errors"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// User represents a system user scoped to a tenant.
type User struct {
	Username string   `json:"username" yaml:"username"`
	Roles    []string `json:"roles" yaml:"roles"`
	TenantID string   `json:"tenantID" yaml:"-"`
}

var (
	mu      sync.RWMutex
	users   = make(map[string][]User) // tenantID -> []User
	loaded  = make(map[string]bool)
	persist bool
)

// EnablePersistence toggles writing users to disk under configs/<tenantID>/users.yaml.
func EnablePersistence(p bool) { persist = p }

// filePath returns the persistence path for a tenant.
func filePath(tenantID string) string {
	return "configs/" + tenantID + "/users.yaml"
}

// load ensures users for a tenant are loaded from disk if persistence is enabled.
func load(tenantID string) {
	if !persist || loaded[tenantID] {
		return
	}
	path := filePath(tenantID)
	data, err := os.ReadFile(path)
	if err != nil {
		loaded[tenantID] = true
		return
	}
	var wrapper struct {
		Users []User `yaml:"users"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err == nil {
		for i := range wrapper.Users {
			wrapper.Users[i].TenantID = tenantID
		}
		users[tenantID] = wrapper.Users
	}
	loaded[tenantID] = true
}

// save persists users for a tenant if enabled.
func save(tenantID string) {
	if !persist {
		return
	}
	path := filePath(tenantID)
	if err := os.MkdirAll("configs/"+tenantID, 0755); err != nil {
		return
	}
	wrapper := struct {
		Users []User `yaml:"users"`
	}{Users: users[tenantID]}
	data, err := yaml.Marshal(wrapper)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0644)
}

// Create adds a new user under a tenant.
func Create(tenantID, username string, roles []string) (User, error) {
	mu.Lock()
	defer mu.Unlock()
	load(tenantID)
	for _, u := range users[tenantID] {
		if u.Username == username {
			return User{}, errors.New("user exists")
		}
	}
	u := User{Username: username, Roles: roles, TenantID: tenantID}
	users[tenantID] = append(users[tenantID], u)
	save(tenantID)
	return u, nil
}

// AssignRoles sets roles for an existing user.
func AssignRoles(tenantID, username string, roles []string) error {
	mu.Lock()
	defer mu.Unlock()
	load(tenantID)
	for i, u := range users[tenantID] {
		if u.Username == username {
			users[tenantID][i].Roles = roles
			save(tenantID)
			return nil
		}
	}
	return errors.New("user not found")
}

// Delete removes a user from the tenant.
func Delete(tenantID, username string) error {
	mu.Lock()
	defer mu.Unlock()
	load(tenantID)
	arr := users[tenantID]
	for i, u := range arr {
		if u.Username == username {
			users[tenantID] = append(arr[:i], arr[i+1:]...)
			save(tenantID)
			return nil
		}
	}
	return errors.New("user not found")
}

// List returns all users for a tenant.
func List(tenantID string) []User {
	mu.RLock()
	defer mu.RUnlock()
	load(tenantID)
	list := users[tenantID]
	cp := make([]User, len(list))
	copy(cp, list)
	return cp
}

// Get returns a user by username.
func Get(tenantID, username string) (User, error) {
	mu.RLock()
	defer mu.RUnlock()
	load(tenantID)
	for _, u := range users[tenantID] {
		if u.Username == username {
			return u, nil
		}
	}
	return User{}, errors.New("user not found")
}

// HasRole checks if a user has any of the provided roles.
func HasRole(tenantID, username string, roles ...string) bool {
	u, err := Get(tenantID, username)
	if err != nil {
		return false
	}
	roleSet := make(map[string]struct{}, len(u.Roles))
	for _, r := range u.Roles {
		roleSet[r] = struct{}{}
	}
	for _, r := range roles {
		if _, ok := roleSet[r]; ok {
			return true
		}
	}
	return false
}

// Reset clears all users (used in tests).
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	users = make(map[string][]User)
	loaded = make(map[string]bool)
}
