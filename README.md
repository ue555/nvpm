# nvpm

A modern Neovim plugin manager written in Go.

## Name Origin

**nvpm** stands for:
- **nv** = **N**eo**v**im
- **pm** = **P**ackage **M**anager

## Features

- **Lazy-loading Framework (incomplete)**: Trigger definitions exist, but Neovim integration must be configured separately (see [Current Limitations](#current-limitations))
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

## Requirements

- **Go** 1.27 or later
- **Git** (for plugin management)
- **Neovim** 0.8.0+ (recommended 0.9.0+)

## Installation

```bash
# Clone the repository
git clone https://github.com/ue555/nvpm.git
cd nvpm

# Prepare Go dependencies (does not install the executable)
make install

# Build the application
make build
```

`make install` runs `go mod download` and `go mod tidy` only. `make build`
creates `./bin/nvpm`; use that path from the repository directory. To use
`nvpm` from another directory, optionally copy the binary to a directory
on your `PATH` (for example, `~/.local/bin`). This is separate from installing
Neovim plugins with `-cmd install`.

If `make` is unavailable, build directly with:

```bash
go build -o bin/nvpm ./cmd/nvpm
```

## Quick Start: One Plugin in Neovim

After cloning and building above, run these commands from the repository directory.
This example uses [TokyoNight](https://github.com/folke/tokyonight.nvim), which
requires Neovim 0.8.0 or later and needs no additional plugin build step.

1. Install the plugin using [examples/plugins-minimal.json](examples/plugins-minimal.json):

   ```bash
   ./bin/nvpm -config examples/plugins-minimal.json -cmd install
   ```

   Its configuration is simply `{"plugins": ["folke/tokyonight.nvim"]}`.
   The plugin is placed in `~/.local/share/nvim/nvpm/tokyonight.nvim`.

2. Start Neovim with the supplied minimal configuration:

   ```bash
   nvim -u examples/init-minimal.lua
   ```

   [examples/init-minimal.lua](examples/init-minimal.lua) contains:

   ```lua
   -- Match nvpm's install directory on Linux and macOS.
   local nvpm_path = vim.fn.expand("~/.local/share/nvim/nvpm")
   local plugin_path = nvpm_path .. "/tokyonight.nvim"
   assert(vim.fn.isdirectory(plugin_path) == 1, "Install tokyonight.nvim with nvpm first")
   vim.opt.runtimepath:prepend(plugin_path)

   -- Put plugin setup() calls after adding their directories to runtimepath.
   vim.opt.termguicolors = true
   require("tokyonight").setup({ style = "night" })
   vim.cmd("colorscheme tokyonight-night")
   ```

3. Inside Neovim, run `:colorscheme`. It should display `tokyonight-night`.
   You can also verify it without opening the UI:

   ```bash
   nvim --headless -u examples/init-minimal.lua -i NONE \
     "+lua assert(vim.g.colors_name == 'tokyonight-night'); print('nvpm: OK')" +qa
   ```

4. To use this on normal startup, merge the Lua example into your `init.lua`.
   Find its directory using `:echo stdpath('config')` (normally `~/.config/nvim`).
   Add runtime paths before `require(...).setup()` calls, then restart Neovim.
   Keep any existing configuration you need.

## Neovim Integration

### Install Paths on Linux and macOS

The CLI currently uses these fixed paths under your home directory:

| Content | Path |
| --- | --- |
| Plugins | `~/.local/share/nvim/nvpm/<plugin-name>` |
| Lockfile | `~/.local/share/nvim/nvpm-lock.json` |
| Cache | `~/.local/share/nvim/nvpm/cache` |

It does not use Neovim's `stdpath("data")`, `XDG_DATA_HOME`, or `NVIM_APPNAME`
to choose the install location. On macOS or with customized Neovim paths,
`vim.fn.stdpath("data") .. "/nvpm"` may therefore point elsewhere. Use
`vim.fn.expand("~/.local/share/nvim/nvpm")` as in the example above; changing
Neovim's data directory alone does not move nvpm's plugins. The current CLI
also does not read a custom `root` from the JSON configuration.

### Loading More Plugins

Add each plugin directory to `runtimepath` in `init.lua`, and append its
`after/` directory when present. [examples/init.lua](examples/init.lua) shows
how to add all installed plugin directories, excluding `cache`, followed by
plugin-specific configuration. Customize it for the plugins you actually use.
These examples load plugins at startup; they do not implement lazy loading.
Use a normal Neovim startup after editing `init.lua`, so Neovim can source
plugin scripts on the configured runtime path.

### Current Limitations

| Setting or feature | Current behavior |
| --- | --- |
| Git install/update, branch/tag/commit selection, lockfile | Managed by the CLI; Neovim still needs runtime paths and configuration. |
| `build` | Runs during install/update. `:` commands run in a separate headless Neovim, not your interactive session. |
| `lazy`, `event`, `cmd`, `ft`, `keys` | Framework only; the CLI does not create Neovim autocmds, commands, or mappings to load plugins. |
| `config`, `init` | Lua strings are not executed by the normal loader. Put setup and initialization in your own `init.lua`. |
| `dependencies` | Does not automatically install dependency repositories. List each required repository in the top-level `plugins` array. |

The current JSON parser also does not normalize arrays for trigger fields or
`dependencies`, so the array examples below are not working lazy-loading or
dependency-resolution recipes. Configure loading order, mappings, and any
on-demand behavior yourself in Neovim. The CLI's `loaded` statistic describes
its internal state, not plugins loaded in a running Neovim process.

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
- `lazy` (bool): Lazy-loading metadata (default: true; Neovim integration incomplete)
- `event` ([]string): Intended event triggers (not active in Neovim)
- `cmd` ([]string): Intended command triggers (not active in Neovim)
- `ft` ([]string): Intended filetype triggers (not active in Neovim)
- `keys` ([]string): Intended key triggers (not active in Neovim)
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
- `config` (string): Lua configuration text (not executed by the normal loader)
- `init` (string): Lua initialization text (not executed by the normal loader)
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
The Go loader tracks internal plugin state. Executing `init` and `config`
Lua strings is not implemented; use the [Neovim integration](#neovim-integration)
steps to load and configure plugins in your editor.

### 3. Lazy Loading Handlers
Event, command, filetype, and key handlers are a conceptual framework.
They do not register triggers with Neovim or invoke the loader on a trigger.
See [Current Limitations](#current-limitations).

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
4. `pull` - Fast-forward the current branch; skip pinned commits, tags, and detached HEAD
5. `build` - Run build command

If a task fails, the remaining tasks for that plugin are skipped and the command returns an error. Other plugins continue processing. The existing lockfile is preserved on failure, even if some plugins were updated successfully.

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
