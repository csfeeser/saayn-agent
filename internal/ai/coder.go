package ai

// SAAYN:CHUNK_START:coder-imports-v1-k1l2m3n4
// BUSINESS_PURPOSE: Imports for LLM communication, string sanitization, and the adapter system for syntax validation.
// SPEC_LINK: SpecBook v1.7 Chapter 4, 5 & 6
import (
	"fmt"
	"strings"

	"github.com/sfeeser/saayn-agent/internal/adapter"
	"github.com/sfeeser/saayn-agent/internal/registry"
)

// SAAYN:CHUNK_END:coder-imports-v1-k1l2m3n4

// SAAYN:CHUNK_START:coder-methods-v1-c1o2d3e4
// BUSINESS_PURPOSE: Orchestrates the prompt building and sanitization for a single chunk.
func (c *Coder) Generate(chunk registry.Chunk, intent string) (string, error) {
	// This is where the LLM call happens.
	// Returning the original content for now to allow a clean build.
	return "/* Generated Code Placeholder */", nil
}

// SAAYN:CHUNK_END:coder-methods-v1-c1o2d3e4

// SAAYN:CHUNK_START:coder-struct-v1-o5p6q7r8
// BUSINESS_PURPOSE: Defines the Coder agent responsible for synthesizing code changes based on human intent and existing chunk context.
// SPEC_LINK: SpecBook v1.7 Chapter 3 (Phase 2)
type Coder struct {
	Model         string
	PromptVersion string
	InferenceURL  string
}

func NewCoder(model, url string) *Coder {
	return &Coder{
		Model:         model,
		PromptVersion: "v1",
		InferenceURL:  url,
	}
}

// SAAYN:CHUNK_END:coder-struct-v1-o5p6q7r8

// SAAYN:CHUNK_START:coder-sanitize-v1-s9t0u1v2
// BUSINESS_PURPOSE: Implements the Zero-Markdown Protocol. Strips triple-backticks and language identifiers, failing loudly if the response contains conversational filler.
// SPEC_LINK: SpecBook v1.7 Chapter 6
func (c *Coder) Sanitize(rawResponse string, lang adapter.Adapter) (string, error) {
	// 1. Hard Check: No backticks allowed in the final code
	if strings.Contains(rawResponse, "```") {
		// Attempt to extract content between fences if the model failed the prompt
		parts := strings.Split(rawResponse, "```")
		if len(parts) >= 3 {
			// Extract the middle part (ignoring the language hint like 'go' or 'html')
			rawResponse = parts[1]
			// Strip the language hint if it exists (e.g., "go\npackage main")
			lines := strings.SplitN(rawResponse, "\n", 2)
			if len(lines) > 1 && !strings.Contains(lines[0], " ") {
				rawResponse = lines[1]
			}
		} else {
			return "", fmt.Errorf("LLM violated Zero-Markdown protocol: ambiguous backtick usage")
		}
	}

	cleanCode := strings.TrimSpace(rawResponse)
	if cleanCode == "" {
		return "", fmt.Errorf("LLM returned empty code block")
	}

	// 2. Syntax Validation: The "Bouncer" gate
	ok, err := lang.SyntaxCheck(cleanCode)
	if !ok || err != nil {
		// If 'ok' is false, the code is structurally unsound.
		// We wrap the error to provide context to the AI or Human Director.
		return "", fmt.Errorf("syntactic validation failed: %w", err)
	}

	return cleanCode, nil
}

// SAAYN:CHUNK_END:coder-sanitize-v1-s9t0u1v2

// SAAYN:CHUNK_START:coder-prompt-v1-w3x4y5z6
// BUSINESS_PURPOSE: Constructs the deterministic prompt for the Coder model, including the logical unit's purpose and existing content.
// SPEC_LINK: SpecBook v1.7 Chapter 5 (Idempotency)
func (c *Coder) BuildPrompt(chunk registry.Chunk, intent string) string {
	return fmt.Sprintf(`### INSTRUCTION
You are a precision code editor. Your task is to modify the code below based on the user's intent.

### CONTEXT
UUID: %s
Business Purpose: %s
File: %s

### USER INTENT
%s

### TARGET CODE
%s

### CONSTRAINTS
- Return RAW CODE ONLY. 
- Do NOT use markdown backticks (e.g., no %s%s%s).
- Do NOT include explanations or conversational filler.
- Ensure the output is a complete replacement for the TARGET CODE block.
`, chunk.UUID, chunk.BusinessPurpose, chunk.FilePath, intent, "original_content_placeholder", "```", "go", "```")
}

// SAAYN:CHUNK_END:coder-prompt-v1-w3x4y5z6
