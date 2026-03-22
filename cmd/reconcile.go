package cmd

// SAAYN:CHUNK_START:reconcile-imports-v1-a1b2c3d4
// BUSINESS_PURPOSE: Imports for terminal I/O, registry management, and hashing.
// SPEC_LINK: SpecBook v1.7 Chapter 9
import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sfeeser/saayn-agent/internal/adapter"
	"github.com/sfeeser/saayn-agent/internal/registry"
	"github.com/spf13/cobra"
)

// SAAYN:CHUNK_END:reconcile-imports-v1-a1b2c3d4

// SAAYN:CHUNK_START:reconcile-command-definition-v1-e5f6g7h8
// BUSINESS_PURPOSE: Defines the 'reconcile' command which provides a UI for resolving cryptographic drift.
// SPEC_LINK: SpecBook v1.7 Chapter 9
var reconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Interactively update the registry to match manual code changes",
	Run: func(cmd *cobra.Command, args []string) {
		runReconcile()
	},
}

func init() {
	rootCmd.AddCommand(reconcileCmd)
}

// SAAYN:CHUNK_END:reconcile-command-definition-v1-e5f6g7h8

// SAAYN:CHUNK_START:reconcile-logic-v2-r1e2c3o4
// BUSINESS_PURPOSE: Scans files, assigns OrderIndex, and populates LineSpan for v1.8 compliance.
func runReconcile() {
	reg := loadRegistry()
	reader := bufio.NewReader(os.Stdin)
	updated := false

	// We'll build a fresh slice to ensure the ORDER is exactly what's on disk
	var orderedChunks []registry.Chunk
	currentIndex := 0

	fmt.Println("🔄 Reconciling Inventory (v1.8 Order-Aware Scan)...")

	// For each file in the registry (simplified for the demo)
	// In a full version, we'd walk the directory, but for now, we use existing paths
	for _, chunk := range reg.Chunks {
		content, err := os.ReadFile(chunk.FilePath)
		if err != nil {
			fmt.Printf("⚠️  Skipping %s: %v\n", chunk.FilePath, err)
			continue
		}

		adp, _ := adapter.Get(chunk.LanguageHint)
		extracted, startLine, endLine, err := extractChunk(string(content), chunk.UUID, adp)

		if err != nil {
			fmt.Printf("❌ Failed to extract %s: %v\n", chunk.UUID, err)
			continue
		}

		// Calculate the new v1.8 metadata
		newContentHash := registry.ComputeContentHash(extracted)
		newMarkerHash := registry.ComputeMarkerHash(
			chunk.UUID,
			strconv.Itoa(startLine),
			strconv.Itoa(endLine),
		)

		// Create the updated chunk with Order and Span
		updatedChunk := chunk
		updatedChunk.ContentHash = newContentHash
		updatedChunk.MarkerHash = newMarkerHash
		updatedChunk.OrderIndex = currentIndex
		updatedChunk.LineSpan = registry.LineSpan{
			Start:      startLine,
			End:        endLine,
			Confidence: "high",
		}

		// Detect if something actually changed
		if newContentHash != chunk.ContentHash || newMarkerHash != chunk.MarkerHash {
			fmt.Printf("\n📢 Drift in [%s]\n   Old: %s\n   New: %s\n   Update? (y/N): ", chunk.UUID, chunk.ContentHash[:8], newContentHash[:8])
			response, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(response)) == "y" {
				updatedChunk.Version++
				updatedChunk.LastModified = time.Now()
				updated = true
			} else {
				// If they say no, we keep the OLD hashes but still update order/span
				updatedChunk = chunk
			}
		}

		orderedChunks = append(orderedChunks, updatedChunk)
		currentIndex++
	}

	if updated || len(orderedChunks) > 0 {
		reg.Chunks = orderedChunks // The list is now physically ordered!
		saveRegistry(reg)
		fmt.Println("\n💾 Registry synchronized and re-ordered.")
	}
}

// SAAYN:CHUNK_END:reconcile-logic-v2-r1e2c3o4
