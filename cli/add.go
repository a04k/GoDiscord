package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <module-source>",
	Short: "Add a module to an existing bot",
	Long: `Add a module to an existing bot project. The module source can be:
  - A Git repository URL (e.g., github.com/user/bot-module)
  - A local directory path
  - A module name from the registry`,
	Args: cobra.ExactArgs(1),
	Run:  runAdd,
}

func runAdd(cmd *cobra.Command, args []string) {
	moduleSource := args[0]

	// Check if we're in a bot project directory
	if !isBotProject() {
		fmt.Println("Error: Not in a bot project directory")
		fmt.Println("Run 'botcli create <name>' to create a new bot project")
		return
	}

	// Load existing bot configuration
	config, err := loadBotConfig()
	if err != nil {
		fmt.Printf("Error loading bot configuration: %v\n", err)
		return
	}

	fmt.Printf("Adding module: %s\n", moduleSource)

	// Determine module source type and handle accordingly
	var modulePath string

	if isGitURL(moduleSource) {
		// Clone git repository to temporary directory
		modulePath, err = cloneModule(moduleSource)
		if err != nil {
			fmt.Printf("Error cloning module: %v\n", err)
			return
		}
		defer os.RemoveAll(modulePath)
	} else if isLocalPath(moduleSource) {
		// Use local directory
		modulePath = moduleSource
	} else {
		fmt.Printf("Error: Invalid module source '%s'\n", moduleSource)
		fmt.Println("Module source must be a Git URL or local directory path")
		return
	}

	// Discover modules in the source
	modules, err := discoverModulesInPath(modulePath)
	if err != nil {
		fmt.Printf("Error discovering modules: %v\n", err)
		return
	}

	if len(modules) == 0 {
		fmt.Println("No modules found in the specified source")
		return
	}

	// Let user select which modules to add
	selectedModules := selectModulesToAdd(modules, config)

	if len(selectedModules) == 0 {
		fmt.Println("No modules selected")
		return
	}

	// Copy selected modules to bot project
	for _, moduleName := range selectedModules {
		module := modules[moduleName]
		err = copyModuleToProject(modulePath, module)
		if err != nil {
			fmt.Printf("Error copying module %s: %v\n", moduleName, err)
			continue
		}

		// Add to bot configuration
		config.Modules = append(config.Modules, moduleName)

		// Add commands to configuration (enabled by default)
		for _, cmd := range module.Commands {
			config.Commands[cmd.Name] = true
		}

		fmt.Printf("✅ Added module: %s\n", moduleName)
	}

	// Save updated configuration
	err = saveBotConfig(config)
	if err != nil {
		fmt.Printf("Error saving bot configuration: %v\n", err)
		return
	}

	fmt.Println("\n🎉 Modules added successfully!")
	fmt.Println("Run 'go mod tidy' to update dependencies")
	fmt.Println("Run 'go build' to rebuild your bot")
}

func isBotProject() bool {
	_, err := os.Stat("bot.config.json")
	return err == nil
}

func loadBotConfig() (*BotConfig, error) {
	data, err := os.ReadFile("bot.config.json")
	if err != nil {
		return nil, err
	}

	var config BotConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Initialize maps if nil
	if config.Commands == nil {
		config.Commands = make(map[string]bool)
	}
	if config.Config == nil {
		config.Config = make(map[string]interface{})
	}

	return &config, nil
}

func saveBotConfig(config *BotConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("bot.config.json", data, 0644)
}

func isGitURL(source string) bool {
	return strings.HasPrefix(source, "http://") ||
		strings.HasPrefix(source, "https://") ||
		strings.HasPrefix(source, "git@") ||
		strings.Contains(source, "github.com") ||
		strings.Contains(source, "gitlab.com") ||
		strings.Contains(source, "bitbucket.org")
}

func isLocalPath(source string) bool {
	_, err := os.Stat(source)
	return err == nil
}

func cloneModule(gitURL string) (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "botcli-module-*")
	if err != nil {
		return "", err
	}

	// Clone repository
	cmd := exec.Command("git", "clone", gitURL, tempDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}

func discoverModulesInPath(path string) (map[string]ModuleInfo, error) {
	modules := make(map[string]ModuleInfo)

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "module.go" {
			module, err := parseModuleFile(filePath)
			if err != nil {
				fmt.Printf("Warning: Failed to parse module file %s: %v\n", filePath, err)
				return nil
			}

			if module != nil {
				module.Path = filepath.Dir(filePath)
				modules[module.Name] = *module
			}
		}

		return nil
	})

	return modules, err
}

func selectModulesToAdd(modules map[string]ModuleInfo, config *BotConfig) []string {
	fmt.Println("\n📦 Available modules to add:")
	fmt.Println()

	var selected []string

	for name, module := range modules {
		// Check if module is already added
		alreadyAdded := false
		for _, existingModule := range config.Modules {
			if existingModule == name {
				alreadyAdded = true
				break
			}
		}

		if alreadyAdded {
			fmt.Printf("  %s - %s (already added)\n", name, module.Description)
			continue
		}

		fmt.Printf("  %s - %s\n", name, module.Description)
		fmt.Printf("    Commands: %d\n", len(module.Commands))

		if promptYesNo(fmt.Sprintf("Add %s module?", name)) {
			selected = append(selected, name)
		}
		fmt.Println()
	}

	return selected
}

func copyModuleToProject(sourcePath string, module ModuleInfo) error {
	// Determine destination path
	destPath := filepath.Join("commands", strings.ToLower(module.Name))

	// Create destination directory
	err := os.MkdirAll(destPath, 0755)
	if err != nil {
		return err
	}

	// Copy module files
	moduleSourcePath := filepath.Join(sourcePath, module.Path)

	return filepath.Walk(moduleSourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(moduleSourcePath, path)
		if err != nil {
			return err
		}

		// Skip non-Go files and test files
		if !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		// Copy file
		destFile := filepath.Join(destPath, relPath)

		// Create destination directory if needed
		destDir := filepath.Dir(destFile)
		err = os.MkdirAll(destDir, 0755)
		if err != nil {
			return err
		}

		// Copy file content
		sourceData, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destFile, sourceData, info.Mode())
	})
}
