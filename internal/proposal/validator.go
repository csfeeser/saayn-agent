package proposal
package proposal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/sfeeser/saayn-agent/internal/registry"
	"github.com/sfeeser/saayn-agent/internal/transaction"
)

// SAAYN:CHUNK_START:proposal-validator-v1.8-v2a3l4i5
// BUSINESS_PURPOSE: Machine-validates a Change Proposal before it reaches human review. Provides cryptographic seal verification and registry lookups.
// SPEC_LINK: SpecBook v1.8 Chapter 6.3
type Validator struct {
	Registry *registry.Registry
}

func NewValidator(reg *registry.Registry) *Validator {
	return &Validator{Registry: reg}
}

// ValidateProposal runs the full gauntlet of Chapter 6.3 checks.
func (v *Validator) ValidateProposal(p *transaction.Proposal) (bool, []string) {
	var logs []string

	// 1. Verify the Seal (Tamper Detection)
	isValidSeal, err := v.VerifySeal(p)
	if err != nil || !isValidSeal {
		logs = append(logs, "❌ SEAL_ERROR: Proposal seal is invalid or tampered.")
		return false, logs
	}

	// 2. Validate Operations
	for _, op := range p.Human.Operations {
		switch op.Type {
		case "CHUNK_REPLACE", "CHUNK_DELETE":
			if !v.uuidExists(op.TargetUUID) {
				logs = append(logs, fmt.Sprintf("❌ UUID_NOT_FOUND: %s does not exist in registry.", op.TargetUUID))
			}
		case "CHUNK_CREATE":
			if op.AfterUUID != "" && !v.uuidExists(op.AfterUUID) {
				logs = append(logs, fmt.Sprintf("❌ ANCHOR_NOT_FOUND: Insertion point %s does not exist.", op.AfterUUID))
			}
			if v.uuidExists(op.TargetUUID) {
				logs = append(logs, fmt.Sprintf("❌ DUPLICATE_UUID: Cannot create %s, it already exists.", op.TargetUUID))
			}
		default:
			logs = append(logs, fmt.Sprintf("❌ INVALID_OP: Unknown operation type '%s'.", op.Type))
		}
	}

	if len(logs) > 0 {
		return false, logs
	}

	logs = append(logs, "✅ Validation successful. Proposal is structurally sound.")
	return true, logs
}

// VerifySeal re-computes the hash of the Saayn section to ensure it hasn't been edited.
func (v *Validator) VerifySeal(p *transaction.Proposal) (bool, error) {
    // Note: json.Marshal in Go generally sorts map keys alphabetically, 
    // which gives us basic canonicalization. 
    data, err := json.Marshal(p.Saayn)
    if err != nil {
        return false, fmt.Errorf("failed to marshal saayn section: %w", err)
    }

    hash := sha256.Sum256(data)
    computedSeal := hex.EncodeToString(hash[:])
    
    // We compare what's in the 'Seal' field with what we just computed
    return p.Seal == computedSeal, nil
}

// uuidExists is a simple check, but let's make a more useful version 
// that returns the chunk itself for further validation (like LanguageHint).
func (v *Validator) getChunk(uuid string) (registry.Chunk, bool) {
    for _, c := range v.Registry.Chunks {
        if c.UUID == uuid {
            return c, true
        }
    }
    return registry.Chunk{}, false
}

func (v *Validator) uuidExists(uuid string) bool {
    _, exists := v.getChunk(uuid)
    return exists
}
// SAAYN:CHUNK_END:proposal-validator-v1.8-v2a3l4i5