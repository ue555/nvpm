package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kouji/nvpm/pkg/core/config"
	"github.com/kouji/nvpm/pkg/manage/task"
)

// setupTestRepo creates a test git repository
func setupTestRepo(t *testing.T) (bareDir, cloneDir string, cleanup func()) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_COUNT", "0")

	// Create temporary directories
	tmpDir := t.TempDir()
	bareDir = filepath.Join(tmpDir, "bare.git")
	cloneDir = filepath.Join(tmpDir, "clone")

	// Initialize bare repository
	cmd := exec.Command("git", "init", "--bare", "--initial-branch=main", bareDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init bare repo: %v", err)
	}

	// Clone the bare repository
	cmd = exec.Command("git", "clone", bareDir, cloneDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to clone repo: %v", err)
	}

	// Configure git in clone
	cmd = exec.Command("git", "-C", cloneDir, "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.name: %v", err)
	}

	cmd = exec.Command("git", "-C", cloneDir, "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.email: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(cloneDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "-C", cloneDir, "add", "test.txt")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "-C", cloneDir, "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	cmd = exec.Command("git", "-C", cloneDir, "push", "origin", "main")
	if err := cmd.Run(); err != nil {
		// Try master if main doesn't work
		cmd = exec.Command("git", "-C", cloneDir, "push", "origin", "master")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to git push: %v", err)
		}
	}

	cleanup = func() {
		// Cleanup is handled by t.TempDir()
	}

	return bareDir, cloneDir, cleanup
}

// addRemoteCommit adds a new commit to the bare repository
func addRemoteCommit(t *testing.T, bareDir, cloneDir, message string) {
	t.Helper()

	// Make a change
	testFile := filepath.Join(cloneDir, "test.txt")
	content, _ := os.ReadFile(testFile)
	if err := os.WriteFile(testFile, append(content, []byte(message+"\n")...), 0644); err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	// Commit and push
	cmd := exec.Command("git", "-C", cloneDir, "add", "test.txt")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "-C", cloneDir, "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	cmd = exec.Command("git", "-C", cloneDir, "push")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git push: %v", err)
	}
}

func TestPullTask_NoBranchSpecified(t *testing.T) {
	bareDir, cloneDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Add a remote commit
	addRemoteCommit(t, bareDir, cloneDir, "Remote change")

	// Create a second clone for testing
	testCloneDir := filepath.Join(filepath.Dir(cloneDir), "test-clone")
	cmd := exec.Command("git", "clone", bareDir, testCloneDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to clone for testing: %v", err)
	}

	// Reset to previous commit to simulate being behind
	cmd = exec.Command("git", "-C", testCloneDir, "reset", "--hard", "HEAD~1")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to reset: %v", err)
	}

	// Create config and runner
	cfg := &config.Config{
		Root: filepath.Dir(testCloneDir),
		Performance: config.PerformanceConfig{
			Concurrency: 4,
		},
	}

	plugin := &config.Plugin{
		Name:      "test-plugin",
		Dir:       filepath.Base(testCloneDir),
		Installed: true,
	}

	r := NewRunner(cfg)

	// Execute pull task
	pullTask := task.NewTask("pull", plugin)
	err := r.Registry.Execute(pullTask)

	if err != nil {
		t.Errorf("Pull task failed: %v", err)
	}

	if pullTask.Status != task.StatusSuccess {
		t.Errorf("Expected status Success, got %v", pullTask.Status)
	}

	// Verify we're now up to date
	cmd = exec.Command("git", "-C", testCloneDir, "rev-parse", "HEAD")
	localOut, _ := cmd.Output()

	cmd = exec.Command("git", "-C", cloneDir, "rev-parse", "HEAD")
	remoteOut, _ := cmd.Output()

	if string(localOut) != string(remoteOut) {
		t.Errorf("Local not updated to remote HEAD")
	}
}

func TestPullTask_BranchSpecified(t *testing.T) {
	bareDir, cloneDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Add a remote commit
	addRemoteCommit(t, bareDir, cloneDir, "Remote change on main")

	// Create a second clone
	testCloneDir := filepath.Join(filepath.Dir(cloneDir), "test-clone2")
	cmd := exec.Command("git", "clone", bareDir, testCloneDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to clone for testing: %v", err)
	}

	// Reset to previous commit
	cmd = exec.Command("git", "-C", testCloneDir, "reset", "--hard", "HEAD~1")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to reset: %v", err)
	}

	// Get current branch name
	cmd = exec.Command("git", "-C", testCloneDir, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, _ := cmd.Output()
	branchName := string(branchOut[:len(branchOut)-1]) // Remove newline

	cfg := &config.Config{
		Root: filepath.Dir(testCloneDir),
		Performance: config.PerformanceConfig{
			Concurrency: 4,
		},
	}

	plugin := &config.Plugin{
		Name:      "test-plugin",
		Dir:       filepath.Base(testCloneDir),
		Branch:    branchName,
		Installed: true,
	}

	r := NewRunner(cfg)
	pullTask := task.NewTask("pull", plugin)
	err := r.Registry.Execute(pullTask)

	if err != nil {
		t.Errorf("Pull task failed: %v", err)
	}

	if pullTask.Status != task.StatusSuccess {
		t.Errorf("Expected status Success, got %v", pullTask.Status)
	}
}

