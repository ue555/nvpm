package loader

import (
	"fmt"
	"log"
	"sync"

	"github.com/ue555/nvpm/pkg/core/config"
	"github.com/ue555/nvpm/pkg/core/handler"
)

// Loader manages plugin loading
type Loader struct {
	Config   *config.Config
	Handlers *handler.Registry
	mu       sync.Mutex
	loaded   map[string]bool
}

// NewLoader creates a new plugin loader
func NewLoader(cfg *config.Config) *Loader {
	return &Loader{
		Config:   cfg,
		Handlers: handler.NewRegistry(cfg),
		loaded:   make(map[string]bool),
	}
}

// Setup initializes the loader and handlers
func (l *Loader) Setup() error {
	// Initialize handlers
	if err := l.Handlers.Init(); err != nil {
		return fmt.Errorf("failed to initialize handlers: %w", err)
	}

	return nil
}

// Startup runs the startup sequence
func (l *Loader) Startup() error {
	log.Println("Starting plugin loading sequence...")

	// Run init functions for all plugins
	for _, plugin := range l.Config.Plugins {
		if plugin.Init != "" {
			log.Printf("Running init for plugin: %s\n", plugin.Name)
			// TODO: Execute init function
		}
	}

	// Load start plugins (lazy=false)
	for _, plugin := range l.Config.Plugins {
		if !plugin.Lazy && plugin.Cond && plugin.Installed {
			if err := l.Load([]*config.Plugin{plugin}, "startup", nil); err != nil {
				log.Printf("Failed to load plugin %s: %v\n", plugin.Name, err)
			}
		}
	}

	// Setup handler triggers
	if err := l.Handlers.Setup(); err != nil {
		return fmt.Errorf("failed to setup handlers: %w", err)
	}

	log.Println("Plugin loading sequence completed")
	return nil
}

// Load loads the specified plugins
func (l *Loader) Load(plugins []*config.Plugin, reason string, opts map[string]interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, plugin := range plugins {
		if err := l.loadPlugin(plugin, reason, opts); err != nil {
			return err
		}
	}

	return nil
}

// loadPlugin loads a single plugin
func (l *Loader) loadPlugin(plugin *config.Plugin, reason string, opts map[string]interface{}) error {
	// Check if already loaded
	if l.loaded[plugin.Name] {
		return nil
	}

	// Check condition
	if !plugin.Cond {
		log.Printf("Plugin %s condition is false, skipping\n", plugin.Name)
		return nil
	}

	// Check if installed
	if !plugin.Installed {
		log.Printf("Plugin %s is not installed, skipping\n", plugin.Name)
		return nil
	}

	log.Printf("Loading plugin: %s (reason: %s)\n", plugin.Name, reason)

	// Load dependencies first
	for _, depName := range plugin.Dependencies {
		if dep, ok := l.Config.GetPlugin(depName); ok {
			if err := l.loadPlugin(dep, fmt.Sprintf("dependency of %s", plugin.Name), opts); err != nil {
				return fmt.Errorf("failed to load dependency %s: %w", depName, err)
			}
		}
	}

	// Mark as loaded
	l.loaded[plugin.Name] = true
	plugin.Loaded = true

	// Run config function
	if err := l.runConfig(plugin); err != nil {
		return fmt.Errorf("failed to run config for %s: %w", plugin.Name, err)
	}

	log.Printf("Successfully loaded plugin: %s\n", plugin.Name)
	return nil
}

// runConfig runs the plugin's configuration
func (l *Loader) runConfig(plugin *config.Plugin) error {
	if plugin.Config != "" {
		log.Printf("Running config for plugin: %s\n", plugin.Name)
		// TODO: Execute config function
	}
	return nil
}

// IsLoaded checks if a plugin is loaded
func (l *Loader) IsLoaded(name string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.loaded[name]
}

// GetLoadedPlugins returns all loaded plugins
func (l *Loader) GetLoadedPlugins() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	var plugins []string
	for name := range l.loaded {
		plugins = append(plugins, name)
	}
	return plugins
}
