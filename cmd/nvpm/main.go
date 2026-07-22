package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kouji/nvpm/pkg/core/config"
	"github.com/kouji/nvpm/pkg/nvpm"
)

func main() {
	// Define flags
	configFile := flag.String("config", "", "Path to configuration file")
	command := flag.String("cmd", "", "Command to execute (install, update, clean, sync, check, restore, stats)")
	plugin := flag.String("plugin", "", "Plugin name (optional, for specific plugin operations)")
	flag.Parse()

	// Create nvpm instance
	l := nvpm.New()

	// Load configuration from file if specified
	var specs []interface{}
	var opts *config.Config

	if *configFile != "" {
		data, err := os.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("Failed to read config file: %v", err)
		}

		var cfg map[string]interface{}
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("Failed to parse config file: %v", err)
		}

		if s, ok := cfg["plugins"].([]interface{}); ok {
			specs = s
		}
	}

	// Example default configuration
	if len(specs) == 0 {
		specs = []interface{}{
			"folke/tokyonight.nvim",
			"nvim-telescope/telescope.nvim",
			"neovim/nvim-lspconfig",
		}
	}

	// Setup nvpm
	if err := l.Setup(specs, opts); err != nil {
		log.Fatalf("Setup failed: %v", err)
	}

	// Execute command
	if *command != "" {
		if err := executeCommand(l, *command, *plugin); err != nil {
			log.Fatalf("Command failed: %v", err)
		}
	} else {
		// Default: print stats
		l.PrintStats()
	}
}

func executeCommand(l *nvpm.NVPM, cmd string, pluginName string) error {
	var err error

	switch cmd {
	case "install":
		if pluginName != "" {
			if plugin, ok := l.GetPlugin(pluginName); ok {
				err = l.Install(plugin)
			} else {
				return fmt.Errorf("plugin not found: %s", pluginName)
			}
		} else {
			err = l.Install()
		}

	case "update":
		if pluginName != "" {
			if plugin, ok := l.GetPlugin(pluginName); ok {
				err = l.Update(plugin)
			} else {
				return fmt.Errorf("plugin not found: %s", pluginName)
			}
		} else {
			err = l.Update()
		}

	case "clean":
		err = l.Clean()

	case "sync":
		err = l.Sync()

	case "check":
		err = l.Check()

	case "restore":
		err = l.Restore()

	case "stats":
		l.PrintStats()

	case "list":
		fmt.Println("\nInstalled Plugins:")
		for name, plugin := range l.Plugins() {
			status := "not installed"
			if plugin.Installed {
				status = "installed"
			}
			if plugin.Loaded {
				status += " (loaded)"
			}
			fmt.Printf("  - %s: %s\n", name, status)
		}

	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}

	return err
}
