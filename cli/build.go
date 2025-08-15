package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the bot binary",
	Long:  `Build the bot into an executable binary.`,
	Run:   runBuild,
}

var (
	outputName string
	buildOS    string
	buildArch  string
)

func init() {
	buildCmd.Flags().StringVarP(&outputName, "output", "o", "", "Output binary name (default: bot name)")
	buildCmd.Flags().StringVar(&buildOS, "os", "", "Target OS (linux, windows, darwin)")
	buildCmd.Flags().StringVar(&buildArch, "arch", "", "Target architecture (amd64, arm64)")
}

func runBuild(cmd *cobra.Command, args []string) {
	// Check if we're in a bot project directory
	if !isBotProject() {
		fmt.Println("Error: Not in a bot project directory")
		fmt.Println("Run 'botcli create <name>' to create a new bot project")
		return
	}

	// Load bot configuration
	config, err := loadBotConfig()
	if err != nil {
		fmt.Printf("Error loading bot configuration: %v\n", err)
		return
	}

	// Set default output name
	if outputName == "" {
		outputName = config.Name
		if buildOS == "windows" {
			outputName += ".exe"
		}
	}

	fmt.Printf("Building bot: %s\n", config.Name)

	// Prepare build command
	buildArgs := []string{"build", "-o", outputName}

	// Set environment variables for cross-compilation
	env := os.Environ()
	if buildOS != "" {
		env = append(env, "GOOS="+buildOS)
	}
	if buildArch != "" {
		env = append(env, "GOARCH="+buildArch)
	}

	// Run go build
	buildCmd := exec.Command("go", buildArgs...)
	buildCmd.Env = env
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	err = buildCmd.Run()
	if err != nil {
		fmt.Printf("Build failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully built: %s\n", outputName)

	if buildOS == "" && buildArch == "" {
		fmt.Println("\nTo run your bot:")
		fmt.Printf("  ./%s\n", outputName)
	}
}
