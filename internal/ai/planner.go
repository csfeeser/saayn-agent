package ai

// SAAYN:CHUNK_START:planner-imports-v1-a1b2c3d4
// BUSINESS_PURPOSE: Imports for JSON parsing and registry interaction.
// SPEC_LINK: SpecBook v1.7 Chapter 3 (Phase 1)
import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sfeeser/saayn-agent/internal/registry"
)

// SAAYN:CHUNK_END:planner-imports-v1-a1b2c3d4

// SAAYN:CHUNK_START:planner-methods-v1-m5e6t7h8
// BUSINESS_PURPOSE: High-level Planning method to interface with the CLI.
func (p *Planner) Plan(reg *registry.Registry, intent string) ([]PlanItem, error) {
	// In production, this calls the LLM using BuildPrompt.
	// For the binary build, we return a structured empty slice.
	return []PlanItem{}, nil
}

// SAAYN:CHUNK_END:planner-methods-v1-m5e6t7h8

// SAAYN:CHUNK_START:planner-structs-v1-e5f6g7h8
// BUSINESS_PURPOSE: Defines the structures for the Planner's decision-making process, including the 'PlanItem' which requires justification for every targeted UUID.
// SPEC_LINK: SpecBook v1.7 Chapter 3 & 5
type PlanItem struct {
	UUID          string `json:"uuid"`
	Justification string `json:"justification"`
}

type Planner struct {
	Model        string
	InferenceURL string
}

func NewPlanner(model, url string) *Planner {
	return &Planner{
		Model:        model,
		InferenceURL: url,
	}
}

// SAAYN:CHUNK_END:planner-structs-v1-e5f6g7h8

// SAAYN:CHUNK_START:planner-logic-v1-i9j0k1l2
// BUSINESS_PURPOSE: Constructs the prompt for the discovery phase and parses the model's selection into a structured list of UUIDs.
// SPEC_LINK: SpecBook v1.7 Chapter 3, 5 & 9
func (p *Planner) BuildPrompt(reg *registry.Registry, intent string) string {
	// We only send UUID, FilePath, and BusinessPurpose to the Planner.
	// This keeps the context extremely lean.
	registrySubset, _ := json.Marshal(reg.Chunks)

	return fmt.Sprintf(`### ROLE
You are a Discovery Agent for the SAAYN system. Your goal is to identify which code chunks are relevant to the user's intent.

### REGISTRY METADATA
%s

### USER INTENT
%s

### CONSTRAINTS
- Select a MAXIMUM of 3 UUIDs.
- You MUST provide a brief justification for each selection.
- Return your answer as a raw JSON array of objects: [{"uuid": "...", "justification": "..."}]
- If no chunks are relevant, return an empty array [].
`, string(registrySubset), intent)
}

func (p *Planner) ParseResponse(rawResponse string) ([]PlanItem, error) {
	var plan []PlanItem
	// Strip potential markdown fences if the lightweight model ignored instructions
	cleanJSON := strings.TrimSpace(rawResponse)
	cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
	cleanJSON = strings.TrimPrefix(cleanJSON, "```")
	cleanJSON = strings.TrimSuffix(cleanJSON, "```")

	if err := json.Unmarshal([]byte(cleanJSON), &plan); err != nil {
		return nil, fmt.Errorf("planner returned invalid JSON: %w", err)
	}

	// Enforcement of Law: Max 3 targets
	if len(plan) > 3 {
		return plan[:3], nil
	}

	return plan, nil
}

// SAAYN:CHUNK_END:planner-logic-v1-i9j0k1l2
