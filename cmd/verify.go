package cmd

// SAAYN:CHUNK_START:verify-imports-v1-v1w2x3y4
// BUSINESS_PURPOSE: Imports for registry access, file reading, and structured JSON output for observability.
// SPEC_LINK: SpecBook v1.7 Chapter 7 & 9
import (
	"fmt"
	"os"
	"strconv"

	"github.com/sfeeser/saayn-agent/internal/adapter"
	"github.com/sfeeser/saayn-agent/internal/registry"
	"github.com/spf13/cobra"
)

// SAAYN:CHUNK_END:verify-imports-v1-v1w2x3y4

// SAAYN:CHUNK_START:verify-command-definition-v1-z5a6b7c8
// BUSINESS_PURPOSE: Defines the 'verify' command which audits the codebase for cryptographic drift.
// SPEC_LINK: SpecBook v1.7 Chapter 9
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Audit the codebase to detect drift between registry and physical files",
	Run: func(cmd *cobra.Command, args []string) {
		runVerify()
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}

// SAAYN:CHUNK_END:verify-command-definition-v1-z5a6b7c8

// SAAYN:CHUNK_START:verify-logic-v2-d9e0f1g2
// BUSINESS_PURPOSE: Implements the verification loop. Checks every chunk for existence, marker integrity, and content drift using cryptographic binding.
// SPEC_LINK: SpecBook v1.7 Chapter 7 & 10 (Verify Rule)
func runVerify() {
	reg := loadRegistry()
	fmt.Printf("🔍 Auditing %d chunks...\n", len(reg.Chunks))

	allSync := true
	for _, chunk := range reg.Chunks {
		status := "SYNC"
		details := ""

		// 1. File Existence Check (Using modern 'os' instead of 'ioutil')
		content, err := os.ReadFile(chunk.FilePath)
		if err != nil {
			status = "MISSING"
			details = fmt.Sprintf("File %s not found", chunk.FilePath)
		} else {
			// 2. Marker & Content Extraction
			adp, _ := adapter.Get(chunk.LanguageHint)
			extracted, startLine, endLine, err := extractChunk(string(content), chunk.UUID, adp)

			if err != nil {
				status = "CORRUPTED"
				details = err.Error()
			} else {
				// 3. Cryptographic Hash Validation
				// Convert int to string using strconv.Itoa to match hashing signature
				currentContentHash := registry.ComputeContentHash(extracted)
				currentMarkerHash := registry.ComputeMarkerHash(
					chunk.UUID,
					strconv.Itoa(startLine),
					strconv.Itoa(endLine),
				)

				if currentMarkerHash != chunk.MarkerHash {
					status = "MODIFIED"
					details = "Markers have been moved or tampered with"
				} else if currentContentHash != chunk.ContentHash {
					status = "MODIFIED"
					details = "Content drifted from registry"
				}
			}
		}

		if status != "SYNC" {
			allSync = false
			fmt.Printf("❌ [%s] %s: %s\n", status, chunk.UUID, details)
		}
	}

	if allSync {
		fmt.Println("✅ All chunks are synchronized and cryptographically valid.")
	} else {
		fmt.Println("\n⚠️  Drift detected. Use 'saayn reconcile' to update the registry or revert manual changes.")
		os.Exit(1)
	}
}

// SAAYN:CHUNK_END:verify-logic-v2-d9e0f1g2
