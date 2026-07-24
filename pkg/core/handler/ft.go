package handler

import (
	"log"

	"github.com/ue555/nvpm/pkg/core/config"
)

// FtHandler handles filetype-based lazy loading
type FtHandler struct {
	*BaseHandler
}

// NewFtHandler creates a new filetype handler
func NewFtHandler(cfg *config.Config, registry *Registry) *FtHandler {
	return &FtHandler{
		BaseHandler: NewBaseHandler("ft", cfg, registry),
	}
}

// Add adds filetype triggers for a plugin
func (h *FtHandler) Add(plugin *config.Plugin) error {
	for _, ft := range plugin.Ft {
		log.Printf("Adding filetype trigger: %s for plugin %s\n", ft, plugin.Name)
		h.AddTrigger(ft, plugin)
	}
	return nil
}

// Remove removes filetype triggers for a plugin
func (h *FtHandler) Remove(plugin *config.Plugin) error {
	for _, ft := range plugin.Ft {
		h.RemoveTrigger(ft, plugin)
	}
	return nil
}

// Setup sets up the filetype handler
func (h *FtHandler) Setup() error {
	// In a real implementation, this would register FileType autocmds
	log.Printf("Filetype handler setup complete. Registered filetypes:\n")
	for ft, plugins := range h.triggers {
		log.Printf("  - %s: %d plugins\n", ft, len(plugins))
	}
	return nil
}

// Trigger triggers the filetype handler
func (h *FtHandler) Trigger(ft string) error {
	return h.TriggerLoad(ft)
}
