package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// SimpleModuleDiscovery uses the new modular system to discover modules
func discoverModulesSimple() (map[string]ModuleInfo, error) {
	modules := make(map[string]ModuleInfo)

	// Scan for module.go files and extract module info directly
	err := filepath.Walk("commands", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "module.go" {
			moduleInfo, err := parseModuleFile(path)
			if err != nil {
				fmt.Printf("Warning: Failed to parse module file %s: %v\n", path, err)
				return nil
			}

			if moduleInfo != nil {
				moduleInfo.Path = filepath.Dir(path)
				modules[moduleInfo.Name] = *moduleInfo
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan modules: %v", err)
	}

	return modules, nil
}

// These functions are no longer needed since we use the modular system
