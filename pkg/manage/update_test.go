package manage

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ue555/nvpm/pkg/core/config"
	"github.com/ue555/nvpm/pkg/manage/runner"
	"github.com/ue555/nvpm/pkg/manage/task"
)

func testGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	argv := []string{"-C", dir, "-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}
	out, err := exec.Command("git", append(argv, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func testCommit(t *testing.T, dir, content string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "plugin.txt"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	testGit(t, dir, "add", "plugin.txt")
	testGit(t, dir, "commit", "-m", content)
	return testGit(t, dir, "rev-parse", "HEAD")
}

func testRemote(t *testing.T) (root, remote, source, initial string) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_COUNT", "0")
	root = t.TempDir()
	remote = filepath.Join(root, "remote.git")
	source = filepath.Join(root, "source")
	testGit(t, root, "init", "--bare", "--initial-branch=main", remote)
	testGit(t, root, "clone", remote, source)
	initial = testCommit(t, source, "initial")
	testGit(t, source, "push", "-u", "origin", "main")
	return
}

func testManager(root string, plugins ...*config.Plugin) *Manager {
	cfg := config.DefaultConfig()
	cfg.Root = root
	cfg.Lockfile = filepath.Join(root, "lock.json")
	for _, p := range plugins {
		cfg.AddPlugin(p)
	}
	return NewManager(cfg)
}

func TestUpdateAdvancesBranchesAndPreservesPins(t *testing.T) {
	for _, mode := range []string{"default", "main", "feature", "tag", "commit", "detached"} {
		t.Run(mode, func(t *testing.T) {
			root, remote, source, initial := testRemote(t)
			dir := filepath.Join(root, "plugin")
			testGit(t, root, "clone", remote, dir)
			p := &config.Plugin{Name: "plugin", Dir: "plugin", Installed: true}
			want := initial
			switch mode {
			case "main":
				p.Branch = "main"
			case "feature":
				p.Branch = "feature"
				testGit(t, source, "checkout", "-b", "feature")
				testGit(t, source, "push", "-u", "origin", "feature")
			case "tag":
				p.Tag = "v1.0.0"
				testGit(t, source, "tag", p.Tag)
				testGit(t, source, "push", "origin", p.Tag)
			case "commit":
				p.Commit = initial
			case "detached":
				testGit(t, dir, "checkout", "--detach", initial)
			}
			latest := testCommit(t, source, "updated")
			testGit(t, source, "push")
			if mode == "default" || mode == "main" || mode == "feature" {
				want = latest
			}
			if mode == "tag" || mode == "commit" {
				// Force checkout to return from a newer version to the pin.
				testGit(t, dir, "pull", "--ff-only")
			}
			m := testManager(root, p)
			if err := m.Update(); err != nil {
				t.Fatal(err)
			}
			if got := testGit(t, dir, "rev-parse", "HEAD"); got != want {
				t.Fatalf("HEAD = %s, want %s", got, want)
			}
			content, err := os.ReadFile(filepath.Join(dir, "plugin.txt"))
			if err != nil {
				t.Fatal(err)
			}
			wantContent := "initial"
			if want == latest {
				wantContent = "updated"
			}
			if string(content) != wantContent {
				t.Fatalf("worktree = %q, want %q", content, wantContent)
			}
			if mode == "feature" && testGit(t, dir, "branch", "--show-current") != "feature" {
				t.Fatal("wrong branch selected")
			}
			// Read the saved lockfile, rather than only checking the in-memory state.
			saved := NewManager(m.Config)
			if err := saved.LockManager.Load(); err != nil {
				t.Fatal(err)
			}
			locked, ok := saved.LockManager.Get(p.Name)
			if !ok || locked.Commit != want {
				t.Fatalf("incorrect saved lock: %+v", locked)
			}
		})
	}
}

func TestUpdatePartialFailurePreservesLock(t *testing.T) {
	root, remote, source, initial := testRemote(t)
	var plugins []*config.Plugin
	for _, name := range []string{"healthy", "diverged"} {
		testGit(t, root, "clone", remote, filepath.Join(root, name))
		plugins = append(plugins, &config.Plugin{Name: name, Dir: name, Installed: true})
	}
	m := testManager(root, plugins...)
	if err := m.LockManager.Update(); err != nil {
		t.Fatal(err)
	}
	if err := m.LockManager.Save(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(m.Config.Lockfile)
	if err != nil {
		t.Fatal(err)
	}
	local := testCommit(t, filepath.Join(root, "diverged"), "local change")
	latest := testCommit(t, source, "remote change")
	testGit(t, source, "push")
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previous)
	if err := m.Update(); err == nil {
		t.Fatal("expected non-fast-forward error")
	}
	for name, want := range map[string]string{"healthy": latest, "diverged": local} {
		if got := testGit(t, filepath.Join(root, name), "rev-parse", "HEAD"); got != want {
			t.Fatalf("%s HEAD = %s, want %s", name, got, want)
		}
	}
	after, err := os.ReadFile(m.Config.Lockfile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("lockfile changed after partial failure")
	}
	for _, message := range []string{"Results:", "Failed tasks:", "lockfile unchanged"} {
		if !strings.Contains(logs.String(), message) {
			t.Fatalf("missing result message %q", message)
		}
	}
	// Resolve only the temporary repository, then check that a retry saves the new versions.
	testGit(t, filepath.Join(root, "diverged"), "reset", "--hard", initial)
	if err := m.Update(); err != nil {
		t.Fatal(err)
	}
	saved := NewManager(m.Config)
	if err := saved.LockManager.Load(); err != nil {
		t.Fatal(err)
	}
	for _, p := range plugins {
		locked, ok := saved.LockManager.Get(p.Name)
		if !ok || locked.Commit != latest {
			t.Fatalf("incorrect lock after retry: %+v", locked)
		}
	}
}

func TestRunnerStopsFailedPipelineAndContinuesOthers(t *testing.T) {
	r := runner.NewRunner(config.DefaultConfig())
	firstErr, secondErr := errors.New("first"), errors.New("second")
	r.Registry.Register("step", func(job *task.Task) error {
		switch job.Plugin.Name {
		case "first":
			return firstErr
		case "second":
			return secondErr
		default:
			return nil
		}
	})
	ran := make(chan string, 3)
	r.Registry.Register("after", func(job *task.Task) error { ran <- job.Plugin.Name; return nil })
	for _, name := range []string{"first", "second", "healthy"} {
		r.QueuePipeline(&config.Plugin{Name: name}, &task.Pipeline{Steps: []string{"step", "after"}})
	}
	err := r.Start()
	if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) {
		t.Fatalf("missing errors: %v", err)
	}
	close(ran)
	var executed []string
	for name := range ran {
		executed = append(executed, name)
	}
	if len(executed) != 1 || executed[0] != "healthy" {
		t.Fatalf("unexpected downstream executions: %v", executed)
	}
	stats := r.GetStats()
	if stats["pending"] != 0 || stats["failed"] != 2 || stats["skipped"] != 2 || stats["success"] != 2 {
		t.Fatalf("incorrect stats: %v", stats)
	}
	for _, name := range []string{"first", "second"} {
		jobs := r.GetResults()[name]
		if jobs[0].FinishAt.IsZero() || jobs[1].FinishAt.IsZero() || !jobs[1].StartAt.IsZero() {
			t.Fatal("incorrect terminal timestamps")
		}
		if !strings.Contains(strings.Join(jobs[1].Output, "\n"), "step failed") {
			t.Fatal("missing skip reason")
		}
	}
}

func TestPullBranchLookupFailure(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Root = t.TempDir()
	r := runner.NewRunner(cfg)
	job := task.NewTask("pull", &config.Plugin{Name: "missing", Dir: "missing", Installed: true})
	if err := r.Registry.Execute(job); err == nil {
		t.Fatal("expected branch lookup failure")
	}
	if job.Status != task.StatusFailed {
		t.Fatalf("status = %s, want failed", job.Status)
	}
}

func TestTaskPreservesSkippedStatus(t *testing.T) {
	r := runner.NewRunner(config.DefaultConfig())
	job := task.NewTask("checkout", &config.Plugin{Name: "plugin", Installed: true})
	if err := r.Registry.Execute(job); err != nil {
		t.Fatal(err)
	}
	if job.Status != task.StatusSkipped || job.FinishAt.IsZero() {
		t.Fatalf("skipped task not finalized: %+v", job)
	}
}
