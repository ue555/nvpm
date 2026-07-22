package handler

import (
	"log"

	"github.com/kouji/nvpm/pkg/core/config"
)

// CmdHandler handles command-based lazy loading
type CmdHandler struct {
	*BaseHandler
}

// NewCmdHandler creates a new command handler
func NewCmdHandler(cfg *config.Config, registry *Registry) *CmdHandler {
	return &CmdHandler{
		BaseHandler: NewBaseHandler("cmd", cfg, registry),
	}
}

// Add adds command triggers for a plugin
func (h *CmdHandler) Add(plugin *config.Plugin) error {
	for _, cmd := range plugin.Cmd {
		log.Printf("Adding command trigger: %s for plugin %s\n", cmd, plugin.Name)
		h.AddTrigger(cmd, plugin)
	}
	return nil
}

// Remove removes command triggers for a plugin
func (h *CmdHandler) Remove(plugin *config.Plugin) error {
	for _, cmd := range plugin.Cmd {
		h.RemoveTrigger(cmd, plugin)
	}
	return nil
}

// Setup sets up the command handler
func (h *CmdHandler) Setup() error {
	// In a real implementation, this would create placeholder commands
	log.Printf("Command handler setup complete. Registered commands:\n")
	for cmd, plugins := range h.triggers {
		log.Printf("  - :%s: %d plugins\n", cmd, len(plugins))
	}
	return nil
}

// Trigger triggers the command handler
func (h *CmdHandler) Trigger(cmd string) error {
	return h.TriggerLoad(cmd)
}
