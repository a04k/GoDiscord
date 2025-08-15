package commands

import (
	"DiscordBot/bot"

	"github.com/bwmarrin/discordgo"
)

// CommandFunc defines the signature for command handlers
type CommandFunc func(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string)

// CommandInfo holds detailed information about a command
type CommandInfo struct {
	Name        string   `json:"name"`
	Aliases     []string `json:"aliases"`
	Description string   `json:"description"`
	Usage       string   `json:"usage"`
	Category    string   `json:"category"`
}

// SlashCommandInfo holds information about slash commands
type SlashCommandInfo struct {
	Name        string                                                           `json:"name"`
	Description string                                                           `json:"description"`
	Options     []*discordgo.ApplicationCommandOption                            `json:"options"`
	Handler     func(*bot.Bot, *discordgo.Session, *discordgo.InteractionCreate) `json:"-"`
}

// ModuleInfo represents a complete module with its commands and metadata
type ModuleInfo struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	Author        string                 `json:"author"`
	Category      string                 `json:"category"`
	Commands      []CommandInfo          `json:"commands"`       // Full command info
	SlashCommands []SlashCommandInfo     `json:"slash_commands"` // Slash commands in this module
	Dependencies  []string               `json:"dependencies"`   // Other modules this depends on
	Config        map[string]interface{} `json:"config"`         // Default config for this module
}

// CategoryInfo represents a category that contains multiple modules
type CategoryInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Modules     []string `json:"modules"` // List of module names in this category
}

// Global registries
var (
	RegisteredModules    = make(map[string]*ModuleInfo)
	RegisteredCategories = make(map[string]*CategoryInfo)
	CommandDetails       = make(map[string]CommandInfo)                                                      // Auto-compiled from modules
	SlashCommandHandlers = make(map[string]func(*bot.Bot, *discordgo.Session, *discordgo.InteractionCreate)) // Auto-compiled slash handlers
	CommandMap           = make(map[string]CommandFunc)                                                      // Legacy command map
	CommandAliases       = make(map[string]string)                                                           // Command aliases
)

// RegisterCommand registers individual commands (used by modules)
func RegisterCommand(name string, handler CommandFunc, aliases ...string) {
	CommandMap[name] = handler
	for _, alias := range aliases {
		CommandAliases[alias] = name
	}
}

// RegisterModule registers a complete module and auto-compiles command info
func RegisterModule(module *ModuleInfo) {
	RegisteredModules[module.Name] = module

	// Auto-compile command info from module
	for _, cmd := range module.Commands {
		CommandDetails[cmd.Name] = cmd
	}

	// Auto-compile slash command handlers
	for _, slashCmd := range module.SlashCommands {
		SlashCommandHandlers[slashCmd.Name] = slashCmd.Handler
	}

	// Auto-register category if it doesn't exist
	if module.Category != "" {
		if _, exists := RegisteredCategories[module.Category]; !exists {
			RegisteredCategories[module.Category] = &CategoryInfo{
				Name:        module.Category,
				Description: module.Category + " related modules",
				Modules:     []string{},
			}
		}

		// Add module to category if not already there
		category := RegisteredCategories[module.Category]
		found := false
		for _, modName := range category.Modules {
			if modName == module.Name {
				found = true
				break
			}
		}
		if !found {
			category.Modules = append(category.Modules, module.Name)
		}
	}
}

// GetModuleByCommand returns the module that contains a specific command
func GetModuleByCommand(commandName string) *ModuleInfo {
	for _, module := range RegisteredModules {
		for _, cmd := range module.Commands {
			if cmd.Name == commandName {
				return module
			}
		}
	}
	return nil
}

// GetCommandsByModule returns all commands in a specific module
func GetCommandsByModule(moduleName string) []CommandInfo {
	if module, exists := RegisteredModules[moduleName]; exists {
		return module.Commands
	}
	return []CommandInfo{}
}

// GetModulesByCategory returns all modules in a specific category
func GetModulesByCategory(categoryName string) []*ModuleInfo {
	var modules []*ModuleInfo
	for _, module := range RegisteredModules {
		if module.Category == categoryName {
			modules = append(modules, module)
		}
	}
	return modules
}

// GetAllModules returns all registered modules
func GetAllModules() map[string]*ModuleInfo {
	return RegisteredModules
}

// GetAllCategories returns all registered categories
func GetAllCategories() map[string]*CategoryInfo {
	return RegisteredCategories
}

// GetCommandsByCategory returns all commands in a specific category using registered modules
func GetCommandsByCategory(category string) []CommandInfo {
	var commands []CommandInfo
	for _, module := range RegisteredModules {
		if module.Category == category {
			commands = append(commands, module.Commands...)
		}
	}
	return commands
}

// GetAllSlashCommands returns all registered slash commands for registration
func GetAllSlashCommands() []*discordgo.ApplicationCommand {
	var commands []*discordgo.ApplicationCommand
	for _, module := range RegisteredModules {
		for _, slashCmd := range module.SlashCommands {
			commands = append(commands, &discordgo.ApplicationCommand{
				Name:        slashCmd.Name,
				Description: slashCmd.Description,
				Options:     slashCmd.Options,
			})
		}
	}
	return commands
}
