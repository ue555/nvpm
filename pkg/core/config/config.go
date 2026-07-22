package config

import (
	"os"
	"path/filepath"
)

// Config represents the main configuration for nvpm
type Config struct {
	// Root directory where plugins are installed
	Root string

	// Lockfile path
	Lockfile string

	// Git configuration
	Git GitConfig

	// Install configuration
	Install InstallConfig

	// Performance configuration
	Performance PerformanceConfig

	// Plugins map
	Plugins map[string]*Plugin

	// ToClean list of plugins to clean
	ToClean []string
}

// GitConfig contains git-related settings
type GitConfig struct {
	// Log command to show git commits
	Log []string

	// URL format for git remotes
	URLFormat string

	// Timeout for git operations (in seconds)
	Timeout int

	// Use partial clone
	Filter string
}

// InstallConfig contains installation settings
type InstallConfig struct {
	// Missing plugins are installed automatically on startup
	Missing bool

	// Colorscheme to use during installation
	Colorscheme string
}

// PerformanceConfig contains performance settings
type PerformanceConfig struct {
	// Enable module cache
	Cache CacheConfig

	// Concurrency limit for git operations
	Concurrency int
}

// CacheConfig contains cache settings
type CacheConfig struct {
	Enabled bool
	Path    string
}

// Plugin represents a plugin configuration
type Plugin struct {
	// Plugin name (derived from repo)
	Name string

	// Git repository URL or short name (e.g., "folke/tokyonight.nvim")
	URL string

	// Directory name (defaults to name)
	Dir string

	// Git branch, tag, or commit
	Branch string
	Tag    string
	Commit string
	Version string

	// Lazy loading triggers
	Lazy   bool
	Event  []string
	Cmd    []string
	Ft     []string
	Keys   []string

	// Dependencies
	Dependencies []string

	// Condition to load
	Cond bool

	// Build command
	Build string

	// Configuration function (stored as string for now)
	Config string

	// Init function (runs before loading)
	Init string

	// Development mode (use local directory)
	Dev bool

	// Installation state
	Installed bool
	Loaded    bool

	// Git info
	GitInfo *GitInfo
}

// GitInfo contains git repository information
type GitInfo struct {
	Commit  string
	Branch  string
	Tag     string
	Remotes map[string]string
}

var (
	// Global configuration instance
	Options *Config
)

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	home := os.Getenv("HOME")
	dataPath := filepath.Join(home, ".local", "share", "nvim")

	return &Config{
		Root:     filepath.Join(dataPath, "nvpm"),
		Lockfile: filepath.Join(dataPath, "nvpm-lock.json"),
		Git: GitConfig{
			Log:       []string{"--pretty=format:%h %s (%cr)", "--abbrev-commit", "--decorate"},
			URLFormat: "https://github.com/%s.git",
			Timeout:   120,
			Filter:    "blob:none",
		},
		Install: InstallConfig{
			Missing:     true,
			Colorscheme: "default",
		},
		Performance: PerformanceConfig{
			Cache: CacheConfig{
				Enabled: true,
				Path:    filepath.Join(dataPath, "nvpm", "cache"),
			},
			Concurrency: 4,
		},
		Plugins: make(map[string]*Plugin),
		ToClean: []string{},
	}
}

// Setup initializes the configuration with user options
func Setup(opts *Config) {
	Options = DefaultConfig()

	if opts != nil {
		// Merge user options with defaults
		if opts.Root != "" {
			Options.Root = opts.Root
		}
		if opts.Lockfile != "" {
			Options.Lockfile = opts.Lockfile
		}
		if opts.Git.Timeout > 0 {
			Options.Git.Timeout = opts.Git.Timeout
		}
		if opts.Performance.Concurrency > 0 {
			Options.Performance.Concurrency = opts.Performance.Concurrency
		}
		// ... merge other options
	}

	// Ensure directories exist
	os.MkdirAll(Options.Root, 0755)
	if Options.Performance.Cache.Enabled {
		os.MkdirAll(Options.Performance.Cache.Path, 0755)
	}
}

// GetPluginDir returns the directory path for a plugin
func (c *Config) GetPluginDir(name string) string {
	return filepath.Join(c.Root, name)
}

// AddPlugin adds a plugin to the configuration
func (c *Config) AddPlugin(plugin *Plugin) {
	c.Plugins[plugin.Name] = plugin
}

// GetPlugin retrieves a plugin by name
func (c *Config) GetPlugin(name string) (*Plugin, bool) {
	plugin, ok := c.Plugins[name]
	return plugin, ok
}

// NormalizeURL converts short URLs to full git URLs
func (c *Config) NormalizeURL(url string) string {
	// If it's already a full URL, return as-is
	if len(url) > 4 && (url[:4] == "http" || url[:3] == "git") {
		return url
	}

	// Otherwise, assume it's a GitHub short name (e.g., "folke/tokyonight.nvim")
	return c.Git.URLFormat
}
