# Architecture Documentation

This document describes the architecture of nvpm, a Go port of the Neovim plugin manager nvpm.

## Overview

nvpm is a complete reimplementation of nvpm in Go, maintaining the same core concepts and architecture while adapting them for a standalone CLI application.

## Directory Structure

```
.
├── cmd/
│   └── lazy/           # Main CLI application
│       └── main.go     # Entry point
├── pkg/
│   ├── core/           # Core engine
│   │   ├── cache/      # Caching system
│   │   ├── config/     # Configuration management
│   │   ├── handler/    # Lazy-loading handlers
│   │   ├── loader/     # Plugin loading system
│   │   └── plugin/     # Plugin specification parser
│   ├── manage/         # Management layer
│   │   ├── git/        # Git operations
│   │   ├── lock/       # Lockfile management
│   │   ├── runner/     # Task runner
│   │   ├── task/       # Task definitions
│   │   └── manager.go  # Management operations
│   └── lazy/           # Main package
│       └── lazy.go     # Public API
├── examples/           # Example configurations
└── bin/                # Compiled binaries
```

## Core Components

### 1. Configuration (`pkg/core/config`)

**Purpose**: Manages all configuration settings and plugin metadata.

**Key Types**:
- `Config`: Main configuration structure
- `Plugin`: Plugin specification and state
- `GitConfig`: Git-related settings
- `PerformanceConfig`: Performance settings including cache

**Responsibilities**:
- Load and merge default/user configurations
- Store plugin registry
- Provide plugin lookup functions
- Manage directory paths

### 2. Plugin Parser (`pkg/core/plugin`)

**Purpose**: Parses plugin specifications and normalizes them.

**Key Types**:
- `Spec`: Plugin specification loader

**Responsibilities**:
- Parse string specs (e.g., "folke/nvpm")
- Parse table specs with options
- Extract plugin names from URLs
- Normalize plugin specifications
- Update installation state

**Example**:
```go
specs := []interface{}{
    "folke/tokyonight.nvim",
    map[string]interface{}{
        "url": "hrsh7th/nvim-cmp",
        "event": []string{"InsertEnter"},
    },
}
plugin.Load(config, specs)
```

### 3. Loader (`pkg/core/loader`)

**Purpose**: Manages the plugin loading lifecycle.

**Key Components**:
- Handler registry
- Loading sequence management
- Dependency resolution

**Loading Sequence**:
1. Run `init` functions for all plugins
2. Load start plugins (lazy=false)
3. Setup lazy-loading handlers
4. Wait for triggers

**Responsibilities**:
- Initialize plugins
- Load dependencies
- Execute config functions
- Track loaded state

### 4. Handlers (`pkg/core/handler`)

**Purpose**: Implement lazy-loading triggers.

**Handler Types**:

#### Event Handler (`event.go`)
- Triggers on autocmd events
- Examples: `BufReadPost`, `InsertEnter`, `VeryLazy`

#### Command Handler (`cmd.go`)
- Triggers on user commands
- Creates placeholder commands that load the plugin

#### Filetype Handler (`ft.go`)
- Triggers on filetype detection
- Uses FileType autocmd internally

#### Keys Handler (`keys.go`)
- Triggers on key mappings
- Creates lazy key mappings

**Common Pattern**:
```go
type Handler interface {
    Name() string
    Add(plugin *Plugin) error
    Remove(plugin *Plugin) error
    Setup() error
    Trigger(value string) error
}
```

### 5. Git Operations (`pkg/manage/git`)

**Purpose**: Handles all Git operations for plugin management.

**Key Operations**:
- `Clone`: Clone a repository
- `Fetch`: Fetch updates
- `Checkout`: Checkout branch/tag/commit
- `Pull`: Pull latest changes
- `GetInfo`: Get repository information
- `HasUpdates`: Check for updates

