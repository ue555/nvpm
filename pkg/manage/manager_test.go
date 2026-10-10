package manage

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ue555/nvpm/pkg/core/config"
)

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	output, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

func TestUpdate_PartialFailurePreservesLockfile(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_COUNT", "0")
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	source := filepath.Join(root, "source")
	runGit(t, root, "init", "--bare", "--initial-branch=main", remote)
	runGit(t, root, "clone", remote, source)
	commit := func(dir, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, "file"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		runGit(t, dir, "add", "file")
		runGit(t, dir, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", content)
	}
	commit(source, "initial")
	runGit(t, source, "push", "-u", "origin", "main")
	initial := runGit(t, source, "rev-parse", "HEAD")
	cfg := config.DefaultConfig()
	cfg.Root = root
	cfg.Lockfile = filepath.Join(root, "lock.json")
	for _, name := range []string{"healthy", "diverged"} {
		runGit(t, root, "clone", remote, filepath.Join(root, name))
		cfg.AddPlugin(&config.Plugin{Name: name, Dir: name, Installed: true, Cond: true})
	}
	m := NewManager(cfg)
	if err := m.LockManager.Update(); err != nil {
		t.Fatal(err)
	}
	if err := m.LockManager.Save(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(cfg.Lockfile)
	if err != nil {
		t.Fatal(err)
	}
	commit(filepath.Join(root, "diverged"), "local change")
	commit(source, "remote change")
	runGit(t, source, "push")
	latest := runGit(t, source, "rev-parse", "HEAD")
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previous)
	if err := m.Update(); err == nil {
		t.Fatal("expected non-fast-forward failure")
	}
	if got := runGit(t, filepath.Join(root, "healthy"), "rev-parse", "HEAD"); got != latest {
		t.Fatalf("healthy plugin did not update: %s", got)
	}
	after, err := os.ReadFile(cfg.Lockfile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("lockfile changed after partial failure")
	}
	for _, text := range []string{"Results:", "Failed tasks:", "lockfile unchanged"} {
		if !strings.Contains(logs.String(), text) {
			t.Fatalf("missing report %q: %s", text, logs.String())
		}
	}
	// Resolve only the temporary test repository, then verify a successful retry saves both commits.
	runGit(t, filepath.Join(root, "diverged"), "reset", "--hard", initial)
	if err := m.Update(); err != nil {
		t.Fatal(err)
	}
	if err := m.LockManager.Load(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"healthy", "diverged"} {
		locked, ok := m.LockManager.Get(name)
		if !ok || locked.Commit != latest {
			t.Fatalf("incorrect lock for %s: %+v", name, locked)
		}
	}
}
