package transaction

// SAAYN:CHUNK_START:engine-imports-v1-a1b2c3d4
// BUSINESS_PURPOSE: Standard library and internal package imports for transaction orchestration, including the adapter and registry systems.
// SPEC_LINK: SpecBook v1.7 Chapter 4 & 6
import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"saayn/internal/adapter"
	"saayn/internal/registry"
)
// SAAYN:CHUNK_END:engine-imports-v1-a1b2c3d4

// SAAYN:CHUNK_START:engine-struct-v1-e5f6g7h8
// BUSINESS_PURPOSE: Defines the Engine which holds the state of a single transaction, including the lock status and the active registry.
// SPEC_LINK: SpecBook v1.7 Chapter 0 & 3
type Engine struct {
	Registry    *registry.Registry
	OperationID string
	WorkDir     string
	LockFile    *os.File
}

func NewEngine(workDir string, opID string) *Engine {
	return &Engine{
		OperationID: opID,
		WorkDir:     workDir,
	}
}
// SAAYN:CHUNK_END:engine-struct-v1-e5f6g7h8

// SAAYN:CHUNK_START:engine-lifecycle-v1-i9j0k1l2
// BUSINESS_PURPOSE: Implements the Lock and Recovery logic. Ensures no two agents run at once and restores state from the Journal if a crash is detected.
// SPEC_LINK: SpecBook v1.7 Chapter 6 (Steps 1 & 2)
func (e *Engine) AcquireLock() error {
	lockPath := filepath.Join(e.WorkDir, ".saayn.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("could not acquire lock: %w (is another saayn process running?)", err)
	}
	e.LockFile = f
	
	// Step 2: Immediate Recovery Check
	journalPath := filepath.Join(e.WorkDir, ".saayn", "journal", e.OperationID+".json")
	return Recover(journalPath)
}

func (e *Engine) ReleaseLock() {
	if e.LockFile != nil {
		e.LockFile.Close()
		os.Remove(filepath.Join(e.WorkDir, ".saayn.lock"))
	}
}
// SAAYN:CHUNK_END:engine-lifecycle-v1-i9j0k1l2

// SAAYN:CHUNK_START:engine-atomic-commit-v1-m3n4o5p6
// BUSINESS_PURPOSE: The "Point of No Return." Executes the 16-step commit pipeline: staging, journaling, backing up, and renaming.
// SPEC_LINK: SpecBook v1.7 Chapter 3 & 6 (Steps 8-16)
func (e *Engine) Commit(actions []FileAction) error {
	// 1. Stage .tmp files and registry.tmp (Must fsync)
	// 2. Write Journal (Must fsync)
	journalPath := filepath.Join(e.WorkDir, ".saayn", "journal", e.OperationID+".json")
	journal := Journal{
		OperationID: e.OperationID,
		Actions:     actions,
	}
	if err := journal.Persist(journalPath); err != nil {
		return err
	}

	// 3. Backup originals to .saayn/backup/<op_id>/
	backupDir := filepath.Join(e.WorkDir, ".saayn", "backup", e.OperationID)
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return err
	}

	for _, action := range actions {
		backupPath := filepath.Join(backupDir, filepath.Base(action.OriginalPath))
		if err := os.Rename(action.OriginalPath, backupPath); err != nil {
			return fmt.Errorf("backup failed for %s: %w", action.OriginalPath, err)
		}
	}

	// 4. Apply: Rename .tmp to original
	for _, action := range actions {
		tmpPath := action.OriginalPath + ".tmp"
		if err := os.Rename(tmpPath, action.OriginalPath); err != nil {
			// Trigger Recovery Chapter 6, Step 12
			return fmt.Errorf("apply failed: %w (Run saayn to recover)", err)
		}
	}

	// 5. Finalize: Registry commit + Cleanup
	// (Registry logic goes here)
	
	os.Remove(journalPath)
	os.RemoveAll(backupDir)
	return nil
}
// SAAYN:CHUNK_END:engine-atomic-commit-v1-m3n4o5p6
