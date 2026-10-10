package runner

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kouji/nvpm/pkg/core/config"
	"github.com/kouji/nvpm/pkg/manage/task"
)

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

func runnerForDir(dir string) (*Runner, *config.Plugin) {
	cfg := config.DefaultConfig()
	cfg.Root = filepath.Dir(dir)
	p := &config.Plugin{Name: filepath.Base(dir), Dir: filepath.Base(dir), Installed: true}
	return NewRunner(cfg), p
}

func TestPullTask_BranchLookupFailure(t *testing.T) {
	r, p := runnerForDir(t.TempDir()) // Not a Git repository.
	job := task.NewTask("pull", p)
	if err := r.Registry.Execute(job); err == nil {
		t.Fatal("expected branch lookup error")
	}
	if job.Status != task.StatusFailed || job.FinishAt.IsZero() {
		t.Fatalf("expected finalized failure, got %+v", job)
	}
}

func TestPullTask_DetachedHEAD(t *testing.T) {
	remote, source, cleanup := setupTestRepo(t)
	defer cleanup()
	dir := filepath.Join(filepath.Dir(source), "detached")
	gitOutput(t, filepath.Dir(source), "clone", remote, dir)
	before := gitOutput(t, dir, "rev-parse", "HEAD")
	gitOutput(t, dir, "checkout", "--detach", before)
	addRemoteCommit(t, remote, source, "new remote commit")
	r, p := runnerForDir(dir)
	job := task.NewTask("pull", p)
	if err := r.Registry.Execute(job); err != nil {
		t.Fatal(err)
	}
	if job.Status != task.StatusSkipped {
		t.Fatalf("expected skipped, got %s", job.Status)
	}
	if got := gitOutput(t, dir, "rev-parse", "HEAD"); got != before {
		t.Fatalf("detached HEAD moved: %s -> %s", before, got)
	}
}

func TestUpdatePipeline_Versions(t *testing.T) {
	for _, mode := range []string{"default", "branch", "tag", "commit"} {
		t.Run(mode, func(t *testing.T) {
			remote, source, cleanup := setupTestRepo(t)
			defer cleanup()
			initial := gitOutput(t, source, "rev-parse", "HEAD")
			gitOutput(t, source, "tag", "v1.0.0")
			gitOutput(t, source, "push", "origin", "v1.0.0")
			dir := filepath.Join(filepath.Dir(source), "target")
			gitOutput(t, filepath.Dir(source), "clone", remote, dir)
			r, p := runnerForDir(dir)
			switch mode {
			case "branch":
				// The target starts on main; the pipeline must select feature.
				gitOutput(t, source, "checkout", "-b", "feature")
				gitOutput(t, source, "push", "-u", "origin", "feature")
				p.Branch = "feature"
			case "tag":
				p.Tag = "v1.0.0"
			case "commit":
				p.Commit = initial
			}
			addRemoteCommit(t, remote, source, "next version")
			expected := gitOutput(t, source, "rev-parse", "HEAD")
			pinned := mode == "tag" || mode == "commit"
			if pinned {
				// Start newer than the pin, so checkout must restore the old commit.
				gitOutput(t, dir, "pull", "--ff-only")
				expected = initial
			}
			r.QueuePipeline(p, task.UpdatePipeline)
			if err := r.Start(); err != nil {
				t.Fatal(err)
			}
			if got := gitOutput(t, dir, "rev-parse", "HEAD"); got != expected {
				t.Fatalf("HEAD = %s, want %s", got, expected)
			}
			if mode == "branch" && gitOutput(t, dir, "branch", "--show-current") != "feature" {
				t.Fatal("specified branch was not selected")
			}
			for _, job := range r.GetResults()[p.Name] {
				if job.Status == task.StatusPending || job.Status == task.StatusFailed {
					t.Fatalf("unexpected status for %s: %s", job.Name, job.Status)
				}
				if job.Name == "pull" {
					want := task.StatusSuccess
					if pinned {
						want = task.StatusSkipped
					}
					if job.Status != want {
						t.Fatalf("pull status = %s, want %s", job.Status, want)
					}
				}
			}
		})
	}
}

func TestRunner_FailureIsolation(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Performance.Concurrency = 2
	r := NewRunner(cfg)
	firstErr := errors.New("first failure")
	secondErr := errors.New("second failure")
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
	r.Registry.Register("after", func(job *task.Task) error {
		ran <- job.Plugin.Name
		return nil
	})
	for _, name := range []string{"first", "second", "healthy"} {
		r.QueuePipeline(&config.Plugin{Name: name}, &task.Pipeline{Steps: []string{"step", "after"}})
	}
	err := r.Start()
	if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) {
		t.Fatalf("expected both causes, got %v", err)
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
		t.Fatalf("unexpected stats: %v", stats)
	}
	for _, name := range []string{"first", "second"} {
		jobs := r.GetResults()[name]
		if jobs[0].FinishAt.IsZero() || jobs[1].FinishAt.IsZero() || !jobs[1].StartAt.IsZero() {
			t.Fatalf("incorrect terminal timestamps for %s", name)
		}
		if !strings.Contains(strings.Join(jobs[1].Output, "\n"), "step failed") {
			t.Fatalf("missing skip reason for %s", name)
		}
	}
}
