package handler

import (
	"fmt"
	"log"

	"github.com/ue555/nvpm/pkg/core/config"
)

// Handler represents a lazy-loading trigger handler
type Handler interface {
	// Name returns the handler name
	Name() string

	// Add adds a trigger for a plugin
	Add(plugin *config.Plugin) error

	// Remove removes a trigger for a plugin
	Remove(plugin *config.Plugin) error

	// Setup sets up the handler
	Setup() error

	// Trigger triggers the handler for a specific value
	Trigger(value string) error
}

// Registry manages all handlers
type Registry struct {
	Config   *config.Config
	handlers map[string]Handler
}

// NewRegistry creates a new handler registry
func NewRegistry(cfg *config.Config) *Registry {
	return &Registry{
		Config:   cfg,
		handlers: make(map[string]Handler),
	}
}

// Init initializes all handlers
func (r *Registry) Init() error {
	// Create handlers
	r.handlers["event"] = NewEventHandler(r.Config, r)
	r.handlers["cmd"] = NewCmdHandler(r.Config, r)
	r.handlers["ft"] = NewFtHandler(r.Config, r)
	r.handlers["keys"] = NewKeysHandler(r.Config, r)

	log.Println("Initialized handlers: event, cmd, ft, keys")
	return nil
}

// Setup sets up all handlers
func (r *Registry) Setup() error {
	for name, handler := range r.handlers {
		log.Printf("Setting up handler: %s\n", name)
		if err := handler.Setup(); err != nil {
			return fmt.Errorf("failed to setup handler %s: %w", name, err)
		}
	}
	return nil
}

// Get retrieves a handler by name
func (r *Registry) Get(name string) (Handler, bool) {
	handler, ok := r.handlers[name]
	return handler, ok
}

// Add adds a trigger to a handler
func (r *Registry) Add(handlerName string, plugin *config.Plugin) error {
	handler, ok := r.Get(handlerName)
	if !ok {
		return fmt.Errorf("handler not found: %s", handlerName)
	}
	return handler.Add(plugin)
}

// Remove removes a trigger from a handler
func (r *Registry) Remove(handlerName string, plugin *config.Plugin) error {
	handler, ok := r.Get(handlerName)
	if !ok {
		return fmt.Errorf("handler not found: %s", handlerName)
	}
	return handler.Remove(plugin)
}

// BaseHandler provides common handler functionality
type BaseHandler struct {
	name     string
	Config   *config.Config
	Registry *Registry
	triggers map[string][]*config.Plugin // value -> plugins
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(name string, cfg *config.Config, registry *Registry) *BaseHandler {
	return &BaseHandler{
		name:     name,
		Config:   cfg,
		Registry: registry,
		triggers: make(map[string][]*config.Plugin),
	}
}

// Name returns the handler name
func (h *BaseHandler) Name() string {
	return h.name
}

// AddTrigger adds a trigger value for a plugin
func (h *BaseHandler) AddTrigger(value string, plugin *config.Plugin) {
	h.triggers[value] = append(h.triggers[value], plugin)
}

// RemoveTrigger removes a trigger value for a plugin
func (h *BaseHandler) RemoveTrigger(value string, plugin *config.Plugin) {
	plugins := h.triggers[value]
	for i, p := range plugins {
		if p.Name == plugin.Name {
			h.triggers[value] = append(plugins[:i], plugins[i+1:]...)
			break
		}
	}
}

// GetPlugins returns plugins for a trigger value
func (h *BaseHandler) GetPlugins(value string) []*config.Plugin {
	return h.triggers[value]
}

// TriggerLoad triggers loading of plugins for a value
func (h *BaseHandler) TriggerLoad(value string) error {
	plugins := h.GetPlugins(value)
	if len(plugins) == 0 {
		return nil
	}

	log.Printf("Handler %s triggered for value: %s (%d plugins)\n", h.name, value, len(plugins))

	// TODO: Call loader.Load() to load the plugins
	// For now, just log
	for _, plugin := range plugins {
		log.Printf("  - Would load plugin: %s\n", plugin.Name)
	}

	return nil
}
