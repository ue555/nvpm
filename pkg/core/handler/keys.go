package handler

import (
	"log"

	"github.com/kouji/nvpm/pkg/core/config"
)

// KeysHandler handles key mapping-based lazy loading
type KeysHandler struct {
	*BaseHandler
}

// NewKeysHandler creates a new keys handler
func NewKeysHandler(cfg *config.Config, registry *Registry) *KeysHandler {
	return &KeysHandler{
		BaseHandler: NewBaseHandler("keys", cfg, registry),
	}
}

// Add adds key triggers for a plugin
func (h *KeysHandler) Add(plugin *config.Plugin) error {
	for _, key := range plugin.Keys {
		log.Printf("Adding key trigger: %s for plugin %s\n", key, plugin.Name)
		h.AddTrigger(key, plugin)
	}
	return nil
}

// Remove removes key triggers for a plugin
func (h *KeysHandler) Remove(plugin *config.Plugin) error {
	for _, key := range plugin.Keys {
		h.RemoveTrigger(key, plugin)
	}
	return nil
}

// Setup sets up the keys handler
func (h *KeysHandler) Setup() error {
	// In a real implementation, this would create lazy key mappings
	log.Printf("Keys handler setup complete. Registered keys:\n")
	for key, plugins := range h.triggers {
		log.Printf("  - %s: %d plugins\n", key, len(plugins))
	}
	return nil
}

// Trigger triggers the keys handler
func (h *KeysHandler) Trigger(key string) error {
	return h.TriggerLoad(key)
}