func TestPullTask_SkipForTag(t *testing.T) {
	_, cloneDir, cleanup := setupTestRepo(t)
	defer cleanup()

	cfg := &config.Config{
		Root: filepath.Dir(cloneDir),
		Performance: config.PerformanceConfig{
			Concurrency: 4,
		},
	}

	plugin := &config.Plugin{
		Name:      "test-plugin",
		Dir:       filepath.Base(cloneDir),
		Tag:       "v1.0.0",
		Installed: true,
	}

	r := NewRunner(cfg)
	pullTask := task.NewTask("pull", plugin)
	err := r.Registry.Execute(pullTask)

	if err != nil {
		t.Errorf("Pull task should not error for tag: %v", err)
	}

	if pullTask.Status != task.StatusSkipped {
		t.Errorf("Expected status Skipped for tag, got %v", pullTask.Status)
	}
}

func TestPullTask_SkipForCommit(t *testing.T) {
	_, cloneDir, cleanup := setupTestRepo(t)
	defer cleanup()

	cfg := &config.Config{
		Root: filepath.Dir(cloneDir),
		Performance: config.PerformanceConfig{
			Concurrency: 4,
		},
	}

	plugin := &config.Plugin{
		Name:      "test-plugin",
		Dir:       filepath.Base(cloneDir),
		Commit:    "abc123",
		Installed: true,
	}

	r := NewRunner(cfg)
	pullTask := task.NewTask("pull", plugin)
	err := r.Registry.Execute(pullTask)

	if err != nil {
		t.Errorf("Pull task should not error for commit: %v", err)
	}

	if pullTask.Status != task.StatusSkipped {
		t.Errorf("Expected status Skipped for commit, got %v", pullTask.Status)
	}
}

func TestUpdatePipeline_FailsOnNonFFPull(t *testing.T) {
	bareDir, cloneDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create divergent history
	// First, create a commit in the original clone and push
	addRemoteCommit(t, bareDir, cloneDir, "Remote divergent change")

	// Create test clone and make conflicting local change
	testCloneDir := filepath.Join(filepath.Dir(cloneDir), "test-clone3")
	cmd := exec.Command("git", "clone", bareDir, testCloneDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to clone for testing: %v", err)
	}

	// Reset to before the remote change
	cmd = exec.Command("git", "-C", testCloneDir, "reset", "--hard", "HEAD~1")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to reset: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "-C", testCloneDir, "config", "user.name", "Test User")
	cmd.Run()
	cmd = exec.Command("git", "-C", testCloneDir, "config", "user.email", "test@example.com")
	cmd.Run()

	// Make a conflicting local commit
	testFile := filepath.Join(testCloneDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("conflicting change\n"), 0644); err != nil {
		t.Fatalf("Failed to write conflicting change: %v", err)
	}

	cmd = exec.Command("git", "-C", testCloneDir, "add", "test.txt")
	cmd.Run()

	cmd = exec.Command("git", "-C", testCloneDir, "commit", "-m", "Local divergent change")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit local change: %v", err)
	}

	// Now try to update - should fail because --ff-only can't merge
	cfg := &config.Config{
		Root: filepath.Dir(testCloneDir),
		Performance: config.PerformanceConfig{
			Concurrency: 4,
		},
	}

	plugin := &config.Plugin{
		Name:      "test-plugin",
		Dir:       filepath.Base(testCloneDir),
		Installed: true,
	}

	r := NewRunner(cfg)

	// Queue update pipeline
	r.QueuePipeline(plugin, task.UpdatePipeline)

	// Execute should return error
	err := r.Start()
	if err == nil {
		t.Error("Expected error from non-fast-forward pull, got nil")
	}

	// Check that pull task failed
	results := r.GetResults()
	pluginTasks := results["test-plugin"]

	var pullTask *task.Task
	for _, t := range pluginTasks {
		if t.Name == "pull" {
			pullTask = t
			break
		}
	}

	if pullTask == nil {
		t.Fatal("Pull task not found in results")
	}

	if pullTask.Status != task.StatusFailed {
		t.Errorf("Expected pull task to fail, got status %v", pullTask.Status)
	}
}
