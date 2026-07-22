package task

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kouji/nvpm/pkg/core/config"
)

// Status represents task status
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusSuccess
	StatusFailed
	StatusSkipped
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusSuccess:
		return "success"
	case StatusFailed:
		return "failed"
	case StatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// Task represents a single task
type Task struct {
	Name     string
	Plugin   *config.Plugin
	Status   Status
	Error    error
	Output   []string
	StartAt  time.Time
	FinishAt time.Time
	mu       sync.Mutex
}

// NewTask creates a new task
func NewTask(name string, plugin *config.Plugin) *Task {
	return &Task{
		Name:   name,
		Plugin: plugin,
		Status: StatusPending,
		Output: []string{},
	}
}

// SetStatus sets the task status
func (t *Task) SetStatus(status Status) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = status

	if status == StatusRunning {
		t.StartAt = time.Now()
	} else if status == StatusSuccess || status == StatusFailed || status == StatusSkipped {
		t.FinishAt = time.Now()
	}
}

// SetError sets the task error
func (t *Task) SetError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Error = err
	t.Status = StatusFailed
}

// AddOutput adds output to the task
func (t *Task) AddOutput(output string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Output = append(t.Output, output)
}

// Log logs a message
func (t *Task) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	t.AddOutput(msg)
	log.Printf("[%s] %s: %s\n", t.Plugin.Name, t.Name, msg)
}

// Duration returns the task duration
func (t *Task) Duration() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.StartAt.IsZero() {
		return 0
	}

	end := t.FinishAt
	if end.IsZero() {
		end = time.Now()
	}

	return end.Sub(t.StartAt)
}

// TaskFunc is a function that executes a task
type TaskFunc func(*Task) error

// Pipeline represents a task pipeline
type Pipeline struct {
	Name  string
	Steps []string
}

// Common pipelines
var (
	InstallPipeline = &Pipeline{
		Name: "install",
		Steps: []string{
			"exists",
			"clone",
			"checkout",
			"build",
		},
	}

	UpdatePipeline = &Pipeline{
		Name: "update",
		Steps: []string{
			"exists",
			"fetch",
			"checkout",
			"build",
		},
	}

	CleanPipeline = &Pipeline{
		Name: "clean",
		Steps: []string{
			"remove",
		},
	}

	CheckPipeline = &Pipeline{
		Name: "check",
		Steps: []string{
			"check_updates",
		},
	}
)

// Registry holds all registered tasks
type Registry struct {
	tasks map[string]TaskFunc
	mu    sync.RWMutex
}

// NewRegistry creates a new task registry
func NewRegistry() *Registry {
	return &Registry{
		tasks: make(map[string]TaskFunc),
	}
}

// Register registers a task function
func (r *Registry) Register(name string, fn TaskFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[name] = fn
}

// Get retrieves a task function
func (r *Registry) Get(name string) (TaskFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.tasks[name]
	return fn, ok
}

// Execute executes a task
func (r *Registry) Execute(task *Task) error {
	fn, ok := r.Get(task.Name)
	if !ok {
		return fmt.Errorf("task not found: %s", task.Name)
	}

	task.SetStatus(StatusRunning)
	task.Log("Starting task")

	err := fn(task)
	if err != nil {
		task.SetError(err)
		task.Log("Task failed: %v", err)
		return err
	}

	task.SetStatus(StatusSuccess)
	task.Log("Task completed successfully")
	return nil
}
