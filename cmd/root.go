package cmd

// SAAYN:CHUNK_START:root-imports-v1-r1o2o3t4
// BUSINESS_PURPOSE: Standard CLI imports and environment loader for 12-factor compliance.
// SPEC_LINK: SpecBook v1.7 Chapter 5
import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)
// SAAYN:CHUNK_END:root-imports-v1-r1o2o3t4

// SAAYN:CHUNK_START:root-definition-v1-d5e6f7g8
// BUSINESS_PURPOSE: Defines the root 'saayn' command and the PersistentPreRun logic for config loading.
// SPEC_LINK: SpecBook v1.7 Chapter 0 & 5
var rootCmd = &cobra.Command{
	Use:   "saayn",
	Short: "SAAYN: The UPC Barcode System for AI-Native Codebases",
	Long: `A sovereign, high-integrity CLI tool for managing AI-driven code edits 
using cryptographic chunking and durable transaction journals.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load .env before any subcommand executes
		if err := godotenv.Load(); err != nil {
			// We don't hard-fail here because env vars might be set in the shell
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
// SAAYN:CHUNK_END:root-definition-v1-d5e6f7g8

// SAAYN:CHUNK_START:root-global-flags-v1-h9i0j1k2
// BUSINESS_PURPOSE: Global flags available to all subcommands.
// SPEC_LINK: SpecBook v1.7 Chapter 5
func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose structured logging")
}
// SAAYN:CHUNK_END:root-global-flags-v1-h9i0j1k2
