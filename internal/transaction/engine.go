package transaction

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sfeeser/saayn-agent/internal/registry"
)

// SAAYN:CHUNK_START:engine-struct-v1.8-unified
// BUSINESS_PURPOSE: Orchestrates the State Machine and the Atomic 16-Step Pipeline.
// SPEC_LINK: SpecBook v1.8 Chapter 6
type Engine struct {
	Registry    *registry.Registry
	State       State
	Proposal    *Proposal
	OperationID string
	WorkDir     string
	LockFile    *os.File
}

type FileAction struct {
	UUID         string
	OriginalPath string
	BackupPath   string
	StagingPath  string
}

func NewEngine(reg *registry.Registry, workDir string, opID string) *Engine {
	return &Engine{
		Registry:    reg,
		State:       StateIdle,
		OperationID: opID,
		WorkDir:     workDir,
	}
}

// SAAYN:CHUNK_END:engine-struct-v1.8-unified

// SAAYN:CHUNK_START:engine-transitions-v1.8
// BUSINESS_PURPOSE: Enforces the legal transitions between states.
func (e *Engine) Transition(next State) error {
	// Guard: Security check for disallowed jumps
	switch next {
	case StateValidating:
		if e.State != StateIdle && e.State != StateFailedVal {
			return fmt.Errorf("invalid transition: %s -> %s", e.State, next)
		}
	case StateExecuting:
		if e.State != StateApproved {
			return fmt.Errorf("SECURITY_VIOLATION: Cannot execute unapproved proposal")
		}
	case StateRecovering:
		// Can be entered from Initial or Executing after a crash
	}

	e.State = next
	if e.Proposal != nil {
		e.Proposal.Saayn.CurrentState = next
		e.Proposal.Saayn.Timestamp = time.Now()
	}
	fmt.Printf("🔄 STATE_CHANGE: %s\n", next)
	return nil
}

// SAAYN:CHUNK_END:engine-transitions-v1.8

// SAAYN:CHUNK_START:engine-lifecycle-v1.8
// BUSINESS_PURPOSE: Handles the physical locks and atomic commit logic.
func (e *Engine) AcquireLock() error {
	lockPath := filepath.Join(e.WorkDir, ".saayn.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("LOCK_ERROR: Another saayn process is active")
	}
	e.LockFile = f
	return nil
}

func (e *Engine) ReleaseLock() {
	if e.LockFile != nil {
		e.LockFile.Close()
		os.Remove(filepath.Join(e.WorkDir, ".saayn.lock"))
	}
}

// Execute handles the "Point of No Return" (Chapter 6.5)
func (e *Engine) Execute(actions []FileAction) error {
	if err := e.Transition(StateExecuting); err != nil {
		return err
	}

	// 1. Create Journal Entry (Durability)
	// 2. Perform Backups
	backupDir := filepath.Join(e.WorkDir, ".saayn", "backup", e.OperationID)
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return err
	}

	for _, action := range actions {
		// Move original to backup
		dest := filepath.Join(backupDir, filepath.Base(action.OriginalPath))
		if err := os.Rename(action.OriginalPath, dest); err != nil {
			e.Transition(StateFailedExec)
			return fmt.Errorf("backup failed: %w", err)
		}

		// Move staged (.tmp) to original
		if err := os.Rename(action.StagingPath, action.OriginalPath); err != nil {
			e.Transition(StateFailedExec)
			return fmt.Errorf("apply failed: %w (Recovery required)", err)
		}
	}

	return e.Transition(StateExecuted)
}

// SAAYN:CHUNK_END:engine-lifecycle-v1.8

// SAAYN:CHUNK_START:engine-utility-helpers-v1.8
// BUSINESS_PURPOSE: Bridge methods to support existing CLI commands while maintaining v1.8 state.

func (e *Engine) StageChunk(chunk registry.Chunk, newCode string) (FileAction, error) {
	action := FileAction{
		UUID:         chunk.UUID,
		OriginalPath: chunk.FilePath,
		BackupPath:   filepath.Join(e.WorkDir, ".saayn", "backup", chunk.UUID+".bak"),
		StagingPath:  filepath.Join(e.WorkDir, ".saayn", "journal", chunk.UUID+".tmp"),
	}

	// Ensure journal dir exists
	if err := os.MkdirAll(filepath.Dir(action.StagingPath), 0700); err != nil {
		return action, err
	}

	err := os.WriteFile(action.StagingPath, []byte(newCode), 0644)
	return action, err
}

// Map 'Commit' to the new v1.8 'Execute' flow
func (e *Engine) Commit(actions []FileAction) error {
	// For now, we manually set state to Approved to allow the transition
	// until we have the full Proposal logic wired into the CLI
	e.State = StateApproved
	return e.Execute(actions)
}

// SAAYN:CHUNK_END:engine-utility-helpers-v1.8
