package transaction

// SAAYN:CHUNK_START:journal-imports-v1-f9g0h1i2
// BUSINESS_PURPOSE: Standard library imports for durable file operations, JSON serialization, and filesystem-level safety.
// SPEC_LINK: SpecBook v1.7 Chapter 3
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)
// SAAYN:CHUNK_END:journal-imports-v1-f9g0h1i2

// SAAYN:CHUNK_START:journal-structs-v1-j3k4l5m6
// BUSINESS_PURPOSE: Data structures for the Durable Rollback Journal. Tracks every single file operation within a transaction to ensure all-or-nothing atomicity.
// SPEC_LINK: SpecBook v1.7 Chapter 3 & 6
type FileAction struct {
	OriginalPath string `json:"original_path"`
	BackupPath   string `json:"backup_path"`
	ExpectedHash string `json:"expected_hash"`
}

type Journal struct {
	OperationID string       `json:"operation_id"`
	Actions     []FileAction `json:"actions"`
	RegistryTmp string       `json:"registry_tmp"`
}
// SAAYN:CHUNK_END:journal-structs-v1-j3k4l5m6

// SAAYN:CHUNK_START:journal-persistence-v1-n7o8p9q0
// BUSINESS_PURPOSE: Implements fsync-safe persistence of the journal. This is the 'Point of No Return' in a transaction.
// SPEC_LINK: SpecBook v1.7 Chapter 3 & 10 (Law 6)
func (j *Journal) Persist(journalPath string) error {
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}

	// Create journal file with restrictive permissions
	f, err := os.OpenFile(journalPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}

	// MANDATORY: Flush to physical platter before continuing
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to fsync journal: %w", err)
	}

	return nil
}
// SAAYN:CHUNK_END:journal-persistence-v1-n7o8p9q0

// SAAYN:CHUNK_START:journal-recovery-v1-r1s2t3u4
// BUSINESS_PURPOSE: The 'Self-Healing' logic. If a journal exists on startup, it indicates a failed previous run. This function restores files to their original state.
// SPEC_LINK: SpecBook v1.7 Chapter 6 (Step 2)
func Recover(journalPath string) error {
	data, err := os.ReadFile(journalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No journal, nothing to recover
		}
		return err
	}

	var j Journal
	if err := json.Unmarshal(data, &j); err != nil {
		return fmt.Errorf("corrupted journal found: %w", err)
	}

	fmt.Printf("⚠️  [SAAYN] Incomplete transaction %s detected. Rolling back...\n", j.OperationID)

	// Restore originals from backup
	for _, action := range j.Actions {
		if _, err := os.Stat(action.BackupPath); err == nil {
			if err := os.Rename(action.BackupPath, action.OriginalPath); err != nil {
				return fmt.Errorf("failed to restore %s: %w", action.OriginalPath, err)
			}
		}
	}

	// Cleanup
	os.Remove(journalPath)
	os.RemoveAll(filepath.Dir(j.Actions[0].BackupPath))
	
	fmt.Println("✅ Recovery complete. Codebase is back to a clean state.")
	return nil
}
// SAAYN:CHUNK_END:journal-recovery-v1-r1s2t3u4
