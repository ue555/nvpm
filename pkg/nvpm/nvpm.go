package nvpm

import (
	"fmt"
	"log"

	"github.com/kouji/nvpm/pkg/core/cache"
	"github.com/kouji/nvpm/pkg/core/config"
	"github.com/kouji/nvpm/pkg/core/loader"
	"github.com/kouji/nvpm/pkg/core/plugin"
	"github.com/kouji/nvpm/pkg/manage"
)

// NVPM is the main nvpm instance
type NVPM struct {
	Config  *config.Config
	Loader  *loader.Loader
	Manager *manage.Manager
	Cache   *cache.Cache
}

// New creates a new NVPM instance
func New() *NVPM {
	cfg := config.DefaultConfig()
	return &NVPM{
		Config:  cfg,
		Loader:  loader.NewLoader(cfg),
		Manager: manage.NewManager(cfg),
		Cache:   cache.NewCache(cfg),
	}
}

// Setup initializes nvpm with user configuration
func (n *NVPM) Setup(specs []interface{}, opts *config.Config) error {
	log.Println("Setting up nvpm...")

	// Setup configuration
	config.Setup(opts)
	n.Config = config.Options

	// Recreate components with new config
	n.Loader = loader.NewLoader(n.Config)
	n.Manager = manage.NewManager(n.Config)
	n.Cache = cache.NewCache(n.Config)

	// Enable cache
	if err := n.Cache.Enable(); err != nil {
		return fmt.Errorf("failed to enable cache: %w", err)
	}

	// Load plugin specifications
	if err := plugin.Load(n.Config, specs); err != nil {
		return fmt.Errorf("failed to load plugin specs: %w", err)
	}

	// Setup loader and handlers
	if err := n.Loader.Setup(); err != nil {
		return fmt.Errorf("failed to setup loader: %w", err)
	}

	// Install missing plugins if enabled
	if n.Config.Install.Missing {
		var missingPlugins []*config.Plugin
		for _, p := range n.Config.Plugins {
			if !p.Installed && p.Cond {
				missingPlugins = append(missingPlugins, p)
			}
		}

		if len(missingPlugins) > 0 {
			log.Printf("Installing %d missing plugins...\n", len(missingPlugins))
			if err := n.Manager.Install(missingPlugins...); err != nil {
				return fmt.Errorf("failed to install missing plugins: %w", err)
			}
		}
	}

	// Run startup sequence
	if err := n.Loader.Startup(); err != nil {
		return fmt.Errorf("failed to run startup: %w", err)
	}

	log.Println("Setup completed successfully")
	return nil
}

// Install installs plugins
func (n *NVPM) Install(plugins ...*config.Plugin) error {
	return n.Manager.Install(plugins...)
}

// Update updates plugins
func (n *NVPM) Update(plugins ...*config.Plugin) error {
	return n.Manager.Update(plugins...)
}

// Clean removes unused plugins
func (n *NVPM) Clean() error {
	return n.Manager.Clean()
}

// Sync synchronizes all plugins
func (n *NVPM) Sync() error {
	return n.Manager.Sync()
}

// Check checks for updates
func (n *NVPM) Check() error {
	return n.Manager.Check()
}

// Restore restores plugins from lockfile
func (n *NVPM) Restore() error {
	return n.Manager.Restore()
}

// Plugins returns all configured plugins
func (n *NVPM) Plugins() map[string]*config.Plugin {
	return n.Config.Plugins
}

// GetPlugin retrieves a plugin by name
func (n *NVPM) GetPlugin(name string) (*config.Plugin, bool) {
	return n.Config.GetPlugin(name)
}

// Stats returns statistics
func (n *NVPM) Stats() map[string]interface{} {
	total := len(n.Config.Plugins)
	installed := 0
	loaded := 0

	for _, p := range n.Config.Plugins {
		if p.Installed {
			installed++
		}
		if p.Loaded {
			loaded++
		}
	}

	cacheStats := n.Cache.GetStats()

	return map[string]interface{}{
		"total_plugins":      total,
		"installed_plugins":  installed,
		"loaded_plugins":     loaded,
		"cache_entries":      cacheStats.Entries,
		"cache_size_bytes":   cacheStats.TotalSize,
	}
}

// PrintStats prints statistics
func (n *NVPM) PrintStats() {
	stats := n.Stats()
	fmt.Println("\nNVPM Statistics:")
	fmt.Printf("  Total plugins:     %d\n", stats["total_plugins"])
	fmt.Printf("  Installed plugins: %d\n", stats["installed_plugins"])
	fmt.Printf("  Loaded plugins:    %d\n", stats["loaded_plugins"])
	fmt.Printf("  Cache entries:     %d\n", stats["cache_entries"])
	fmt.Printf("  Cache size:        %d bytes\n", stats["cache_size_bytes"])
}
