package lock

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kouji/nvpm/pkg/core/config"
	"github.com/kouji/nvpm/pkg/manage/git"
)

// Lockfile represents the lockfile structure
type Lockfile struct {
	Plugins map[string]PluginLock `json:"plugins"`
}

// PluginLock represents a plugin entry in the lockfile
type PluginLock struct {
	Commit string `json:"commit"`
	Branch string `json:"branch,omitempty"`
	Tag    string `json:"tag,omitempty"`
}

// Manager manages the lockfile
type Manager struct {
	Config   *config.Config
	Git      *git.Git
	lockfile *Lockfile
}

// NewManager creates a new lockfile manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		Config:   cfg,
		Git:      git.NewGit(cfg),
		lockfile: &Lockfile{Plugins: make(map[string]PluginLock)},
	}
}

// Load loads the lockfile
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.Config.Lockfile)
	if err != nil {
		if os.IsNotExist(err) {
			// Lockfile doesn't exist, start with empty
			return nil
		}
		return fmt.Errorf("failed to read lockfile: %w", err)
	}

	if err := json.Unmarshal(data, &m.lockfile); err != nil {
		return fmt.Errorf("failed to parse lockfile: %w", err)
	}

	return nil
}

// Save saves the lockfile
func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.lockfile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lockfile: %w", err)
	}

	if err := os.WriteFile(m.Config.Lockfile, data, 0644); err != nil {
		return fmt.Errorf("failed to write lockfile: %w", err)
	}

	return nil
}

// Update updates the lockfile with current plugin states
func (m *Manager) Update() error {
	m.lockfile.Plugins = make(map[string]PluginLock)

	for _, plugin := range m.Config.Plugins {
		if !plugin.Installed {
			continue
		}

		info, err := m.Git.GetInfo(plugin)
		if err != nil {
			return fmt.Errorf("failed to get git info for %s: %w", plugin.Name, err)
		}

		m.lockfile.Plugins[plugin.Name] = PluginLock{
			Commit: info.Commit,
			Branch: info.Branch,
			Tag:    info.Tag,
		}
	}

	return nil
}

// Restore restores plugins to lockfile versions
func (m *Manager) Restore() error {
	if err := m.Load(); err != nil {
		return err
	}

	for name, lock := range m.lockfile.Plugins {
		plugin, ok := m.Config.GetPlugin(name)
		if !ok {
			// Plugin in lockfile but not in config, skip
			continue
		}

		if !plugin.Installed {
			// Plugin not installed, skip
			continue
		}

		// Checkout the locked commit
		if err := m.Git.Checkout(plugin, lock.Commit); err != nil {
			return fmt.Errorf("failed to restore %s to %s: %w", name, lock.Commit, err)
		}
	}

	return nil
}

// Get returns the locked version for a plugin
func (m *Manager) Get(name string) (PluginLock, bool) {
	lock, ok := m.lockfile.Plugins[name]
	return lock, ok
}

// Set sets the locked version for a plugin
func (m *Manager) Set(name string, lock PluginLock) {
	m.lockfile.Plugins[name] = lock
}

// Remove removes a plugin from the lockfile
func (m *Manager) Remove(name string) {
	delete(m.lockfile.Plugins, name)
}

// GetLockfile returns the current lockfile
func (m *Manager) GetLockfile() *Lockfile {
	return m.lockfile
}
