# nvpm

A modern Neovim plugin manager written in Go.

## Name Origin

**nvpm** stands for:
- **nv** = **N**eo**v**im
- **pm** = **P**ackage **M**anager

## Features

- **Lazy Loading**: Load plugins on-demand based on events, commands, keys, and filetypes
- **Git Management**: Install, update, and manage plugins via Git
- **Lockfile**: Version pinning and reproducibility with `nvpm-lock.json`
- **Caching**: Module caching system for fast startup
- **Task Pipeline**: Asynchronous task execution with pipelines
- **Concurrent Operations**: Parallel plugin operations with configurable concurrency
- **CLI Interface**: Command-line interface for plugin management

## Architecture

```
pkg/
├── core/           # Core engine
│   ├── config/     # Configuration management
│   ├── loader/     # Plugin loading system
│   ├── plugin/     # Plugin specification parser
│   ├── cache/      # Module caching
│   └── handler/    # Lazy-loading handlers
│       ├── event.go    # Event-based loading
│       ├── cmd.go      # Command-based loading
│       ├── ft.go       # Filetype-based loading
│       └── keys.go     # Key mapping-based loading
├── manage/         # Plugin management
│   ├── git/        # Git operations
│   ├── task/       # Task definitions
│   ├── runner/     # Task runner
│   ├── lock/       # Lockfile management
│   └── manager.go  # Management operations
├── nvpm/           # Main package
└── util/           # Utilities
```

## Installation

```bash
# Clone the repository
git clone https://github.com/ue555/nvpm.git
cd nvpm

# Install dependencies
make install

# Build the application
make build
```

## Usage

### Basic Commands

```bash
# Install missing plugins
./bin/nvpm -config examples/config.json -cmd install

# Update all plugins
./bin/nvpm -config examples/config.json -cmd update

# Sync plugins (clean + install + update)
./bin/nvpm -config examples/config.json -cmd sync

# Check for updates
./bin/nvpm -config examples/config.json -cmd check

# List all plugins
./bin/nvpm -config examples/config.json -cmd list

# Show statistics
./bin/nvpm -config examples/config.json -cmd stats

# Restore from lockfile
./bin/nvpm -config examples/config.json -cmd restore

# Clean unused plugins
./bin/nvpm -config examples/config.json -cmd clean
```

### Using Makefile

```bash
# Run with example config
make example

# Install plugins
make cmd-install

# Update plugins
make cmd-update

# Sync plugins
make cmd-sync

# List plugins
make cmd-list

# Show statistics
make cmd-stats
```

### Configuration File

Create a JSON configuration file with your plugin specifications:

```json
{
  "plugins": [
    "folke/tokyonight.nvim",
    "nvim-telescope/telescope.nvim",
    {
      "url": "hrsh7th/nvim-cmp",
      "event": ["InsertEnter"],
      "dependencies": [
        "L3MON4D3/LuaSnip"
      ]
    },
    {
      "url": "nvim-treesitter/nvim-treesitter",
      "build": ":TSUpdate",
      "event": ["BufReadPost", "BufNewFile"]
    }
  ]
}
```

### Plugin Specification

Plugins can be specified as:

1. **Simple string**: `"folke/tokyonight.nvim"`
2. **Table with options**:
   ```json
   {
     "url": "hrsh7th/nvim-cmp",
     "lazy": true,
     "event": ["InsertEnter"],
     "cmd": ["CmpStatus"],
     "ft": ["lua", "vim"],
     "keys": ["<leader>c"],
     "dependencies": ["L3MON4D3/LuaSnip"],
     "branch": "main",
     "tag": "v1.0.0",
     "commit": "abc123",
     "build": "make install",
     "config": "require('plugin').setup()"
   }
   ```

### Available Options

