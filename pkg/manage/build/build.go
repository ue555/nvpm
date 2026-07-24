package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ue555/nvpm/pkg/core/config"
)

// Builder runs plugin build commands
type Builder struct {
	Config *config.Config
}

// NewBuilder creates a new build runner
func NewBuilder(cfg *config.Config) *Builder {
	return &Builder{Config: cfg}
}

// Run executes the plugin's build command, if any. Commands prefixed with
// ":" are treated as Neovim Ex commands and run inside a headless, isolated
// Neovim instance with the plugin and its dependencies on the runtimepath.
// Any other command is run as a shell command from within the plugin's
// directory.
func (b *Builder) Run(plugin *config.Plugin) (string, error) {
	cmd := strings.TrimSpace(plugin.Build)
	if cmd == "" {
		return "", nil
	}

	if strings.HasPrefix(cmd, ":") {
		return b.runVimCommand(plugin, strings.TrimPrefix(cmd, ":"))
	}
	return b.runShellCommand(plugin, cmd)
}

// runShellCommand runs cmd via the shell inside the plugin's directory
func (b *Builder) runShellCommand(plugin *config.Plugin, cmd string) (string, error) {
	dir := b.Config.GetPluginDir(plugin.Dir)

	c := exec.Command("sh", "-c", cmd)
	c.Dir = dir
	output, err := c.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("build command failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

// runVimCommand runs an Ex command inside a headless, isolated Neovim
// instance that has the plugin and its dependencies on its runtimepath
func (b *Builder) runVimCommand(plugin *config.Plugin, exCmd string) (string, error) {
	nvimPath, err := exec.LookPath("nvim")
	if err != nil {
		return "", fmt.Errorf("nvim not found in PATH, cannot run build command %q", plugin.Build)
	}

	pluginDir := b.Config.GetPluginDir(plugin.Dir)
	dirs := []string{pluginDir}
	for _, depName := range plugin.Dependencies {
		if dep, ok := b.Config.GetPlugin(depName); ok {
			dirs = append(dirs, b.Config.GetPluginDir(dep.Dir))
		}
	}

	args := []string{
		"--headless",
		"-u", "NONE",
		"-i", "NONE",
		"--cmd", "set rtp^=" + strings.Join(dirs, ","),
	}

	// -u NONE disables 'loadplugins', so plugin/ scripts on the runtimepath
	// are not sourced automatically. Source them explicitly.
	args = append(args, "-c", "runtime! plugin/**/*.vim plugin/**/*.lua")

	// Some plugins only register their Ex commands inside setup() rather than
	// a plugin/ script, so best-effort require+setup the plugin's main module
	// before running the requested command.
	if module := b.primaryModule(pluginDir, plugin.Name); module != "" {
		args = append(args, "-c", fmt.Sprintf("lua pcall(function() require(%q).setup({}) end)", module))
	}

	args = append(args, "-c", exCmd, "-c", "qa!")

	c := exec.Command(nvimPath, args...)
	output, err := c.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("vim build command failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

// primaryModule guesses the Lua module name a plugin should be require()'d
// as, by matching the plugin's name against the top-level entries of its
// lua/ directory (e.g. "mason.nvim" -> lua/mason/ -> "mason"). Returns ""
// if no confident match is found.
func (b *Builder) primaryModule(pluginDir, pluginName string) string {
	entries, err := os.ReadDir(filepath.Join(pluginDir, "lua"))
	if err != nil {
		return ""
	}

	base := strings.ToLower(pluginName)
	base = strings.TrimSuffix(base, ".nvim")
	base = strings.TrimSuffix(base, ".lua")
	base = strings.TrimSuffix(base, ".vim")

	var candidates []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			candidates = append(candidates, name)
		} else if strings.HasSuffix(name, ".lua") {
			candidates = append(candidates, strings.TrimSuffix(name, ".lua"))
		}
	}

	for _, c := range candidates {
		if strings.EqualFold(c, base) {
			return c
		}
	}

	if len(candidates) == 1 {
		return candidates[0]
	}

	return ""
}
