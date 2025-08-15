package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type CommandInfo struct {
	Name        string   `json:"name"`
	Aliases     []string `json:"aliases"`
	Description string   `json:"description"`
	Usage       string   `json:"usage"`
	Category    string   `json:"category"`
}

type ModuleInfo struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	Author       string                 `json:"author"`
	Category     string                 `json:"category"`
	Commands     []CommandInfo          `json:"commands"`
	Dependencies []string               `json:"dependencies"`
	Config       map[string]interface{} `json:"config"`
	Path         string                 `json:"path"` // File system path
}

// discoverModules scans the commands directory for module definitions
func discoverModules() (map[string]ModuleInfo, error) {
	modules := make(map[string]ModuleInfo)

	// Look for module.go files in commands subdirectories
	commandsDir := "commands"

	err := filepath.Walk(commandsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "module.go" {
			module, err := parseModuleFile(path)
			if err != nil {
				fmt.Printf("Warning: Failed to parse module file %s: %v\n", path, err)
				return nil
			}

			if module != nil {
				module.Path = filepath.Dir(path)
				modules[module.Name] = *module
			}
		}

		return nil
	})

	return modules, err
}

// parseModuleFile extracts module information from a module.go file
func parseModuleFile(filename string) (*ModuleInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var module *ModuleInfo

	// Look for init() function with RegisterModule call
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "commands" && sel.Sel.Name == "RegisterModule" {
					// Found RegisterModule call, extract module info
					if len(x.Args) > 0 {
						if comp, ok := x.Args[0].(*ast.UnaryExpr); ok && comp.Op == token.AND {
							if compLit, ok := comp.X.(*ast.CompositeLit); ok {
								module = extractModuleFromLiteral(compLit)
							}
						}
					}
				}
			}
		}
		return true
	})

	return module, nil
}

// extractModuleFromLiteral extracts module information from a composite literal
func extractModuleFromLiteral(lit *ast.CompositeLit) *ModuleInfo {
	module := &ModuleInfo{
		Commands:     []CommandInfo{},
		Dependencies: []string{},
		Config:       make(map[string]interface{}),
	}

	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				switch ident.Name {
				case "Name":
					if basic, ok := kv.Value.(*ast.BasicLit); ok && basic.Kind == token.STRING {
						module.Name = strings.Trim(basic.Value, `"`)
					}
				case "Description":
					if basic, ok := kv.Value.(*ast.BasicLit); ok && basic.Kind == token.STRING {
						module.Description = strings.Trim(basic.Value, `"`)
					}
				case "Version":
					if basic, ok := kv.Value.(*ast.BasicLit); ok && basic.Kind == token.STRING {
						module.Version = strings.Trim(basic.Value, `"`)
					}
				case "Author":
					if basic, ok := kv.Value.(*ast.BasicLit); ok && basic.Kind == token.STRING {
						module.Author = strings.Trim(basic.Value, `"`)
					}
				case "Category":
					if basic, ok := kv.Value.(*ast.BasicLit); ok && basic.Kind == token.STRING {
						module.Category = strings.Trim(basic.Value, `"`)
					}
				case "Commands":
					if compLit, ok := kv.Value.(*ast.CompositeLit); ok {
						// Commands are now []string, not []CommandInfo
						commandNames := extractStringSliceFromLiteral(compLit)
						// Convert to CommandInfo by looking up in CommandDetails
						for _, cmdName := range commandNames {
							// Create a basic CommandInfo - the CLI will need to load the actual details
							module.Commands = append(module.Commands, CommandInfo{
								Name:        cmdName,
								Category:    module.Category, // Use module's category
								Description: "Command from " + module.Name + " module",
							})
						}
					}
				case "Dependencies":
					if compLit, ok := kv.Value.(*ast.CompositeLit); ok {
						module.Dependencies = extractStringSliceFromLiteral(compLit)
					}
				}
			}
		}
	}

	return module
}

// extractCommandsFromLiteral extracts command information from a slice literal
func extractCommandsFromLiteral(lit *ast.CompositeLit) []CommandInfo {
	var commands []CommandInfo

	for _, elt := range lit.Elts {
		if compLit, ok := elt.(*ast.CompositeLit); ok {
			cmd := CommandInfo{
				Aliases: []string{},
			}

			for _, cmdElt := range compLit.Elts {
				if kv, ok := cmdElt.(*ast.KeyValueExpr); ok {
					if ident, ok := kv.Key.(*ast.Ident); ok {
						switch ident.Name {
						case "Name":
							if basic, ok := kv.Value.(*ast.BasicLit); ok && basic.Kind == token.STRING {
								cmd.Name = strings.Trim(basic.Value, `"`)
							}
						case "Description":
							if basic, ok := kv.Value.(*ast.BasicLit); ok && basic.Kind == token.STRING {
								cmd.Description = strings.Trim(basic.Value, `"`)
							}
						case "Usage":
							if basic, ok := kv.Value.(*ast.BasicLit); ok && basic.Kind == token.STRING {
								cmd.Usage = strings.Trim(basic.Value, `"`)
							}
						case "Category":
							if basic, ok := kv.Value.(*ast.BasicLit); ok && basic.Kind == token.STRING {
								cmd.Category = strings.Trim(basic.Value, `"`)
							}
						case "Aliases":
							if compLit, ok := kv.Value.(*ast.CompositeLit); ok {
								cmd.Aliases = extractStringSliceFromLiteral(compLit)
							}
						}
					}
				}
			}

			commands = append(commands, cmd)
		}
	}

	return commands
}

// extractStringSliceFromLiteral extracts a string slice from a composite literal
func extractStringSliceFromLiteral(lit *ast.CompositeLit) []string {
	var result []string

	for _, elt := range lit.Elts {
		if basic, ok := elt.(*ast.BasicLit); ok && basic.Kind == token.STRING {
			result = append(result, strings.Trim(basic.Value, `"`))
		}
	}

	return result
}

// saveModulesCache saves discovered modules to a cache file for faster access
func saveModulesCache(modules map[string]ModuleInfo) error {
	cacheFile := ".botcli-cache.json"
	data, err := json.MarshalIndent(modules, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// loadModulesCache loads modules from cache file
func loadModulesCache() (map[string]ModuleInfo, error) {
	cacheFile := ".botcli-cache.json"
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var modules map[string]ModuleInfo
	err = json.Unmarshal(data, &modules)
	return modules, err
}