- `url` (string): Git repository URL or GitHub short name
- `name` (string): Custom plugin name (defaults to repo name)
- `dir` (string): Custom directory name
- `lazy` (bool): Enable lazy loading (default: true)
- `event` ([]string): Events that trigger loading
- `cmd` ([]string): Commands that trigger loading
- `ft` ([]string): Filetypes that trigger loading
- `keys` ([]string): Key mappings that trigger loading
- `dependencies` ([]string): Plugin dependencies
- `branch` (string): Git branch
- `tag` (string): Git tag
- `commit` (string): Git commit hash
- `version` (string): Semver version
- `build` (string): Build command to run after install/update. A plain string
  (e.g. `"make install"`) runs as a shell command inside the plugin's
  directory. A string prefixed with `:` (e.g. `":TSUpdate"`) runs as a Neovim
  Ex command inside an isolated, headless Neovim instance with the plugin
  and its dependencies added to `runtimepath` (see [Build Commands](#build-commands))
- `config` (string): Configuration function
- `init` (string): Initialization function (runs before loading)
- `dev` (bool): Use local development directory
- `cond` (bool): Condition to enable plugin

## Development

```bash
# Format code
make fmt

# Run tests
make test

# Clean build artifacts
make clean
```

## How It Works

### 1. Configuration Loading
The system loads plugin specifications from a JSON config file and parses them into internal plugin structures.

### 2. Plugin Loading
The loader manages the plugin lifecycle:
- Runs `init` functions for all plugins
- Loads start plugins (lazy=false)
- Sets up lazy-loading handlers for lazy plugins

### 3. Lazy Loading Handlers
Four types of handlers trigger plugin loading:
- **Event Handler**: Loads plugins when specific events occur
- **Command Handler**: Loads plugins when commands are executed
- **Filetype Handler**: Loads plugins for specific filetypes
- **Keys Handler**: Loads plugins when key mappings are pressed

### 4. Task Pipeline
Management operations (install, update, etc.) use task pipelines:
- Each operation defines a series of steps
- Tasks run concurrently with configurable concurrency
- Git operations are handled by the git module

### 5. Lockfile
The lockfile (`nvpm-lock.json`) stores the exact commit of each installed plugin:
- Enables reproducible installations
- Can restore to locked versions
- Updated automatically after install/update operations

### 6. Caching
The cache system stores intermediate results to improve performance:
- Cache entries can have TTL (time-to-live)
- Automatically cleans up expired entries
- Persists to disk for reuse across runs

## Task Pipelines

### Install Pipeline
1. `exists` - Check if plugin exists
2. `clone` - Clone repository if not exists
3. `checkout` - Checkout specific version
4. `build` - Run build command

### Update Pipeline
1. `exists` - Check if plugin exists
2. `fetch` - Fetch updates from remote
3. `checkout` - Checkout specific version
4. `build` - Run build command

### Clean Pipeline
Before queuing this pipeline, `clean` scans the plugin install directory and
compares it against the current config: any directory that no longer
corresponds to a plugin in the config (e.g. it was removed from the JSON
file) is treated as unused.

1. `remove` - Remove plugin directory

### Check Pipeline
1. `check_updates` - Check for available updates

## Build Commands

The `build` field supports two forms:

- **Shell command** (any string not starting with `:`): run via `sh -c` from
  inside the plugin's directory. Example: `"make install_jsregexp"`.
- **Neovim Ex command** (a string starting with `:`): run inside a headless,
  isolated Neovim instance (`-u NONE -i NONE`) with only the plugin's own
  directory and its declared `dependencies` added to `runtimepath`. Example:
  `":TSUpdate"`, `":MasonUpdate"`.

Since `-u NONE` also disables automatic sourcing of `plugin/` scripts, nvpm
explicitly runs `:runtime! plugin/**/*.vim plugin/**/*.lua` first. For
plugins whose commands are only registered inside `setup()` (e.g.
`mason.nvim`'s `:MasonUpdate`) rather than a `plugin/` script, nvpm makes a
best-effort attempt to `require(<module>).setup({})` first, guessing the
module name from the plugin's `lua/` directory layout.

Note that some plugins gate certain features behind interactive/headless
detection (e.g. `mason-lspconfig.nvim`'s `ensure_installed` intentionally
skips auto-install when Neovim is running headless), which is outside of
nvpm's control.

## Inspiration

This project is inspired by [lazy.nvim](https://github.com/folke/lazy.nvim) by folke, reimplemented in Go as a standalone plugin manager.

### Key Differences from lazy.nvim

1. **CLI Interface**: Uses command-line interface instead of Neovim UI
2. **Standalone**: Runs as a separate process, not integrated with Neovim
3. **Go Implementation**: Written in Go instead of Lua
4. **JSON Config**: Uses JSON instead of Lua for configuration
5. **Conceptual Lazy Loading**: Handler framework demonstrates lazy-loading concepts

## License

MIT License - See LICENSE file for details
