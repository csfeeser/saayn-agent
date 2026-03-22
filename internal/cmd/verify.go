package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/sfeeser/saayn-agent/internal/adapter"
	"github.com/sfeeser/saayn-agent/internal/registry"
)

// SAAYN:CHUNK_START:verify-command-v1.8-v5e6r7i8
// BUSINESS_PURPOSE: High-level audit tool. Ensures disk content matches the Registry's "Source of Truth."
// SPEC_LINK: SpecBook v1.8 Chapter 7 (Audit & Compliance)
func RunVerify() {
	fmt.Println("🔍 Auditing the Sovereign Market...")

	reg, err := registry.Load(".saayn/chunk-registry.json")
	if err != nil {
		fmt.Printf("❌ REGISTRY_ERROR: Could not load registry: %v\n", err)
		return
	}

	issuesFound := 0

	for _, chunk := range reg.Chunks {
		fmt.Printf("📦 Checking Chunk: %s [%s]\n", chunk.UUID, chunk.FilePath)

		// 1. Check if file exists
		content, err := os.ReadFile(chunk.FilePath)
		if err != nil {
			fmt.Printf("   ❌ FILE_MISSING: %s\n", chunk.FilePath)
			issuesFound++
			continue
		}

		// 2. Extract the actual content between markers
		adp, _ := adapter.Get(chunk.LanguageHint)
		startMarker, endMarker := adapter.MarkerPattern(adp, chunk.UUID)

		actualCode, found := extractChunk(string(content), startMarker, endMarker)
		if !found {
			fmt.Printf("   ❌ MARKER_LOST: Could not find markers for %s\n", chunk.UUID)
			issuesFound++
			continue
		}

		// 3. Verify Hash Integrity
		hash := sha256.Sum256([]byte(strings.TrimSpace(actualCode)))
		currentHash := hex.EncodeToString(hash[:])

		if currentHash != chunk.ContentHash {
			fmt.Printf("   ❌ DRIFT_DETECTED: Content of %s has changed outside of SAAYN!\n", chunk.UUID)
			fmt.Printf("      Expected: %s...\n", chunk.ContentHash[:8])
			fmt.Printf("      Found:    %s...\n", currentHash[:8])
			issuesFound++
		} else {
			fmt.Printf("   ✅ Integrity Verified.\n")
		}
	}

	if issuesFound == 0 {
		fmt.Println("\n✨ ALL CLEAR: Codebase is synchronized with the Sovereign Registry.")
	} else {
		fmt.Printf("\n⚠️  AUDIT FAILED: %d issues detected. The shelf is untidy.\n", issuesFound)
	}
}

// Helper to find text between two markers
func extractChunk(fileContent, start, end string) (string, bool) {
	startIndex := strings.Index(fileContent, start)
	endIndex := strings.Index(fileContent, end)

	if startIndex == -1 || endIndex == -1 || startIndex >= endIndex {
		return "", false
	}

	// Move start index to the end of the start marker line
	actualStart := startIndex + len(start)
	return fileContent[actualStart:endIndex], true
}

// SAAYN:CHUNK_END:verify-command-v1.8-v5e6r7i8
