package plugin

import (
	"testing"

	"github.com/kouji/nvpm/pkg/core/config"
)

func TestExtractName(t *testing.T) {
	cfg := config.DefaultConfig()
	spec := NewSpec(cfg)

	tests := []struct {
		url      string
		expected string
	}{
		{"folke/lazy.nvim", "lazy.nvim"},
		{"https://github.com/folke/lazy.nvim.git", "lazy.nvim"},
		{"https://github.com/folke/lazy.nvim", "lazy.nvim"},
		{"nvim-telescope/telescope.nvim", "telescope.nvim"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := spec.extractName(tt.url)
			if result != tt.expected {
				t.Errorf("extractName(%s) = %s; want %s", tt.url, result, tt.expected)
			}
		})
	}
}

func TestParseStringSpec(t *testing.T) {
	cfg := config.DefaultConfig()
	spec := NewSpec(cfg)

	url := "folke/tokyonight.nvim"
	err := spec.parseStringSpec(url)

	if err != nil {
		t.Fatalf("parseStringSpec failed: %v", err)
	}

	plugin, ok := spec.Plugins["tokyonight.nvim"]
	if !ok {
		t.Fatal("Plugin not found in spec")
	}

	if plugin.Name != "tokyonight.nvim" {
		t.Errorf("Name = %s; want tokyonight.nvim", plugin.Name)
	}

	if plugin.URL != url {
		t.Errorf("URL = %s; want %s", plugin.URL, url)
	}

	if !plugin.Lazy {
		t.Error("Lazy should be true by default")
	}
}

func TestParseTableSpec(t *testing.T) {
	cfg := config.DefaultConfig()
	spec := NewSpec(cfg)

	tableSpec := map[string]interface{}{
		"url":    "hrsh7th/nvim-cmp",
		"lazy":   false,
		"event":  []string{"InsertEnter"},
		"branch": "main",
	}

	err := spec.parseTableSpec(tableSpec)
	if err != nil {
		t.Fatalf("parseTableSpec failed: %v", err)
	}

	plugin, ok := spec.Plugins["nvim-cmp"]
	if !ok {
		t.Fatal("Plugin not found in spec")
	}

	if plugin.Name != "nvim-cmp" {
		t.Errorf("Name = %s; want nvim-cmp", plugin.Name)
	}

	if plugin.Lazy {
		t.Error("Lazy should be false")
	}

	if len(plugin.Event) != 1 || plugin.Event[0] != "InsertEnter" {
		t.Errorf("Event = %v; want [InsertEnter]", plugin.Event)
	}

	if plugin.Branch != "main" {
		t.Errorf("Branch = %s; want main", plugin.Branch)
	}
}
