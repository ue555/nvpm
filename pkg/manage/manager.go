package manage

import (
	"fmt"
	"log"

	"github.com/kouji/nvpm/pkg/core/config"
	"github.com/kouji/nvpm/pkg/manage/lock"
	"github.com/kouji/nvpm/pkg/manage/runner"
	"github.com/kouji/nvpm/pkg/manage/task"
)

// Manager manages plugin operations
type Manager struct {
	Config      *config.Config
	LockManager *lock.Manager
}

// NewManager creates a new plugin manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		Config:      cfg,
		LockManager: lock.NewManager(cfg),
	}
}

// Install installs missing plugins
func (m *Manager) Install(plugins ...*config.Plugin) error {
	log.Println("Starting install operation")

	r := runner.NewRunner(m.Config)

	// If no plugins specified, install all missing plugins
	if len(plugins) == 0 {
		for _, plugin := range m.Config.Plugins {
			if !plugin.Installed && plugin.Cond {
				plugins = append(plugins, plugin)
			}
		}
	}

	// Queue install pipeline for each plugin
	for _, plugin := range plugins {
		log.Printf("Queueing install for: %s\n", plugin.Name)
		r.QueuePipeline(plugin, task.InstallPipeline)
	}

	// Execute tasks
	if err := r.Start(); err != nil {
		return err
	}

	// Print results
	m.printResults(r)

	// Update lockfile
	if err := m.LockManager.Update(); err != nil {
		return fmt.Errorf("failed to update lockfile: %w", err)
	}

	if err := m.LockManager.Save(); err != nil {
		return fmt.Errorf("failed to save lockfile: %w", err)
	}

	log.Println("Install operation completed")
	return nil
}

// Update updates installed plugins
func (m *Manager) Update(plugins ...*config.Plugin) error {
	log.Println("Starting update operation")

	r := runner.NewRunner(m.Config)

	// If no plugins specified, update all installed plugins
	if len(plugins) == 0 {
		for _, plugin := range m.Config.Plugins {
			if plugin.Installed {
				plugins = append(plugins, plugin)
			}
		}
	}

	// Queue update pipeline for each plugin
	for _, plugin := range plugins {
		log.Printf("Queueing update for: %s\n", plugin.Name)
		r.QueuePipeline(plugin, task.UpdatePipeline)
	}

	// Execute tasks
	if err := r.Start(); err != nil {
		return err
	}

	// Print results
	m.printResults(r)

	// Update lockfile
	if err := m.LockManager.Update(); err != nil {
		return fmt.Errorf("failed to update lockfile: %w", err)
	}

	if err := m.LockManager.Save(); err != nil {
		return fmt.Errorf("failed to save lockfile: %w", err)
	}

	log.Println("Update operation completed")
	return nil
}

// Clean removes plugins that are not in the config
func (m *Manager) Clean() error {
	log.Println("Starting clean operation")

	r := runner.NewRunner(m.Config)

	// Queue clean pipeline for plugins to clean
	for _, name := range m.Config.ToClean {
		if plugin, ok := m.Config.GetPlugin(name); ok {
			log.Printf("Queueing clean for: %s\n", name)
			r.QueuePipeline(plugin, task.CleanPipeline)
		}
	}

	// Execute tasks
	if err := r.Start(); err != nil {
		return err
	}

	// Print results
	m.printResults(r)

	// Update lockfile
	for _, name := range m.Config.ToClean {
		m.LockManager.Remove(name)
	}

	if err := m.LockManager.Save(); err != nil {
		return fmt.Errorf("failed to save lockfile: %w", err)
	}

	// Clear to-clean list
	m.Config.ToClean = []string{}

	log.Println("Clean operation completed")
	return nil
}

// Sync synchronizes plugins (clean + install + update)
func (m *Manager) Sync() error {
	log.Println("Starting sync operation")

	// Clean
	if err := m.Clean(); err != nil {
		return fmt.Errorf("clean failed: %w", err)
	}

	// Install
	if err := m.Install(); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	// Update
	if err := m.Update(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	log.Println("Sync operation completed")
	return nil
}

// Check checks for available updates
func (m *Manager) Check() error {
	log.Println("Starting check operation")

	r := runner.NewRunner(m.Config)

	// Queue check pipeline for all installed plugins
	for _, plugin := range m.Config.Plugins {
		if plugin.Installed {
			r.QueuePipeline(plugin, task.CheckPipeline)
		}
	}

	// Execute tasks
	if err := r.Start(); err != nil {
		return err
	}

	// Print results
	m.printResults(r)

	log.Println("Check operation completed")
	return nil
}

// Restore restores plugins from lockfile
func (m *Manager) Restore() error {
	log.Println("Starting restore operation")

	if err := m.LockManager.Restore(); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	log.Println("Restore operation completed")
	return nil
}

// printResults prints task execution results
func (m *Manager) printResults(r *runner.Runner) {
	stats := r.GetStats()
	log.Printf("\nResults:\n")
	log.Printf("  Total:   %d\n", stats["total"])
	log.Printf("  Success: %d\n", stats["success"])
	log.Printf("  Failed:  %d\n", stats["failed"])
	log.Printf("  Skipped: %d\n", stats["skipped"])

	// Print failed tasks
	if stats["failed"] > 0 {
		log.Println("\nFailed tasks:")
		for pluginName, tasks := range r.GetResults() {
			for _, t := range tasks {
				if t.Status == task.StatusFailed {
					log.Printf("  - %s (%s): %v\n", pluginName, t.Name, t.Error)
				}
			}
		}
	}
}
