package transaction

import (
	"time"
)

// SAAYN:CHUNK_START:state-definitions-v1.8-s1t2a3t4
// BUSINESS_PURPOSE: Defines the finite states of the SAAYN execution engine.
// SPEC_LINK: SpecBook v1.8 Chapter 6
type State string

const (
	StateInitial         State = "INITIAL"
	StateIdle            State = "IDLE"
	StateValidating      State = "VALIDATING"
	StateValidated       State = "VALIDATED"
	StateFailedVal       State = "FAILED_VALIDATION"
	StatePendingApproval State = "PENDING_APPROVAL"
	StateApproved        State = "APPROVED"
	StateRejected        State = "REJECTED"
	StateExecuting       State = "EXECUTING"
	StateExecuted        State = "EXECUTED"
	StateFailedExec      State = "FAILED_EXECUTION"
	StateRecovering      State = "RECOVERING"
	StateUndoing         State = "UNDOING"
	StateUndone          State = "UNDONE"
)

// Proposal represents the Change Proposal Format (CPF) v1.0
// SPEC_LINK: SpecBook v1.8 Chapter 5
type Proposal struct {
	Human HumanSection `json:"human"`
	Saayn SaaynSection `json:"saayn"`
	Seal  string       `json:"seal"` // SHA-256 of the SaaynSection
}

type HumanSection struct {
	Intent     string      `json:"intent"`
	Operations []Operation `json:"operations"`
}

type SaaynSection struct {
	OpID          string    `json:"op_id"`
	CurrentState  State     `json:"state"`
	ValidationLog []string  `json:"validation_log"`
	ApprovedBy    string    `json:"approved_by,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	ProtocolVer   string    `json:"protocol_version"`
}

type Operation struct {
	Type            string `json:"type"` // CHUNK_REPLACE, CHUNK_CREATE, etc.
	TargetUUID      string `json:"target_uuid"`
	ReplacementCode string `json:"replacement_code,omitempty"`
	AfterUUID       string `json:"after_uuid,omitempty"` // For CHUNK_CREATE
}

// SAAYN:CHUNK_END:state-definitions-v1.8-s1t2a3t4