**Features**:
- Partial clone support (`--filter=blob:none`)
- Timeout handling
- GitHub short name support (e.g., "folke/nvpm")

### 6. Task System (`pkg/manage/task`)

**Purpose**: Defines reusable task operations.

**Key Concepts**:

#### Task
A single unit of work with:
- Name
- Plugin reference
- Status (pending/running/success/failed/skipped)
- Output logs
- Duration tracking

#### Pipeline
A sequence of tasks:
```go
InstallPipeline = &Pipeline{
    Name: "install",
    Steps: []string{
        "exists",
        "clone",
        "checkout",
        "build",
    },
}
```

#### Registry
Maps task names to implementation functions:
```go
registry.Register("clone", func(t *Task) error {
    return git.Clone(t.Plugin)
})
```

### 7. Runner (`pkg/manage/runner`)

**Purpose**: Executes task pipelines with concurrency control.

**Key Features**:
- Concurrent execution per plugin
- Configurable concurrency limit (semaphore)
- Sequential task execution within a plugin
- Results aggregation
- Statistics tracking

**Execution Flow**:
1. Queue tasks for plugins
2. Group tasks by plugin
3. Process plugins in parallel (with limit)
4. Execute tasks sequentially per plugin
5. Collect results

### 8. Lockfile (`pkg/manage/lock`)

**Purpose**: Manages version pinning via lockfile.

**Lockfile Structure**:
```json
{
  "plugins": {
    "nvpm": {
      "commit": "abc123...",
      "branch": "main"
    }
  }
}
```

**Operations**:
- `Load`: Load lockfile from disk
- `Save`: Save lockfile to disk
- `Update`: Update with current plugin states
- `Restore`: Restore plugins to locked versions

### 9. Cache (`pkg/core/cache`)

**Purpose**: Caches data to improve performance.

**Features**:
- Key-value storage
- TTL (time-to-live) support
- Disk persistence
- Automatic cleanup of expired entries
- Statistics tracking

**Usage**:
```go
cache.Set("key", data, 1*time.Hour)
if data, ok := cache.Get("key"); ok {
    // Use cached data
}
```

### 10. Manager (`pkg/manage/manager`)

**Purpose**: High-level management operations.

**Operations**:
- `Install`: Install missing plugins
- `Update`: Update installed plugins
- `Clean`: Remove unused plugins
- `Sync`: Clean + Install + Update
- `Check`: Check for updates
- `Restore`: Restore from lockfile

**Flow Example (Install)**:
```
1. Identify missing plugins
2. Create runner with install pipeline
3. Queue tasks for each plugin
4. Execute tasks concurrently
5. Update lockfile
6. Save lockfile
```

### 11. Main Package (`pkg/nvpm`)

**Purpose**: Provides the public API.

**Key Functions**:
- `New()`: Create a new Lazy instance
- `Setup(specs, opts)`: Initialize with configuration
- `Install/Update/Clean/Sync()`: Management operations
- `Stats()`: Get statistics
- `Plugins()`: Get all plugins

## Data Flow

### Initialization Flow

```
main.go
  ↓
lazy.New()
  ↓
lazy.Setup(specs, opts)
  ├─→ config.Setup(opts)           # Load configuration
  ├─→ cache.Enable()               # Enable cache
  ├─→ plugin.Load(specs)           # Parse plugin specs
  ├─→ loader.Setup()               # Setup loader & handlers
  ├─→ manager.Install()            # Install missing plugins (if enabled)
  └─→ loader.Startup()             # Run startup sequence
```

### Plugin Install Flow

```
manager.Install()
  ↓
runner.NewRunner()
  ├─→ Register task functions
  └─→ Create semaphore for concurrency
  ↓
For each plugin:
  ↓
  runner.QueuePipeline(plugin, InstallPipeline)
    ├─→ Queue "exists" task
    ├─→ Queue "clone" task
    ├─→ Queue "checkout" task
    └─→ Queue "build" task
  ↓
runner.Start()
  ├─→ Group tasks by plugin
  ├─→ For each plugin (in parallel with limit):
  │     ├─→ Acquire semaphore
  │     ├─→ Execute tasks sequentially
  │     └─→ Release semaphore
  └─→ Wait for all to complete
  ↓
lock.Update()  # Update lockfile
  ↓
lock.Save()    # Save to disk
```

