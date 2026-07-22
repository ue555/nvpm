package runner

import (
	"fmt"
	"log"
	"sync"

	"github.com/kouji/nvpm/pkg/core/config"
	"github.com/kouji/nvpm/pkg/manage/git"
	"github.com/kouji/nvpm/pkg/manage/task"
)

// Runner executes task pipelines
type Runner struct {
	Config   *config.Config
	Git      *git.Git
	Registry *task.Registry
	Tasks    []*task.Task
	mu       sync.Mutex
	wg       sync.WaitGroup
	sem      chan struct{} // Semaphore for concurrency control
}

// NewRunner creates a new task runner
func NewRunner(cfg *config.Config) *Runner {
	r := &Runner{
		Config:   cfg,
		Git:      git.NewGit(cfg),
		Registry: task.NewRegistry(),
		Tasks:    []*task.Task{},
		sem:      make(chan struct{}, cfg.Performance.Concurrency),
	}

	r.registerTasks()
	return r
}

// registerTasks registers all available tasks
func (r *Runner) registerTasks() {
	// Exists task - check if plugin exists
	r.Registry.Register("exists", func(t *task.Task) error {
		exists := r.Git.Exists(t.Plugin)
		if exists {
			t.Log("Plugin already exists")
			t.Plugin.Installed = true
			return nil
		}
		t.Log("Plugin does not exist")
		t.Plugin.Installed = false
		return nil
	})

	// Clone task - clone plugin repository
	r.Registry.Register("clone", func(t *task.Task) error {
		if t.Plugin.Installed {
			t.SetStatus(task.StatusSkipped)
			t.Log("Already installed, skipping clone")
			return nil
		}

		t.Log("Cloning repository: %s", t.Plugin.URL)
		if err := r.Git.Clone(t.Plugin); err != nil {
			return fmt.Errorf("failed to clone: %w", err)
		}

		t.Plugin.Installed = true
		t.Log("Successfully cloned")
		return nil
	})

	// Fetch task - fetch updates
	r.Registry.Register("fetch", func(t *task.Task) error {
		if !t.Plugin.Installed {
			return fmt.Errorf("plugin not installed")
		}

		t.Log("Fetching updates")
		if err := r.Git.Fetch(t.Plugin); err != nil {
			return fmt.Errorf("failed to fetch: %w", err)
		}

		t.Log("Successfully fetched")
		return nil
	})

	// Checkout task - checkout specific version
	r.Registry.Register("checkout", func(t *task.Task) error {
		if !t.Plugin.Installed {
			return fmt.Errorf("plugin not installed")
		}

		// Determine what to checkout
		var ref string
		if t.Plugin.Commit != "" {
			ref = t.Plugin.Commit
		} else if t.Plugin.Tag != "" {
			ref = t.Plugin.Tag
		} else if t.Plugin.Branch != "" {
			ref = t.Plugin.Branch
		}

		if ref == "" {
			t.SetStatus(task.StatusSkipped)
			t.Log("No version specified, skipping checkout")
			return nil
		}

		t.Log("Checking out: %s", ref)
		if err := r.Git.Checkout(t.Plugin, ref); err != nil {
			return fmt.Errorf("failed to checkout: %w", err)
		}

		t.Log("Successfully checked out")
		return nil
	})

	// Build task - run build command
	r.Registry.Register("build", func(t *task.Task) error {
		if t.Plugin.Build == "" {
			t.SetStatus(task.StatusSkipped)
			t.Log("No build command specified, skipping")
			return nil
		}

		t.Log("Running build command: %s", t.Plugin.Build)
		// TODO: Execute build command
		t.Log("Build completed")
		return nil
	})

	// Remove task - remove plugin
	r.Registry.Register("remove", func(t *task.Task) error {
		t.Log("Removing plugin")
		if err := r.Git.Remove(t.Plugin); err != nil {
			return fmt.Errorf("failed to remove: %w", err)
		}

		t.Plugin.Installed = false
		t.Log("Successfully removed")
		return nil
	})

	// Check updates task - check for available updates
	r.Registry.Register("check_updates", func(t *task.Task) error {
		if !t.Plugin.Installed {
			t.SetStatus(task.StatusSkipped)
			return nil
		}

		t.Log("Checking for updates")
		hasUpdates, err := r.Git.HasUpdates(t.Plugin)
		if err != nil {
			return fmt.Errorf("failed to check updates: %w", err)
		}

		if hasUpdates {
			t.Log("Updates available")
		} else {
			t.Log("Up to date")
		}

		return nil
	})
}

// Queue queues a task for execution
func (r *Runner) Queue(plugin *config.Plugin, step string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := task.NewTask(step, plugin)
	r.Tasks = append(r.Tasks, t)
}

// QueuePipeline queues all tasks in a pipeline
func (r *Runner) QueuePipeline(plugin *config.Plugin, pipeline *task.Pipeline) {
	for _, step := range pipeline.Steps {
		r.Queue(plugin, step)
	}
}

// Start starts the runner
func (r *Runner) Start() error {
	log.Printf("Starting runner with %d tasks\n", len(r.Tasks))

	// Group tasks by plugin
	pluginTasks := make(map[string][]*task.Task)
	for _, t := range r.Tasks {
		pluginTasks[t.Plugin.Name] = append(pluginTasks[t.Plugin.Name], t)
	}

	// Execute tasks for each plugin in parallel
	for pluginName, tasks := range pluginTasks {
		r.wg.Add(1)
		go func(name string, tasks []*task.Task) {
			defer r.wg.Done()

			// Acquire semaphore
			r.sem <- struct{}{}
			defer func() { <-r.sem }()

			log.Printf("Processing plugin: %s (%d tasks)\n", name, len(tasks))

			// Execute tasks sequentially for this plugin
			for _, t := range tasks {
				if err := r.Registry.Execute(t); err != nil {
					log.Printf("Task %s failed for %s: %v\n", t.Name, name, err)
					// Continue with other tasks even if one fails
				}
			}

			log.Printf("Completed plugin: %s\n", name)
		}(pluginName, tasks)
	}

	// Wait for all tasks to complete
	r.wg.Wait()

	log.Println("Runner completed")
	return nil
}

// GetResults returns task results grouped by plugin
func (r *Runner) GetResults() map[string][]*task.Task {
	r.mu.Lock()
	defer r.mu.Unlock()

	results := make(map[string][]*task.Task)
	for _, t := range r.Tasks {
		results[t.Plugin.Name] = append(results[t.Plugin.Name], t)
	}

	return results
}

// GetStats returns statistics about task execution
func (r *Runner) GetStats() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()

	stats := map[string]int{
		"total":   len(r.Tasks),
		"success": 0,
		"failed":  0,
		"skipped": 0,
		"pending": 0,
	}

	for _, t := range r.Tasks {
		switch t.Status {
		case task.StatusSuccess:
			stats["success"]++
		case task.StatusFailed:
			stats["failed"]++
		case task.StatusSkipped:
			stats["skipped"]++
		case task.StatusPending, task.StatusRunning:
			stats["pending"]++
		}
	}

	return stats
}
