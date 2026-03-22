package adapter

// SAAYN:CHUNK_START:adapter-interface-v1-a1b2c3d4
// BUSINESS_PURPOSE: Defines the contract for language-specific logic. Ensures the agent can identify markers, validate syntax, and format code across different file types.
// SPEC_LINK: SpecBook v1.7 Chapter 4
import (
	"fmt"
)

// Adapter defines the behavior for different programming and markup languages.
type Adapter interface {
	Name() string
	CommentPrefix() string
	// Updated to v1.8: returns bool for explicit pass/fail + error for details
	SyntaxCheck(code string) (bool, error)
	Format(code string) (string, error)
}

// SAAYN:CHUNK_END:adapter-interface-v1-a1b2c3d4

// SAAYN:CHUNK_START:adapter-registry-v1-e5f6g7h8
// BUSINESS_PURPOSE: Manages the collection of available adapters and provides a lookup mechanism based on file extensions or language hints.
// SPEC_LINK: SpecBook v1.7 Chapter 1 & 4
var adapters = make(map[string]Adapter)

// Register adds a new language adapter to the global registry.
func Register(a Adapter) {
	adapters[a.Name()] = a
}

// Get finds the appropriate adapter for a given language hint or file extension.
func Get(hint string) (Adapter, error) {
	if a, ok := adapters[hint]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("no SAAYN adapter found for language: %s", hint)
}

// MarkerPattern returns the full START/END marker strings for a specific adapter.
func MarkerPattern(a Adapter, uuid string) (start, end string) {
	prefix := a.CommentPrefix()
	start = fmt.Sprintf("%s SAAYN:CHUNK_START:%s", prefix, uuid)
	end = fmt.Sprintf("%s SAAYN:CHUNK_END:%s", prefix, uuid)
	return start, end
}

// SAAYN:CHUNK_END:adapter-registry-v1-e5f6g7h8