### Lazy Loading Flow (Conceptual)

```
User triggers event/command/key/filetype
  ↓
Handler.Trigger(value)
  ↓
Handler.GetPlugins(value)  # Get plugins for this trigger
  ↓
For each plugin:
  ↓
  loader.Load(plugin)
    ├─→ Check if already loaded
    ├─→ Check condition
    ├─→ Load dependencies recursively
    ├─→ Mark as loaded
    └─→ Run config function
```

## Design Patterns

### 1. Registry Pattern
Used for handlers and tasks to allow dynamic registration and lookup.

### 2. Pipeline Pattern
Management operations are defined as pipelines of reusable tasks.

### 3. Concurrent Worker Pattern
Runner uses goroutines with semaphore for controlled concurrency.

### 4. State Machine
Tasks transition through states: pending → running → success/failed/skipped

### 5. Lazy Initialization
Components are created on-demand and cached.

## Concurrency Model

### Parallelism
- **Per-plugin parallelism**: Multiple plugins processed in parallel
- **Concurrency limit**: Configurable semaphore controls max parallel operations
- **Sequential tasks**: Tasks for a single plugin run sequentially

### Synchronization
- **Mutex**: Protects shared state (loader.loaded, cache.entries)
- **WaitGroup**: Coordinates goroutine completion
- **Semaphore**: Controls concurrency (buffered channel)

## Extension Points

### Adding New Tasks
```go
registry.Register("my_task", func(t *Task) error {
    // Implementation
    return nil
})
```

### Adding New Handlers
```go
type MyHandler struct {
    *BaseHandler
}

func (h *MyHandler) Trigger(value string) error {
    return h.TriggerLoad(value)
}
```

### Adding New Pipelines
```go
var MyPipeline = &Pipeline{
    Name: "my_operation",
    Steps: []string{"step1", "step2"},
}
```

## Testing Strategy

### Unit Tests
- Test individual functions (e.g., `extractName`)
- Mock dependencies where needed
- Focus on business logic

### Integration Tests
- Test component interactions
- Use real file system operations
- Test full pipelines

### Example Test
```go
func TestParseStringSpec(t *testing.T) {
    cfg := config.DefaultConfig()
    spec := NewSpec(cfg)
    err := spec.parseStringSpec("folke/nvpm")
    // Assertions...
}
```

## Performance Considerations

### Caching
- Cache expensive operations
- Use TTL to prevent stale data
- Persist cache to disk

### Concurrency
- Parallel git operations
- Configurable concurrency limit
- Avoid blocking operations

### Git Operations
- Use partial clones
- Limit fetch depth
- Reuse connections

## Differences from Original nvpm

### Architecture Similarities
✅ Same module organization (core, manage, handlers)
✅ Same lazy-loading trigger types
✅ Same task pipeline concept
✅ Same lockfile format

### Key Differences
❌ CLI instead of Neovim UI
❌ JSON config instead of Lua
❌ No Neovim integration (standalone)
❌ Go concurrency instead of Lua coroutines
✅ Additional features: explicit stats, cache API

## Future Improvements

1. **Neovim Integration**: Add Neovim RPC support
2. **More Task Types**: Health checks, documentation generation
3. **Better UI**: Progress bars, rich terminal output
4. **Plugin Hooks**: Pre/post install hooks
5. **Dependency Graph**: Visualize plugin dependencies
6. **Parallel Tests**: More comprehensive test suite
7. **Metrics**: Performance metrics and profiling
8. **Config Validation**: JSON schema validation
