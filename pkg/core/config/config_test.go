package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectOrphans(t *testing.T) {
	root := t.TempDir()

	dirs := []string{"telescope.nvim", "cache", "old-plugin"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "stray.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("failed to create stray file: %v", err)
	}

	cfg := &Config{
		Root: root,
		Performance: PerformanceConfig{
			Cache: CacheConfig{Path: filepath.Join(root, "cache")},
		},
		Plugins: map[string]*Plugin{
			"telescope.nvim": {Name: "telescope.nvim", Dir: "telescope.nvim"},
		},
	}

	orphans, err := cfg.DetectOrphans()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orphans) != 1 {
		t.Fatalf("expected 1 orphan, got %d: %+v", len(orphans), orphans)
	}
	if orphans[0].Name != "old-plugin" || orphans[0].Dir != "old-plugin" {
		t.Errorf("unexpected orphan: %+v", orphans[0])
	}
}

func TestDetectOrphansNoUnusedPlugins(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "telescope.nvim"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		Root: root,
		Plugins: map[string]*Plugin{
			"telescope.nvim": {Name: "telescope.nvim", Dir: "telescope.nvim"},
		},
	}

	orphans, err := cfg.DetectOrphans()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orphans) != 0 {
		t.Errorf("expected no orphans, got %+v", orphans)
	}
}

func TestDetectOrphansMissingRoot(t *testing.T) {
	cfg := &Config{
		Root:    filepath.Join(t.TempDir(), "does-not-exist"),
		Plugins: map[string]*Plugin{},
	}

	orphans, err := cfg.DetectOrphans()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orphans != nil {
		t.Errorf("expected nil orphans, got %+v", orphans)
	}
}
