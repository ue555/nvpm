package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ue555/nvpm/pkg/core/config"
)

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create dir %s: %v", path, err)
	}
}

func TestRunEmptyBuild(t *testing.T) {
	b := NewBuilder(&config.Config{})
	output, err := b.Run(&config.Plugin{Name: "no-build"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}
}

func TestRunShellCommand(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "some-plugin"))

	cfg := &config.Config{Root: root}
	plugin := &config.Plugin{
		Name:  "some-plugin",
		Dir:   "some-plugin",
		Build: "echo hello > out.txt",
	}

	b := NewBuilder(cfg)
	if _, err := b.Run(plugin); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "some-plugin", "out.txt"))
	if err != nil {
		t.Fatalf("expected build command to run in plugin dir: %v", err)
	}
	if strings.TrimSpace(string(data)) != "hello" {
		t.Errorf("unexpected file contents: %q", data)
	}
}

func TestRunShellCommandFailure(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "broken-plugin"))

	cfg := &config.Config{Root: root}
	plugin := &config.Plugin{
		Name:  "broken-plugin",
		Dir:   "broken-plugin",
		Build: "exit 1",
	}

	b := NewBuilder(cfg)
	if _, err := b.Run(plugin); err == nil {
		t.Fatal("expected an error for a failing build command")
	}
}

func TestRunVimCommand(t *testing.T) {
	if _, err := exec.LookPath("nvim"); err != nil {
		t.Skip("nvim not found in PATH")
	}

	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "dummy-plugin"))

	cfg := &config.Config{Root: root, Plugins: map[string]*config.Plugin{}}
	plugin := &config.Plugin{
		Name:  "dummy-plugin",
		Dir:   "dummy-plugin",
		Build: ":version",
	}

	b := NewBuilder(cfg)
	output, err := b.Run(plugin)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "NVIM") {
		t.Errorf("expected :version output to mention NVIM, got: %s", output)
	}
}

func TestPrimaryModule(t *testing.T) {
	b := &Builder{}

	t.Run("exact match wins over other candidates", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdirAll(t, filepath.Join(dir, "lua", "mason"))
		mustMkdirAll(t, filepath.Join(dir, "lua", "mason-core"))

		got := b.primaryModule(dir, "mason.nvim")
		if got != "mason" {
			t.Errorf("got %q, want %q", got, "mason")
		}
	})

	t.Run("single candidate fallback", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdirAll(t, filepath.Join(dir, "lua", "insx"))

		got := b.primaryModule(dir, "nvim-insx")
		if got != "insx" {
			t.Errorf("got %q, want %q", got, "insx")
		}
	})

	t.Run("lua file entries count as candidates", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdirAll(t, filepath.Join(dir, "lua"))
		if err := os.WriteFile(filepath.Join(dir, "lua", "foo.lua"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		got := b.primaryModule(dir, "foo")
		if got != "foo" {
			t.Errorf("got %q, want %q", got, "foo")
		}
	})

	t.Run("ambiguous multiple candidates return empty", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdirAll(t, filepath.Join(dir, "lua", "foo"))
		mustMkdirAll(t, filepath.Join(dir, "lua", "bar"))

		got := b.primaryModule(dir, "something-else")
		if got != "" {
			t.Errorf("expected no match, got %q", got)
		}
	})

	t.Run("no lua directory returns empty", func(t *testing.T) {
		dir := t.TempDir()

		got := b.primaryModule(dir, "whatever")
		if got != "" {
			t.Errorf("expected empty result, got %q", got)
		}
	})
}
