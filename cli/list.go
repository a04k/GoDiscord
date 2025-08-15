package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available modules and commands",
	Long:  `Display all available modules and their commands that can be included in your bot.`,
	Run:   runList,
}

var (
	listModules  bool
	listCommands bool
	filterModule string
)

func init() {
	listCmd.Flags().BoolVarP(&listModules, "modules", "m", false, "List only modules")
	listCmd.Flags().BoolVarP(&listCommands, "commands", "c", false, "List only commands")
	listCmd.Flags().StringVarP(&filterModule, "filter", "f", "", "Filter by module name")
}

func runList(cmd *cobra.Command, args []string) {
	modules, err := discoverModulesSimple()
	if err != nil {
		fmt.Printf("Error discovering modules: %v\n", err)
		return
	}

	if len(modules) == 0 {
		fmt.Println("No modules found. Make sure you're in a bot project directory.")
		return
	}

	// Filter by module if specified
	if filterModule != "" {
		if module, exists := modules[filterModule]; exists {
			modules = map[string]ModuleInfo{filterModule: module}
		} else {
			fmt.Printf("Module '%s' not found\n", filterModule)
			return
		}
	}

	if listModules {
		displayModules(modules)
	} else if listCommands {
		displayCommands(modules)
	} else {
		displayModulesAndCommands(modules)
	}
}

func displayModules(modules map[string]ModuleInfo) {
	fmt.Println("📦 Available Modules:")
	fmt.Println()

	// Sort modules by name
	var names []string
	for name := range modules {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		module := modules[name]
		fmt.Printf("  %s v%s\n", name, module.Version)
		fmt.Printf("    %s\n", module.Description)
		fmt.Printf("    Author: %s\n", module.Author)
		fmt.Printf("    Commands: %d\n", len(module.Commands))
		if len(module.Dependencies) > 0 {
			fmt.Printf("    Dependencies: %s\n", strings.Join(module.Dependencies, ", "))
		}
		fmt.Println()
	}
}

func displayCommands(modules map[string]ModuleInfo) {
	fmt.Println("🔧 Available Commands:")
	fmt.Println()

	// Group commands by category
	categories := make(map[string][]CommandInfo)

	for _, module := range modules {
		for _, cmd := range module.Commands {
			categories[cmd.Category] = append(categories[cmd.Category], cmd)
		}
	}

	// Sort categories
	var categoryNames []string
	for category := range categories {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)

	for _, category := range categoryNames {
		commands := categories[category]
		fmt.Printf("📂 %s\n", category)

		// Sort commands within category
		sort.Slice(commands, func(i, j int) bool {
			return commands[i].Name < commands[j].Name
		})

		for _, cmd := range commands {
			fmt.Printf("  .%s", cmd.Name)
			if len(cmd.Aliases) > 0 {
				fmt.Printf(" (%s)", strings.Join(cmd.Aliases, ", "))
			}
			fmt.Printf(" - %s\n", cmd.Description)
			if cmd.Usage != "" {
				fmt.Printf("    Usage: %s\n", cmd.Usage)
			}
		}
		fmt.Println()
	}
}

func displayModulesAndCommands(modules map[string]ModuleInfo) {
	fmt.Println("📦 Available Modules and Commands:")
	fmt.Println()

	// Sort modules by name
	var names []string
	for name := range modules {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		module := modules[name]
		fmt.Printf("📦 %s v%s - %s\n", name, module.Version, module.Description)
		fmt.Printf("   Author: %s\n", module.Author)

		if len(module.Dependencies) > 0 {
			fmt.Printf("   Dependencies: %s\n", strings.Join(module.Dependencies, ", "))
		}

		fmt.Println("   Commands:")

		// Sort commands
		sort.Slice(module.Commands, func(i, j int) bool {
			return module.Commands[i].Name < module.Commands[j].Name
		})

		for _, cmd := range module.Commands {
			fmt.Printf("     .%s", cmd.Name)
			if len(cmd.Aliases) > 0 {
				fmt.Printf(" (%s)", strings.Join(cmd.Aliases, ", "))
			}
			fmt.Printf(" - %s\n", cmd.Description)
		}
		fmt.Println()
	}

	// Summary
	totalCommands := 0
	for _, module := range modules {
		totalCommands += len(module.Commands)
	}

	fmt.Printf("📊 Summary: %d modules, %d commands\n", len(modules), totalCommands)
}
