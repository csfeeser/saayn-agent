package cmd

// SAAYN:CHUNK_START:cmd-helpers-v1-h1e2l3p4
// BUSINESS_PURPOSE: Shared utility functions for CLI commands to interact with the registry and file markers.
import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sfeeser/saayn-agent/internal/adapter"
	"github.com/sfeeser/saayn-agent/internal/ai" // <--- Add this
	"github.com/sfeeser/saayn-agent/internal/registry"
)

func loadRegistry() *registry.Registry {
	data, err := os.ReadFile("chunk-registry.json")
	if err != nil {
		return &registry.Registry{Chunks: []registry.Chunk{}}
	}
	var reg registry.Registry
	json.Unmarshal(data, &reg)
	return &reg
}

func saveRegistry(reg *registry.Registry) {
	data, _ := json.MarshalIndent(reg, "", "  ")
	os.WriteFile("chunk-registry.json", data, 0644)
}

func findChunk(reg *registry.Registry, uuid string) registry.Chunk {
	for _, c := range reg.Chunks {
		if c.UUID == uuid {
			return c
		}
	}
	return registry.Chunk{}
}

func extractChunk(content, uuid string, adp adapter.Adapter) (string, int, int, error) {
	lines := strings.Split(content, "\n")
	start, end := -1, -1
	for i, line := range lines {
		if strings.Contains(line, "CHUNK_START:"+uuid) {
			start = i
		}
		if strings.Contains(line, "CHUNK_END:"+uuid) {
			end = i
		}
	}
	if start == -1 || end == -1 {
		return "", 0, 0, fmt.Errorf("markers not found")
	}
	return strings.Join(lines[start+1:end], "\n"), start + 1, end + 1, nil
}

func confirmPlan(plan []ai.PlanItem) {
	fmt.Println("\n--- PROPOSED PLAN ---")
	for _, item := range plan {
		fmt.Printf("📍 Chunk: %s\n💡 Why: %s\n\n", item.UUID, item.Justification)
	}
	fmt.Print("Confirm these changes? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(strings.TrimSpace(response)) != "y" {
		fmt.Println("Aborted.")
		os.Exit(0)
	}
}

// SAAYN:CHUNK_END:cmd-helpers-v1-h1e2l3p4
