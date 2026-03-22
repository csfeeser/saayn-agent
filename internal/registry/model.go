package registry

// SAAYN:CHUNK_START:registry-imports-v1-a1b2c3d4
// BUSINESS_PURPOSE: Standard library imports for cryptographic hashing, hex encoding, and time formatting.
// SPEC_LINK: SpecBook v1.7 Chapter 1
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// SAAYN:CHUNK_END:registry-imports-v1-a1b2c3d4

// SAAYN:CHUNK_START:registry-struct-definitions-v1-e5f6g7h8
// BUSINESS_PURPOSE: Core data structures for the UPC Barcode system. Defines the 'Chunk' and 'Registry' types that map the physical file state to the JSON inventory.
// SPEC_LINK: SpecBook v1.7 Chapter 2
type Chunk struct {
	UUID            string    `json:"uuid"`
	FilePath        string    `json:"file_path"`
	LanguageHint    string    `json:"language_hint"`
	BusinessPurpose string    `json:"business_purpose"`
	ContentHash     string    `json:"content_hash"`
	MarkerHash      string    `json:"marker_hash"`
	Version         int       `json:"version"`
	LineSpan        LineSpan  `json:"line_span"`
	OrderIndex      int       `json:"order_index"`
	LastModified    time.Time `json:"last_modified"`
}

type LineSpan struct {
	Start      int    `json:"start"`
	End        int    `json:"end"`
	Confidence string `json:"confidence"` // "high", "low"
}

type Registry struct {
	ProjectRoot string  `json:"project_root"`
	Chunks      []Chunk `json:"chunks"` // Must remain a slice for ordering
}

// SAAYN:CHUNK_END:registry-struct-definitions-v1-e5f6g7h8

// SAAYN:CHUNK_START:registry-hashing-logic-v2-i9j0k1l2
// BUSINESS_PURPOSE: Cryptographic verification logic. Computes SHA-256 hashes for both chunk content and the markers themselves to detect unauthorized manual drift.
// SPEC_LINK: SpecBook v1.7 Chapter 1 & 4

func ComputeContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// ComputeMarkerHash now accepts uuid to bind the location to the specific chunk identity
func ComputeMarkerHash(uuid, startLine, endLine string) string {
	combined := fmt.Sprintf("%s|%s|%s", uuid, startLine, endLine)
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

func (c *Chunk) ValidateInvariant() error {
	if c.UUID == "" || c.FilePath == "" {
		return fmt.Errorf("chunk %s missing critical identity fields", c.UUID)
	}
	return nil
}

// SAAYN:CHUNK_END:registry-hashing-logic-v2-i9j0k1l2
