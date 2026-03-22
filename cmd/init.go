package cmd

// SAAYN:CHUNK_START:init-imports-v1-a1b2c3d4
// BUSINESS_PURPOSE: Imports for filesystem interaction, environment setup, and registry initialization.
// SPEC_LINK: SpecBook v1.7 Chapter 2 & 5
import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sfeeser/saayn-agent/internal/registry"
	"github.com/spf13/cobra"
)

// SAAYN:CHUNK_END:init-imports-v1-a1b2c3d4

// SAAYN:CHUNK_START:init-command-definition-v1-e5f6g7h8
// BUSINESS_PURPOSE: Defines the 'init' command which prepares a repository for SAAYN orchestration.
// SPEC_LINK: SpecBook v1.7 Chapter 2
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new SAAYN project and create the chunk registry",
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// SAAYN:CHUNK_END:init-command-definition-v1-e5f6g7h8

// SAAYN:CHUNK_START:init-logic-v1-i9j0k1l2
// BUSINESS_PURPOSE: Implements the 'Day 1' setup: creating the .saayn directory, generating the registry, and securing the .env file.
// SPEC_LINK: SpecBook v1.7 Chapter 5 & 9
func runInit() {
	fmt.Println("🏗️  Initializing SAAYN project...")

	// 1. Create .saayn internal directory structure
	paths := []string{
		".saayn/backup",
		".saayn/journal",
	}
	for _, p := range paths {
		if err := os.MkdirAll(p, 0700); err != nil {
			fmt.Printf("❌ Failed to create directory %s: %v\n", p, err)
			return
		}
	}

	// 2. Initialize Empty Registry if not exists
	registryPath := "chunk-registry.json"
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		emptyRegistry := registry.Registry{Chunks: []registry.Chunk{}}
		data, _ := json.MarshalIndent(emptyRegistry, "", "  ")
		if err := os.WriteFile(registryPath, data, 0644); err != nil {
			fmt.Printf("❌ Failed to create registry: %v\n", err)
			return
		}
		fmt.Println("📄 Created chunk-registry.json")
	}

	// 3. Setup .env template (12-Factor Config)
	envPath := ".env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		template := `# SAAYN Configuration
SAAYN_INFERENCE_URL="http://localhost:8000/v1"
SAAYN_PLANNER_MODEL="llama-3-8b"
SAAYN_CODER_MODEL="qwen-2.5-coder-32b"
SAAYN_AUTO_COMMIT=false
`
		os.WriteFile(envPath, []byte(template), 0600)
		fmt.Println("🔐 Created .env template (ensure you add your keys/endpoints)")
	}

	// 4. Safety: Ensure .env and .saayn/ are in .gitignore
	ensureGitIgnore()

	fmt.Println("✅ SAAYN initialization complete. Start adding chunks to your files!")
}

func ensureGitIgnore() {
	giPath := ".gitignore"
	entry := "\n# SAAYN internals\n.env\n.saayn/\n"

	f, err := os.OpenFile(giPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(entry)
}

// SAAYN:CHUNK_END:init-logic-v1-i9j0k1l2
