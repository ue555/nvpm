package handler

import (
	"log"

	"github.com/ue555/nvpm/pkg/core/config"
)

// EventHandler handles event-based lazy loading
type EventHandler struct {
	*BaseHandler
}

// NewEventHandler creates a new event handler
func NewEventHandler(cfg *config.Config, registry *Registry) *EventHandler {
	return &EventHandler{
		BaseHandler: NewBaseHandler("event", cfg, registry),
	}
}

// Add adds event triggers for a plugin
func (h *EventHandler) Add(plugin *config.Plugin) error {
	for _, event := range plugin.Event {
		log.Printf("Adding event trigger: %s for plugin %s\n", event, plugin.Name)
		h.AddTrigger(event, plugin)
	}
	return nil
}

// Remove removes event triggers for a plugin
func (h *EventHandler) Remove(plugin *config.Plugin) error {
	for _, event := range plugin.Event {
		h.RemoveTrigger(event, plugin)
	}
	return nil
}

// Setup sets up the event handler
func (h *EventHandler) Setup() error {
	// In a real implementation, this would register autocmds with Neovim
	// For now, we just log the registered events
	log.Printf("Event handler setup complete. Registered events:\n")
	for event, plugins := range h.triggers {
		log.Printf("  - %s: %d plugins\n", event, len(plugins))
	}
	return nil
}

// Trigger triggers the event handler
func (h *EventHandler) Trigger(event string) error {
	return h.TriggerLoad(event)
}
