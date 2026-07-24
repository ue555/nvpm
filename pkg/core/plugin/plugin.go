package plugin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ue555/nvpm/pkg/core/config"
)

// Spec represents a plugin specification loader
type Spec struct {
	Plugins map[string]*config.Plugin
	Config  *config.Config
}

// NewSpec creates a new plugin spec loader
func NewSpec(cfg *config.Config) *Spec {
	return &Spec{
		Plugins: make(map[string]*config.Plugin),
		Config:  cfg,
	}
}

// Parse parses plugin specifications
func (s *Spec) Parse(specs []interface{}) error {
	for _, spec := range specs {
		if err := s.parseSpec(spec); err != nil {
			return err
		}
	}
	return nil
}

// parseSpec parses a single plugin specification
func (s *Spec) parseSpec(spec interface{}) error {
	switch v := spec.(type) {
	case string:
		// Simple string spec: "folke/tokyonight.nvim"
		return s.parseStringSpec(v)
	case map[string]interface{}:
		// Table spec with options
		return s.parseTableSpec(v)
	default:
		return fmt.Errorf("invalid spec type: %T", spec)
	}
}

// parseStringSpec parses a simple string plugin spec
func (s *Spec) parseStringSpec(url string) error {
	plugin := &config.Plugin{
		URL:       url,
		Lazy:      true,
		Cond:      true,
		Installed: false,
		Loaded:    false,
	}

	// Extract name from URL
	plugin.Name = s.extractName(url)
	plugin.Dir = plugin.Name

	s.Plugins[plugin.Name] = plugin
	s.Config.AddPlugin(plugin)

	return nil
}

// parseTableSpec parses a table-based plugin spec
func (s *Spec) parseTableSpec(spec map[string]interface{}) error {
	// First element should be the URL
	url, ok := spec["url"].(string)
	if !ok {
		// Try to get the first positional argument
		url, ok = spec["1"].(string)
		if !ok {
			return fmt.Errorf("plugin spec must have a URL")
		}
	}

	plugin := &config.Plugin{
		URL:       url,
		Lazy:      true,
		Cond:      true,
		Installed: false,
		Loaded:    false,
	}

	// Extract name
	plugin.Name = s.extractName(url)
	if name, ok := spec["name"].(string); ok {
		plugin.Name = name
	}

	plugin.Dir = plugin.Name
	if dir, ok := spec["dir"].(string); ok {
		plugin.Dir = dir
	}

	// Parse version info
	if branch, ok := spec["branch"].(string); ok {
		plugin.Branch = branch
	}
	if tag, ok := spec["tag"].(string); ok {
		plugin.Tag = tag
	}
	if commit, ok := spec["commit"].(string); ok {
		plugin.Commit = commit
	}
	if version, ok := spec["version"].(string); ok {
		plugin.Version = version
	}

	// Parse lazy loading options
	if lazy, ok := spec["lazy"].(bool); ok {
		plugin.Lazy = lazy
	}

	if event, ok := spec["event"].([]string); ok {
		plugin.Event = event
	} else if event, ok := spec["event"].(string); ok {
		plugin.Event = []string{event}
	}

	if cmd, ok := spec["cmd"].([]string); ok {
		plugin.Cmd = cmd
	} else if cmd, ok := spec["cmd"].(string); ok {
		plugin.Cmd = []string{cmd}
	}

	if ft, ok := spec["ft"].([]string); ok {
		plugin.Ft = ft
	} else if ft, ok := spec["ft"].(string); ok {
		plugin.Ft = []string{ft}
	}

	if keys, ok := spec["keys"].([]string); ok {
		plugin.Keys = keys
	} else if keys, ok := spec["keys"].(string); ok {
		plugin.Keys = []string{keys}
	}

	// Parse dependencies
	if deps, ok := spec["dependencies"].([]string); ok {
		plugin.Dependencies = deps
	}

	// Parse build
	if build, ok := spec["build"].(string); ok {
		plugin.Build = build
	}

	// Parse config and init
	if cfg, ok := spec["config"].(string); ok {
		plugin.Config = cfg
	}
	if init, ok := spec["init"].(string); ok {
		plugin.Init = init
	}

	// Parse dev mode
	if dev, ok := spec["dev"].(bool); ok {
		plugin.Dev = dev
	}

	// Parse condition
	if cond, ok := spec["cond"].(bool); ok {
		plugin.Cond = cond
	}

	s.Plugins[plugin.Name] = plugin
	s.Config.AddPlugin(plugin)

	return nil
}

// extractName extracts the plugin name from a URL
func (s *Spec) extractName(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Handle GitHub short names (e.g., "folke/tokyonight.nvim")
	if !strings.Contains(url, "://") {
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
		return url
	}

	// Handle full URLs
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return filepath.Base(url)
}

// Load loads all plugin specs and updates their state
func Load(cfg *config.Config, specs []interface{}) error {
	spec := NewSpec(cfg)

	if err := spec.Parse(specs); err != nil {
		return err
	}

	// Update installation state
	UpdateState(cfg)

	return nil
}

// UpdateState updates the installation state of all plugins
func UpdateState(cfg *config.Config) {
	for _, plugin := range cfg.Plugins {
		pluginDir := cfg.GetPluginDir(plugin.Dir)

		// Check if plugin is installed
		if info, err := filepath.Glob(filepath.Join(pluginDir, "*")); err == nil && len(info) > 0 {
			plugin.Installed = true
		} else {
			plugin.Installed = false
		}
	}
}

// Find finds a plugin by directory path
func Find(cfg *config.Config, path string) (*config.Plugin, bool) {
	for _, plugin := range cfg.Plugins {
		if cfg.GetPluginDir(plugin.Dir) == path {
			return plugin, true
		}
	}
	return nil, false
}

// Values returns all values for a specific plugin property
func Values(plugin *config.Plugin, prop string) []interface{} {
	var values []interface{}

	switch prop {
	case "event":
		for _, e := range plugin.Event {
			values = append(values, e)
		}
	case "cmd":
		for _, c := range plugin.Cmd {
			values = append(values, c)
		}
	case "ft":
		for _, f := range plugin.Ft {
			values = append(values, f)
		}
	case "keys":
		for _, k := range plugin.Keys {
			values = append(values, k)
		}
	}

	return values
}
