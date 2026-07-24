package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ue555/nvpm/pkg/core/config"
)

// Git manages git operations for plugins
type Git struct {
	Config *config.Config
}

// NewGit creates a new git manager
func NewGit(cfg *config.Config) *Git {
	return &Git{
		Config: cfg,
	}
}

// Clone clones a git repository
func (g *Git) Clone(plugin *config.Plugin) error {
	url := g.normalizeURL(plugin.URL)
	dir := g.Config.GetPluginDir(plugin.Dir)

	// Check if already exists
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("plugin directory already exists: %s", dir)
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Build clone command
	args := []string{"clone"}

	// Add filter for partial clone
	if g.Config.Git.Filter != "" {
		args = append(args, "--filter="+g.Config.Git.Filter)
	}

	args = append(args, url, dir)

	// Execute clone
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// Fetch fetches updates from remote
func (g *Git) Fetch(plugin *config.Plugin) error {
	dir := g.Config.GetPluginDir(plugin.Dir)

	cmd := exec.Command("git", "-C", dir, "fetch", "--tags", "--force")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// Checkout checks out a specific branch, tag, or commit
func (g *Git) Checkout(plugin *config.Plugin, ref string) error {
	dir := g.Config.GetPluginDir(plugin.Dir)

	cmd := exec.Command("git", "-C", dir, "checkout", ref)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// Pull pulls updates from remote
func (g *Git) Pull(plugin *config.Plugin) error {
	dir := g.Config.GetPluginDir(plugin.Dir)

	cmd := exec.Command("git", "-C", dir, "pull", "--ff-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// GetCurrentCommit returns the current commit hash
func (g *Git) GetCurrentCommit(plugin *config.Plugin) (string, error) {
	dir := g.Config.GetPluginDir(plugin.Dir)

	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch returns the current branch name
func (g *Git) GetCurrentBranch(plugin *config.Plugin) (string, error) {
	dir := g.Config.GetPluginDir(plugin.Dir)

	cmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetRemoteURL returns the remote URL
func (g *Git) GetRemoteURL(plugin *config.Plugin, remote string) (string, error) {
	dir := g.Config.GetPluginDir(plugin.Dir)

	cmd := exec.Command("git", "-C", dir, "remote", "get-url", remote)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetLog returns git log entries
func (g *Git) GetLog(plugin *config.Plugin, args ...string) ([]string, error) {
	dir := g.Config.GetPluginDir(plugin.Dir)

	cmdArgs := []string{"-C", dir, "log"}
	cmdArgs = append(cmdArgs, g.Config.Git.Log...)
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return lines, nil
}

// GetInfo returns git repository information
func (g *Git) GetInfo(plugin *config.Plugin) (*config.GitInfo, error) {
	info := &config.GitInfo{
		Remotes: make(map[string]string),
	}

	// Get current commit
	commit, err := g.GetCurrentCommit(plugin)
	if err != nil {
		return nil, err
	}
	info.Commit = commit

	// Get current branch
	branch, err := g.GetCurrentBranch(plugin)
	if err != nil {
		// Not on a branch (detached HEAD)
		branch = ""
	}
	info.Branch = branch

	// Get remote URL
	url, err := g.GetRemoteURL(plugin, "origin")
	if err == nil {
		info.Remotes["origin"] = url
	}

	return info, nil
}

// HasUpdates checks if there are updates available
func (g *Git) HasUpdates(plugin *config.Plugin) (bool, error) {
	dir := g.Config.GetPluginDir(plugin.Dir)

	// Fetch first
	if err := g.Fetch(plugin); err != nil {
		return false, err
	}

	// Get local and remote commits
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	localOutput, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get local commit: %w", err)
	}
	localCommit := strings.TrimSpace(string(localOutput))

	cmd = exec.Command("git", "-C", dir, "rev-parse", "@{u}")
	remoteOutput, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get remote commit: %w", err)
	}
	remoteCommit := strings.TrimSpace(string(remoteOutput))

	return localCommit != remoteCommit, nil
}

// Update updates a plugin to the latest version
func (g *Git) Update(plugin *config.Plugin) error {
	// Fetch updates
	if err := g.Fetch(plugin); err != nil {
		return err
	}

	// Determine what to checkout
	var ref string
	if plugin.Commit != "" {
		ref = plugin.Commit
	} else if plugin.Tag != "" {
		ref = plugin.Tag
	} else if plugin.Branch != "" {
		ref = plugin.Branch
	} else {
		// Use default branch
		ref = "HEAD"
	}

	// Checkout the ref
	if err := g.Checkout(plugin, ref); err != nil {
		return err
	}

	// If on a branch, pull
	if plugin.Branch != "" {
		if err := g.Pull(plugin); err != nil {
			return err
		}
	}

	return nil
}

// normalizeURL converts short URLs to full git URLs
func (g *Git) normalizeURL(url string) string {
	// If it's already a full URL, return as-is
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "git@") {
		return url
	}

	// Otherwise, assume it's a GitHub short name (e.g., "folke/lazy.nvim")
	return fmt.Sprintf(g.Config.Git.URLFormat, url)
}

// Exists checks if a git repository exists
func (g *Git) Exists(plugin *config.Plugin) bool {
	dir := g.Config.GetPluginDir(plugin.Dir)
	gitDir := filepath.Join(dir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// Remove removes a git repository
func (g *Git) Remove(plugin *config.Plugin) error {
	dir := g.Config.GetPluginDir(plugin.Dir)
	return os.RemoveAll(dir)
}

// GetLastUpdateTime returns the last update time of the repository
func (g *Git) GetLastUpdateTime(plugin *config.Plugin) (time.Time, error) {
	dir := g.Config.GetPluginDir(plugin.Dir)

	cmd := exec.Command("git", "-C", dir, "log", "-1", "--format=%ct")
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last update time: %w", err)
	}

	timestamp := strings.TrimSpace(string(output))
	var unixTime int64
	fmt.Sscanf(timestamp, "%d", &unixTime)

	return time.Unix(unixTime, 0), nil
}
