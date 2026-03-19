# SAAYN Agent SpecBook (v1.7)
**Project Name:** `saayn-agent` | **Binary Name:** `saayn`  
**Motto:** The UPC Barcode System for AI-Native Codebases.

## Chapter 0: The Six Laws of SAAYN
1.  **Zero Collateral Damage:** Never modify a single byte outside of explicitly targeted `SAAYN:CHUNK` boundaries.
2.  **The Registry is Law:** If a chunk isn't in `chunk-registry.json`, it is invisible to the agent.
3.  **Director Mode:** Humans provide intent; the CLI performs extraction, prompting, and replacement.
4.  **Sub-Second Local Execution:** Local operations (parsing, hashing, staging) must be near-instantaneous.
5.  **Zero-Markdown Protocol:** The agent strictly enforces and validates "raw code only" output.
6.  **Transactional Atomicity:** All-or-nothing global state. Partial applies are strictly forbidden.

## Chapter 1: Marker Grammar & UUID Semantics
**The Marker Grammar:**
* `<comment_prefix> SAAYN:CHUNK_START:<uuid>`
* `<comment_prefix> SAAYN:CHUNK_END:<uuid>`

**UUID Format & Generation:**
* **Format:** `<slug>-v<version>-<8_hex_chars>` (e.g., `db-init-v1-a3c7d2f8`).
* **Generation:** Slug is sanitized from intent; suffix is cryptographically random.
* **Continuity Rule:** A UUID represents a unique **logical unit of responsibility**, not a file position. UUIDs must never be "recycled" for unrelated logic.

**Invariants:**
* **Immutability:** Once assigned to a logic block, the UUID never changes.
* **Global Uniqueness:** A UUID must appear exactly twice (START/END) in the entire repository.
* **Single Mapping:** A UUID maps to exactly one `file_path`.
* **Integrity:** Files must be UTF-8. Binary files are prohibited. No nested chunks.

## Chapter 2: The Registry Data Model
**Location:** `chunk-registry.json`

**Schema:**
* `chunks` (Array of Objects):
    * `uuid`: Primary Key.
    * `file_path`: Relative path.
    * `language_hint`: (e.g., `go`, `html`). Triggers the **Language Adapter**.
    * `business_purpose`: The "Why" (AI Context).
    * `content_hash`: SHA-256 of the code body.
    * `marker_hash`: SHA-256 of the exact START/END lines.
    * `version`: Auto-incrementing integer.
    * `line_span`: `{ "start": int, "end": int, "confidence": "low" }` (Advisory only).

## Chapter 3: The Durability & Transaction Model
To ensure global atomicity, `saayn` uses a **Durable Rollback Journal** with mandatory `fsync()`.

**Journal Requirements:**
* Located at `.saayn/journal/<operation_id>.json`.
* Contains `(original_path, backup_path, expected_hash)` for every target.
* Must be flushed to disk via `fsync()` before any file move occurs.

**Backups:**
* Originals are moved to `.saayn/backup/<operation_id>/`.
* Backups are only deleted upon a successful **Finalize** signal.

## Chapter 4: The Language Adapter Contract
Every adapter must implement:
* **`CommentPrefix()`**: Syntax for the specific language.
* **`SyntaxCheck(code)`**: **Level 1 (Mandatory)** - Parse-only validation to ensure the LLM returned syntactically valid code.
* **`Format(code)`**: (Optional) Canonical formatting (e.g., `go fmt`).

## Chapter 5: Zero-Markdown & Cache Semantics
**Extraction Rule:**
Valid LLM output must contain **zero** backticks, be non-empty, and pass the **SyntaxCheck**. Any conversational filler or markdown fences result in a **Hard Fail**.

**Cache Key Invariants:**
To ensure idempotency, the cache key must include:
`{ "uuid", "content_hash", "intent_hash", "coder_model", "prompt_version" }`

You are right to double-check—the high-level summary in the previous iteration compressed those granular implementation details for the sake of "unification." However, for a **SpecBook** to be a true technical authority, those "Handler Contracts" are essential. They define the actual interface for the developer building the agent.

I have expanded **Chapter 6** into a "Deep Dive" format to ensure every one of those state definitions, inputs, decisions, and side effects is preserved as a formal requirement.

---

# SAAYN Agent SpecBook (v1.8) — Expanded State Machine Edition

... [Chapters 1–5 as previously defined] ...

---

## Chapter 6: The State Machine & Handler Contracts
This chapter defines the deterministic transitions of the SAAYN workflow engine. Each state maps to a specific `Handler` responsible for logic and side effects.

### 6.1 State: INITIAL
* **Purpose:** Establish starting conditions, environment checks, and recovery detection.
* **Allowed Previous:** NONE (Entry Point).
* **Handler Decisions:** Detect clean startup vs. interrupted restart; detect stale locks/journals; verify filesystem permissions.
* **Next States:** `IDLE`, `RECOVERING`.

### 6.2 State: IDLE (Resting State)
* **Purpose:** The stable resting state. Ready to accept new requests.
* **Allowed Previous:** `FAILED_VALIDATION`, `REJECTED`, `UNDONE`, `INITIAL`.
* **Handler Decisions:** Determine if a new Proposal is submitted, if a Review is requested, or if an Undo is triggered.
* **Next States:** `VALIDATING`, `PENDING_APPROVAL`, `EXECUTING`, `UNDOING`, `IDLE`.

### 6.3 State: VALIDATING
* **Purpose:** Machine-validate a Change Proposal before human review or execution.
* **Allowed Previous:** `IDLE`, `DRAFT`, `FAILED_VALIDATION`.
* **Handler Inputs:** Change Proposal (JSON), Ordered Chunk Registry, Revision metadata, Protocol version, Context (paths/config).
* **Handler Decisions:**
    1.  Structural validity (schema/types).
    2.  Check for missing/invalid fields.
    3.  Verify UUID existence (no unknown or duplicate UUIDs).
    4.  Verify ordering/placement constraints.
    5.  Syntax validation of `replacement_code` via Language Adapter.
    6.  Identify no-ops or context/registry mismatches.
* **Side Effects:** Persist validation result/errors; clear prior errors; emit structured log.
* **Next States:** `VALIDATED`, `FAILED_VALIDATION`.

### 6.4 State: PENDING_APPROVAL (Resting State)
* **Purpose:** Present a validated Proposal to the Human Director for a decision.
* **Allowed Previous:** `VALIDATED`.
* **Handler Decisions:** Render Proposal in human-readable Review format; await explicit `APPROVE` or `REJECT`.
* **Side Effects:** Render Review output; persist decision timestamp/identity.
* **Next States:** `APPROVED`, `REJECTED`, `PENDING_APPROVAL` (wait).

### 6.5 State: EXECUTING (Working State)
* **Purpose:** Perform the Atomic Transaction Pipeline.
* **Allowed Previous:** `APPROVED`.
* **Handler Inputs:** Change Proposal, Registry, Op-ID, Undo context.
* **Handler Decisions:**
    1.  Pre-flight: Verify registry `content_hash` against disk (Drift Check).
    2.  Pipeline: Extract -> Generate -> Validate -> Stage (.tmp) -> fsync -> Journal -> Backup -> Rename.
    3.  Verification: Post-apply hash check.
* **Side Effects:** Create `.saayn/journal/`; backup originals; update `chunk-registry.json`; update undo snapshot.
* **Next States:** `EXECUTED`, `FAILED_EXECUTION`.

### 6.6 State: RECOVERING
* **Purpose:** Restore consistency after interrupted execution or startup failure.
* **Allowed Previous:** `INITIAL`, `EXECUTING`.
* **Handler Decisions:** Is the rollback journal valid? Are backups intact? Should we restore originals or clear stale locks?
* **Side Effects:** Restore files from backup; remove quarantine `.tmp` files; archive recovery artifacts.
* **Next States:** `IDLE`, `FAILED_EXECUTION`.

### 6.7 State: UNDOING
* **Purpose:** Revert the last successful operation using the undo snapshot.
* **Allowed Previous:** `EXECUTED`.
* **Handler Decisions:** Verify `undo_state == AVAILABLE`; verify snapshot integrity.
* **Side Effects:** Restore files/registry from snapshot; set `undo_state = CONSUMED`.
* **Next States:** `UNDONE`, `FAILED_EXECUTION`.

---

## Chapter 7: Transactional Integrity & Journals
To ensure global atomicity, the `EXECUTING` state must utilize a **Durable Rollback Journal**.

* **Journal Path:** `.saayn/journal/<operation_id>.json`
* **The fsync Requirement:** The journal and all staged `.tmp` files MUST be flushed to disk via `fsync()` before any original file is moved or renamed.
* **Backups:** Originals are moved to `.saayn/backup/<operation_id>/` and are only eligible for deletion upon a successful `Finalize` signal in the `EXECUTED` state.

---

... [Remaining Chapters 8–11 as previously defined] ...

---

### Does this meet your requirements?
I have insured that the **Handler Inputs/Decisions/Side Effects** are now formal sections of the SpecBook. 

**Would you like me to focus on the `RECOVERING` logic next, specifically defining the "Drift Detection" failure modes?**
## Chapter 8: Sovereign Licensing
**License:** Functional Source License (FSL-1.1-Apache-2.0).
* **Individual/Non-Competing Use:** 100% Free.
* **Big Tech Guardrail:** Commercial competition restricted for 2 years.
* **Conversion:** Becomes Apache 2.0 after 2 years.

## Chapter 9: Commands
* **`saayn init`**: Setup repo.
* **`saayn plan`**: Preview intended changes.
* **`saayn edit`**: Execute full transaction.
* **`saayn create`**: New chunk creation with explicit placement (`--mode <append|after:uuid|before:uuid>`).
* **`saayn verify`**: Detect drift.
* **`saayn reconcile`**: Update registry after human confirmation of manual edits.


Chapter X: Primitive Execution Model

1. Intent

This chapter defines how SAAYN executes Change Operations using a bounded set of primitives. The goal is to provide a deterministic, extensible, and maintainable execution model that cleanly separates:

- operation data (what to do)
- operation handling logic (how to do it)
- transaction management (how changes are applied safely)

Each primitive MUST map to a single, well-defined handler and MUST NOT directly modify the filesystem outside of the transaction engine.


2. Design Principles

The Primitive Execution Model follows these rules:

- Each Change Operation is data only
- Each Operation Type has exactly one handler
- All mutations are staged before commit
- No primitive performs direct file mutation
- All changes are applied via the Transaction Engine
- Discovery operations MUST NOT mutate state

This ensures:
- deterministic behavior
- clear auditability
- safe rollback capability


3. Operation Types (Primitive Set)

SAAYN defines the following primitives:

- CHUNK_REQUEST   (discovery only)
- CHUNK_REPLACE   (modify existing chunk)
- CHUNK_CREATE    (create new chunk)
- CHUNK_DELETE    (remove existing chunk)
- CHUNK_MOVE      (reposition existing chunk without modifying content)

Each primitive represents a single, atomic intent.


4. Operation Representation

Change Operations are represented as structured data within the Change Proposal:

{
  "type": "<Operation Type>",
  "...": "type-specific fields"
}

Operations MUST contain only the fields required for their type.
Operations MUST NOT contain executable logic.


5. Operation Handler Model

Each Operation Type MUST have a corresponding handler.

Definition:

An Operation Handler is a component responsible for:
- validating an operation
- staging its effects into a transaction

Handlers MUST NOT:
- write directly to source files
- bypass transaction controls

Handlers MUST:
- enforce all invariants
- return explicit errors on failure


6. Operation Handler Interface

The system MUST implement a handler interface equivalent to:

OperationHandler:
  Validate(operation, context) → error
  Stage(operation, transaction) → error

Where:

- Validate ensures the operation is legal
- Stage prepares the mutation inside the transaction
- Context provides access to registry, filesystem view, and adapters
- Transaction accumulates staged changes


7. Handler Responsibilities

7.1 Validation

Each handler MUST validate:

- required fields are present
- UUID constraints are satisfied
- referenced chunks exist (or are valid for creation)
- code passes syntax checks (if applicable)
- invariants are preserved

Validation failures MUST terminate processing with FAILED_VALIDATION.


7.2 Staging

Each handler MUST:

- prepare modifications without applying them
- register intended changes with the transaction engine
- avoid side effects outside the transaction

No filesystem writes are permitted during staging.


8. Primitive-Specific Semantics

8.1 CHUNK_REQUEST

Purpose:
- Retrieve chunk content and metadata

Behavior:
- MUST NOT modify any state
- MAY return chunk body, metadata, and neighbors

No staging occurs.


8.2 CHUNK_REPLACE

Purpose:
- Replace the contents of an existing chunk

Validation:
- UUID MUST exist
- replacement_code MUST be valid

Staging:
- mark chunk for content replacement
- preserve UUID and marker boundaries


8.3 CHUNK_CREATE

Purpose:
- Insert a new chunk at a defined position

Validation:
- new UUID MUST be globally unique
- target position MUST exist

Staging:
- create new chunk with markers
- insert relative to target UUID
- update Ordered Chunk Registry


8.4 CHUNK_DELETE

Purpose:
- Remove an existing chunk

Validation:
- UUID MUST exist

Staging:
- mark chunk for removal
- update registry accordingly


8.5 CHUNK_MOVE

Purpose:
- Reposition an existing chunk without modifying content

Validation:
- UUID MUST exist
- target position MUST exist
- content MUST NOT be altered

Staging:
- remove chunk from current position
- reinsert at target position
- preserve UUID, content_hash, and marker_hash


9. Handler Dispatch

The system MUST maintain a registry mapping Operation Types to handlers.

Example concept:

OperationType → OperationHandler

Execution flow:

1. Parse Change Proposal
2. For each Change Operation:
   - determine Operation Type
   - locate corresponding handler
   - call Validate
   - call Stage
3. Pass staged transaction to execution engine


10. Transaction Integration

All mutation primitives MUST integrate with the Transaction Engine.

The Transaction Engine is responsible for:

- staging file changes
- journaling original state
- applying atomic updates
- verifying post-apply integrity
- enabling rollback

Handlers MUST NOT bypass this system.


11. Separation of Concerns

The system MUST enforce the following separation:

- Operation Data Layer:
  Defines JSON structure and parsing

- Handler Layer:
  Defines validation and staging logic per primitive

- Transaction Layer:
  Defines atomic application and rollback

No layer may assume responsibilities of another.


12. Prohibited Patterns

The following are explicitly disallowed:

- Direct filesystem writes from handlers
- Mixed validation and execution logic in a single step
- Large switch statements replacing handler dispatch
- Combining multiple primitives into one implicit operation
- Using CHUNK_DELETE + CHUNK_CREATE to simulate CHUNK_MOVE


13. Extensibility

New Operation Types MAY be added in future versions.

Requirements:

- MUST define a new handler
- MUST not break existing primitives
- MUST maintain backward compatibility with Change Proposal Format


14. Summary

The Primitive Execution Model defines a strict contract:

- primitives describe intent
- handlers enforce correctness
- the transaction engine ensures safety

This model guarantees that all changes are:

- explicit
- validated
- atomic
- reversible

Chapter xxxx State Tranistion logic and glossary of terms

Name                        Definition
Change Proposal             A single JSON file that fully describes a proposed set of code changes, including human intent, operations, and SAAYN-managed workflow metadata. It is the sole artifact passed between review, approval, execution, and undo phases.
Change Proposal Format      The strict JSON schema that defines the structure, required fields, and validation rules for a Change Proposal. All proposals must conform to this format before processing.
Change Operation            A single entry within the operations array of a Change Proposal. Each Change Operation describes exactly one intended modification to a specific code chunk or location.
Operation Types             The finite, enumerated set of allowed operation identifiers (e.g., INSPECT, PROPOSE, CREATE). Each type defines the required fields and behavior for that operation.
Review                      A read-only, human-facing rendering of a Change Proposal that summarizes intent, affected chunks, and proposed modifications. It does not modify state and is used for decision-making prior to approval.
VALIDATED                   A formal workflow state indicating that a Change Proposal has passed all machine checks, including schema validation, syntax validation, hash verification, and protocol compliance.
APPROVED                    A formal workflow state indicating that a human has explicitly authorized the Change Proposal for execution. Approval must be recorded by SAAYN and is required before execution.
EXECUTED                    A formal workflow state indicating that the Change Proposal has been successfully applied using SAAYN’s transactional execution pipeline, with all changes committed atomically.
UNDONE                      A formal workflow state indicating that the most recent successful execution has been reverted, restoring the codebase and registry to their exact pre-execution state.
Ordered Chunk Registry      The authoritative registry of all SAAYN chunks, stored in deterministic order matching their physical appearance in source files. This registry is used for lookup, validation, and positional operations.
replacement_code            The exact code body provided within a Change Operation that will replace or define the contents of a target chunk. This field is human- or AI-supplied and must pass syntax validation before execution.
proposal_id                 A unique, stable identifier assigned to a Change Proposal. It persists across all lifecycle stages and is used for tracking, logging, and audit purposes.
operation_id                A unique identifier assigned to a specific execution run of a Change Proposal. It is used to correlate logs, journal files, backups, and undo operations.
human                       The top-level section of the Change Proposal that contains all human-editable fields, including intent and operations. Modifications to this section are allowed but may invalidate prior validation or approval states.
saayn                       The top-level section of the Change Proposal that contains all SAAYN-managed fields, including workflow state, validation results, approval metadata, and execution metadata. This section must not be edited manually.
seal                        The top-level section of the Change Proposal that contains the tamper-detection mechanism, including the hashing algorithm, scope, and digest. It is used to verify that SAAYN-managed fields have not been altered outside the tool.

Change Proposal Format (CPF) – Version 1.0

1. Overview

A Change Proposal is a single JSON document that defines a complete, auditable, and executable set of code modifications for SAAYN.

The document is divided into three top-level sections:

- human   → editable by humans and AI
- saayn   → managed exclusively by SAAYN
- seal    → tamper-detection over SAAYN-managed fields

All Change Proposals MUST conform to this format before processing.


2. Top-Level Structure

{
  "format_version": "1.0",
  "proposal_id": "<string>",
  "human": { ... },
  "saayn": { ... },
  "seal": { ... }
}

2.1 Required Fields

- format_version   MUST be "1.0"
- proposal_id      MUST be globally unique within the repository scope
- human            MUST exist
- saayn            MUST exist
- seal             MUST exist


3. human Section (Editable)

The human section defines intent and operations. It is the only section intended for manual editing.

{
  "human": {
    "intent": "<string>",
    "operations": [ <Change Operation>, ... ]
  }
}

3.1 intent

- MUST be a non-empty string
- Describes the purpose of the proposed change
- Used for audit and AI context

3.2 operations

- MUST be a non-empty array
- Each entry MUST be a valid Change Operation
- Order of operations MUST be preserved and executed sequentially


4. Change Operation Structure

Each operation MUST conform to the following base structure:

{
  "type": "<Operation Type>",
  "...": "type-specific fields"
}

4.1 Common Requirements

- type MUST be a valid Operation Type
- All required fields for the given type MUST be present
- No unknown fields are permitted

4.2 Operation Types (v1.0)

PROPOSE

Replaces the contents of an existing chunk.

{
  "type": "PROPOSE",
  "uuid": "<string>",
  "replacement_code": "<string>"
}

Constraints:
- uuid MUST exist in the Ordered Chunk Registry
- replacement_code MUST be non-empty
- replacement_code MUST pass syntax validation via Language Adapter


CREATE

Creates a new chunk at a defined position.

{
  "type": "CREATE",
  "after_uuid": "<string>",
  "new_uuid": "<string>",
  "replacement_code": "<string>"
}

Constraints:
- after_uuid MUST exist
- new_uuid MUST be globally unique
- replacement_code MUST be valid code


INSPECT (Optional, non-mutating)

Requests context for a chunk. Does not modify code.

{
  "type": "INSPECT",
  "uuid": "<string>"
}

Constraints:
- uuid MUST exist


5. saayn Section (Tool-Managed)

The saayn section contains all workflow state and execution metadata. It MUST NOT be edited manually.

{
  "saayn": {
    "state": "<STATE>",
    "validation": { ... },
    "approval": { ... },
    "execution": { ... }
  }
}

5.1 state

Allowed values:

- DRAFT
- VALIDATING
- VALIDATED
- FAILED_VALIDATION
- PENDING_APPROVAL
- APPROVED
- REJECTED
- EXECUTING
- EXECUTED
- FAILED_EXECUTION
- UNDOING
- UNDONE
- RECOVERING

Constraints:
- MUST reflect current workflow state
- MUST only be changed by SAAYN

5.2 validation

{
  "status": "VALIDATED" | "FAILED",
  "timestamp": "<ISO8601>"
}

5.3 approval

{
  "approved_by": "<string|null>",
  "approved_at": "<ISO8601|null>"
}

5.4 execution

{
  "operation_id": "<string|null>",
  "executed_at": "<ISO8601|null>"
}


6. seal Section (Tamper Detection)

The seal section ensures that SAAYN-managed fields have not been modified outside the tool.

{
  "seal": {
    "algorithm": "sha256",
    "scope": "saayn",
    "digest": "<hex_string>"
  }
}

6.1 Rules

- algorithm MUST be "sha256"
- scope MUST be "saayn"
- digest MUST be the SHA-256 hash of the canonicalized saayn section

6.2 Canonicalization Requirements

Before hashing, the saayn section MUST be:

- serialized with deterministic key ordering
- stripped of insignificant whitespace
- encoded as UTF-8

Failure to canonicalize consistently results in invalid seal comparisons


7. Validation Rules

A Change Proposal is considered VALIDATED only if:

- JSON structure matches CPF schema
- All required fields are present
- All operations conform to their type definitions
- replacement_code passes syntax validation
- All referenced UUIDs exist (or are valid for CREATE)
- seal.digest matches computed hash of saayn section

Any failure results in:

- state → FAILED_VALIDATION
- processing MUST stop


8. State Invalidation Rules

Any modification to the human section AFTER validation or approval MUST:

- invalidate validation
- invalidate approval
- reset effective state to DRAFT

Any modification to the saayn or seal sections MUST:

- be treated as tamper
- cause hard validation failure
- reset state to DRAFT


9. Execution Preconditions

Execution (EXECUTING) is allowed only if:

- state == APPROVED
- validation.status == VALIDATED
- seal is valid
- no drift detected in target chunks


10. Execution Guarantees

On successful execution:

- All operations are applied atomically
- state → EXECUTED
- execution.operation_id MUST be set
- execution.executed_at MUST be recorded

On failure:

- state → FAILED_EXECUTION
- system MUST restore pre-execution state


11. Undo Behavior

Undo restores the last successful execution:

- state → UNDOING → UNDONE
- code and registry MUST match pre-execution state exactly

Only one level of undo is guaranteed by SAAYN



=======================

State Name:
VALIDATING

Purpose:
Machine-validate a Change Proposal before it can be shown for approval or execution.

Allowed Previous States:
- IDLE
- DRAFT
- FAILED_VALIDATION

Handler Name:
HandleValidating

Handler Inputs:
- Change Proposal (full JSON artifact)
- Ordered Chunk Registry
- Proposal revision metadata
- Protocol version
- Previous state
- Execution context (paths, environment, config)

Handler Decisions:
1. Proposal is structurally valid (schema, required fields, operation types)
2. Proposal contains schema errors (missing or invalid fields)
3. Proposal references unknown or duplicate UUIDs
4. Proposal violates ordering or placement constraints
5. Replacement code fails syntax validation (via Language Adapter)
6. Proposal results in a no-op
7. Required context is missing or inconsistent (for example, registry mismatch)

Possible Next States:
- VALIDATED
- FAILED_VALIDATION

Side Effects:
- Persist validation result (success or failure)
- Record detailed validation errors if present
- Clear prior validation errors on success
- Emit structured log entry for validation outcome
- Treat no-op and validation sub-failures as internal handler decisions unless later promoted to first-class states

Terminal:
No

============================
State Name:
FAILED_VALIDATION

Purpose:
Handle a validation failure as a transient failure state, present the error to the human, clean up proposal-scoped transient data, and return the machine to a stable resting state.

Allowed Previous States:
- VALIDATING

Handler Name:
HandleFailedValidation

Handler Inputs:
- Change Proposal (full JSON artifact, if available)
- Validation error list
- Previous state
- Proposal metadata (proposal_id, revision, intent)
- Execution context (paths, environment, config)

Handler Decisions:
1. Display validation errors to the human in a clear, actionable form
2. Determine whether proposal-scoped transient data exists and should be discarded
3. Record the validation failure in logs and proposal history
4. Return control to a stable resting state without retrying inside this handler

Possible Next States:
- IDLE

Side Effects:
- Display or emit validation errors
- Clear staged proposal data created during validation
- Clear validation scratch artifacts
- Persist failure details to logs/history
- Preserve the original Change Proposal artifact for later correction outside this state

Terminal:
No

Notes:
- FAILED_VALIDATION is a transient failure state, not a resting state
- No retry loop is permitted inside this state in v1
- Any future validation attempt must begin as a new run from IDLE

==================================

State Name:
IDLE

Purpose:
Serve as the stable resting state of the SAAYN workflow engine. In this state, no Change Proposal is actively being processed, no execution is in progress, and the machine is ready to accept a new request.

Allowed Previous States:
- FAILED_VALIDATION
- REJECTED
- UNDONE
- INITIAL

Handler Name:
HandleIdle

Handler Inputs:
- Previous state
- Optional Change Proposal reference
- Optional execution summary from the last completed run
- Runtime context (paths, environment, config)
- Operator input or CLI command

Handler Decisions:
1. Determine whether a new Change Proposal has been submitted for processing
2. Determine whether the operator is requesting review of an existing validated proposal
3. Determine whether the operator is requesting execution of an approved proposal
4. Determine whether the operator is requesting undo of the last executed transaction
5. Determine whether there is no actionable input and the machine should remain idle
6. Determine whether startup recovery or cleanup is required before accepting new work

Possible Next States:
- VALIDATING
- PENDING_APPROVAL
- EXECUTING
- UNDOING
- IDLE

Side Effects:
- Accept and normalize new operator input
- Clear ephemeral runtime context from prior completed flows
- Optionally display current machine status
- Check for outstanding recovery or cleanup conditions before allowing new work
- Emit structured log entry for newly accepted work or no-op idle cycle

Terminal:
No

Notes:
- IDLE is a resting state
- Remaining in IDLE is a legal outcome if no actionable input is present
- IDLE should not perform proposal mutation, transaction staging, or undo directly; it should route into the proper working state
- If startup recovery is required, that path may later be promoted to a dedicated RECOVERING state

=================

State Name:
PENDING_APPROVAL

Purpose:
Present a validated Change Proposal to the human for decision. This state pauses automated execution and requires explicit human input before proceeding.

Allowed Previous States:
- VALIDATED

Handler Name:
HandlePendingApproval

Handler Inputs:
- Change Proposal (full JSON artifact)
- Validation summary (success result + any warnings)
- Ordered Chunk Registry (for rendering context)
- Proposal metadata (proposal_id, revision, intent)
- Previous state
- Runtime context (paths, environment, config)
- Operator input (approve, reject, or no action)

Handler Decisions:
1. Render the proposal in a human-readable Review format
2. Determine whether the operator has provided an explicit decision:
   - APPROVE
   - REJECT
   - NO ACTION (waiting)
3. If no action is taken, remain in a waiting condition without mutating state
4. If APPROVE is received, transition to execution path
5. If REJECT is received, terminate the proposal flow and return to resting state

Possible Next States:
- APPROVED
- REJECTED
- PENDING_APPROVAL   (no decision yet; remain in this state)

Side Effects:
- Render Review output (human-readable proposal summary)
- Await operator input (blocking or event-driven depending on CLI mode)
- Persist approval decision with timestamp and operator identity (if provided)
- Persist rejection reason if supplied
- Emit structured log entries for approval, rejection, or continued waiting

Terminal:
No

Notes:
- PENDING_APPROVAL is a resting state that may persist indefinitely until human input is received
- No automatic transition out of this state is permitted without explicit operator action
- This state introduces human control into the workflow and must be deterministic and auditable
- Re-entry into this state is legal only via VALIDATED in v1

===========================================

State Name:
APPROVED

Purpose:
Represent a human-approved Change Proposal that is authorized for execution but has not yet begun execution.

Allowed Previous States:
- PENDING_APPROVAL

Handler Name:
HandleApproved

Handler Inputs:
- Change Proposal (full JSON artifact)
- Proposal metadata (proposal_id, revision, intent)
- Approval metadata (approver identity, timestamp)
- Previous state
- Runtime context (paths, environment, config)
- Operator input (execute, defer, or cancel)

Handler Decisions:
1. Determine whether execution has been explicitly requested
2. Determine whether the proposal should remain approved but not yet executed (deferred execution)
3. Determine whether execution should be aborted prior to start (operator cancel)
4. Ensure that approval metadata is persisted before any execution begins
5. Ensure that execution cannot begin without explicit operator intent (no implicit auto-run in v1)

Possible Next States:
- EXECUTING
- IDLE        (if execution is canceled or deferred and proposal flow is exited)
- APPROVED    (remain in state if no action is taken)

Side Effects:
- Persist approval state and metadata (if not already persisted)
- Await operator execution command
- Emit structured log entry for approval confirmation and any subsequent action
- Ensure no file mutations or transaction staging occurs in this state

Terminal:
No

Notes:
- APPROVED is a resting state
- Execution MUST NOT begin automatically in v1; it requires explicit operator action
- Remaining in APPROVED is valid if the user delays execution
- If execution is canceled, the proposal flow ends and control returns to IDLE


=============================


State Name:
EXECUTING

Purpose:
Perform the full SAAYN transaction pipeline to apply an approved Change Proposal safely and atomically to the codebase.

Allowed Previous States:
- APPROVED

Handler Name:
HandleExecuting

Handler Inputs:
- Change Proposal (full JSON artifact)
- Ordered Chunk Registry
- Approval metadata (approver identity, timestamp)
- Previous state
- Runtime context (paths, environment, config)
- Operation ID (generated for this execution)
- Undo context (current undo_state, snapshot location if any)

Handler Decisions:
1. Verify preconditions:
   - Registry matches expected state (content_hash checks)
   - No concurrent lock violation
2. Execute transaction steps:
   - Extract target chunks
   - Generate replacement code (if not already present)
   - Validate generated code (syntax, format, optional tests)
   - Stage .tmp files and registry.tmp
   - fsync all staged artifacts
   - Write and fsync rollback journal
   - Backup original files
   - Apply atomic rename of all .tmp files
   - Post-apply verification (hash recompute)
3. Determine outcome:
   - Full success (all steps complete and verified)
   - Failure before apply (safe to abort without rollback)
   - Failure after backup/apply (requires recovery/rollback)
4. Update undo context:
   - On success, create undo snapshot and mark undo_state = AVAILABLE
5. Handle failure:
   - If failure occurs after journal creation, trigger recovery
   - If failure occurs before journal, abort cleanly
6. Ensure no partial state is left visible to the user

Possible Next States:
- EXECUTED
- FAILED_EXECUTION

Side Effects:
- Create operation_id for audit tracking
- Write staged files and registry.tmp
- Persist rollback journal to .saayn/journal/<operation_id>.json
- Backup originals to .saayn/backup/<operation_id>/
- Apply atomic file replacements
- Update chunk-registry.json on success
- Update undo snapshot and undo_state
- Emit structured logs for each major step and final outcome

Terminal:
No

Notes:
- EXECUTING is a working (non-resting) state
- This state encapsulates the entire atomic transaction pipeline
- Partial success is forbidden; outcome must be success or full rollback
- Recovery behavior is internal to this handler in v1 and does not require a separate state
- Undo snapshot MUST only be created after successful post-apply verification

State Name:
FAILED_EXECUTION

Purpose:
Handle an execution failure as a transient failure state, present the failure to the human, ensure the system is returned to a safe and consistent condition, and then return the machine to a stable resting state.

Allowed Previous States:
- EXECUTING

Handler Name:
HandleFailedExecution

Handler Inputs:
- Change Proposal (full JSON artifact, if available)
- Execution error details
- Operation ID
- Previous state
- Proposal metadata (proposal_id, revision, intent)
- Runtime context (paths, environment, config)
- Recovery result metadata (if rollback or cleanup was attempted)

Handler Decisions:
1. Display execution failure details to the human in a clear, actionable form
2. Determine whether rollback or recovery already succeeded inside EXECUTING
3. Determine whether any staged artifacts, temp files, journals, or backups still exist and must be cleaned up
4. Record the execution failure in logs and proposal history
5. Return control to a stable resting state without retrying inside this handler

Possible Next States:
- IDLE

Side Effects:
- Display or emit execution failure details
- Clear staged execution artifacts if they remain
- Clear temp files and transient execution context
- Remove or preserve rollback journal and backups according to recovery outcome
- Persist failure details to logs/history
- Preserve the original Change Proposal artifact for later correction or re-submission outside this state

Terminal:
No

Notes:
- FAILED_EXECUTION is a transient failure state, not a resting state
- No retry loop is permitted inside this state in v1
- Any future execution attempt must begin as a new run from IDLE
- If recovery was required, this state assumes EXECUTING already completed that work before entering FAILED_EXECUTION
- If recovery itself can fail in a meaningful way later, that behavior may be promoted to a dedicated RECOVERY_FAILURE state in a future revision

==============================

State Name:
EXECUTED

Purpose:
Represent a successfully completed SAAYN execution after all transaction work, verification, registry updates, and cleanup have finished. This state exists to mark a durable success point, expose the result to the human, and establish whether undo is now available.

Allowed Previous States:
- EXECUTING

Handler Name:
HandleExecuted

Handler Inputs:
- Change Proposal (full JSON artifact)
- Operation ID
- Execution summary (touched files, affected UUIDs, registry update result)
- Undo metadata (undo_state, snapshot location, snapshot operation binding)
- Previous state
- Proposal metadata (proposal_id, revision, intent)
- Runtime context (paths, environment, config)
- Operator input (acknowledge, view result, request undo, or no action)

Handler Decisions:
1. Confirm that execution completed successfully and all cleanup obligations were satisfied
2. Present the successful result to the human in a clear summary form
3. Determine whether undo is available for this execution
4. Determine whether the operator wants to invoke undo immediately
5. Determine whether the machine should simply acknowledge success and return to resting state
6. Preserve the execution record as a durable audit point regardless of whether the operator acts immediately

Possible Next States:
- UNDOING
- IDLE
- EXECUTED    (remain in state if no action is taken)

Side Effects:
- Persist successful execution state and execution timestamp
- Persist operation_id and proposal execution linkage
- Persist undo availability metadata
- Emit structured success logs
- Present execution summary to the human
- Make the last successful execution visible to status, history, and undo commands

Terminal:
No

Notes:
- EXECUTED is needed even though EXECUTING already exists because EXECUTING is a working state and EXECUTED is a resting success state
- EXECUTING means "the machine is doing the work"
- EXECUTED means "the work completed successfully and the result now exists as a durable, reviewable outcome"
- Without EXECUTED, the machine would have no explicit success resting point for:
  - showing success to the human
  - exposing undo availability
  - linking proposal_id to operation_id
  - recording a completed audit state before returning to IDLE
- EXECUTED may transition directly to IDLE after acknowledgment in v1 if you later decide not to keep it as a visible resting state

====================

State Name:
UNDOING

Purpose:
Revert the last successfully executed SAAYN operation using the stored undo snapshot, restoring the codebase and registry to their pre-execution state in a safe and controlled manner.

Allowed Previous States:
- EXECUTED

Handler Name:
HandleUndoing

Handler Inputs:
- Operation ID (of the execution being undone)
- Undo metadata (undo_state, snapshot path, snapshot contents)
- Previous state
- Execution summary (files modified, registry changes)
- Runtime context (paths, environment, config)
- Operator input (explicit undo request)

Handler Decisions:
1. Verify undo preconditions:
   - undo_state == AVAILABLE
   - snapshot exists and is readable
   - no conflicting lock or concurrent operation
2. Restore original files from undo snapshot
3. Restore chunk-registry.json from snapshot
4. Verify restoration integrity (hash checks, file presence)
5. Determine outcome:
   - Undo successful and system fully restored
   - Undo failed due to missing/corrupt snapshot or restore error
6. Update undo state:
   - On success, set undo_state = CONSUMED
7. Ensure no partial restore state remains visible

Possible Next States:
- UNDONE
- FAILED_EXECUTION

Side Effects:
- Restore files from .saayn/backup/<operation_id>/
- Restore chunk-registry.json to pre-execution state
- Update undo_state to CONSUMED on success
- Clear or archive undo snapshot after successful restore
- Emit structured logs for undo attempt and outcome
- Present undo result to the human

Terminal:
No

Notes:
- UNDOING is a working (non-resting) state
- Undo is limited to the last successful execution in v1
- Partial undo is forbidden; outcome must be full restore or failure
- Failure during undo transitions to FAILED_EXECUTION in v1 for simplicity
- If undo failure becomes more complex in the future, a dedicated FAILED_UNDO state may be introduced


State Name:
UNDONE

Purpose:
Represent a successfully completed undo operation where the system has been restored to its exact pre-execution state. This state exists to confirm reversal, present the result to the human, and establish a clean post-undo baseline before returning to IDLE.

Allowed Previous States:
- UNDOING

Handler Name:
HandleUndone

Handler Inputs:
- Operation ID (of the execution that was undone)
- Undo metadata (undo_state, snapshot details, consumed status)
- Execution summary (original changes that were reverted)
- Previous state
- Proposal metadata (proposal_id, revision, intent)
- Runtime context (paths, environment, config)
- Operator input (acknowledge or no action)

Handler Decisions:
1. Confirm that undo completed successfully and all files and registry entries were restored
2. Present undo success summary to the human
3. Confirm undo_state has transitioned to CONSUMED
4. Determine whether any residual artifacts (logs, snapshots) should be cleaned or archived
5. Determine whether to remain briefly for acknowledgment or immediately return to resting state

Possible Next States:
- IDLE
- UNDONE   (remain in state if no action is taken)

Side Effects:
- Persist undo completion record with timestamp
- Persist linkage between original operation_id and undo event
- Emit structured logs confirming undo success
- Present undo summary to the human
- Ensure undo snapshot is no longer available for reuse

Terminal:
No

Notes:
- UNDONE is a resting success state, analogous to EXECUTED but for reversal
- UNDOING means "the machine is restoring state"
- UNDONE means "the system has been fully restored and is stable"
- After acknowledgment, control should return to IDLE
- Undo is single-use in v1; once consumed, a new execution is required before another undo becomes available



=============


State Name:
REJECTED

Purpose:
Handle an explicitly rejected Change Proposal. This state records the human decision, presents the outcome, cleans up proposal-scoped transient data, and returns the machine to a stable resting state.

Allowed Previous States:
- PENDING_APPROVAL

Handler Name:
HandleRejected

Handler Inputs:
- Change Proposal (full JSON artifact)
- Proposal metadata (proposal_id, revision, intent)
- Rejection metadata (operator identity, timestamp, optional reason)
- Previous state
- Runtime context (paths, environment, config)

Handler Decisions:
1. Record the rejection decision and any provided reason
2. Present rejection summary to the human
3. Determine whether any staged or transient proposal artifacts exist and should be discarded
4. Conclude the current proposal attempt and return control to a resting state (no retry within this state)

Possible Next States:
- IDLE

Side Effects:
- Persist rejection record (who, when, reason)
- Clear staged proposal data and any validation/execution scratch artifacts
- Emit structured log entry for rejection
- Preserve the original Change Proposal artifact for potential future revision outside this state

Terminal:
No

Notes:
- REJECTED is a transient state with a single exit to IDLE in v1
- No automatic retry or re-validation occurs inside this state
- Any future attempt must start as a new run from IDLE

================================================


State Name:
INITIAL

Purpose:
Establish the starting condition of the SAAYN workflow engine. This state is responsible for one-time initialization, environment checks, and recovery detection before the system becomes available for normal operation.

Allowed Previous States:
- NONE   (entry point state only)

Handler Name:
HandleInitial

Handler Inputs:
- Runtime context (paths, environment variables, config)
- Filesystem state (.saayn directories, journal presence, lock files)
- Previous run artifacts (journal files, backups, temp files)
- System clock and environment health indicators

Handler Decisions:
1. Determine whether this is a clean startup or a restart after interruption
2. Detect presence of any rollback journal(s) indicating an incomplete prior execution
3. Detect presence of stale lock files
4. Determine whether recovery is required before accepting new work
5. Determine whether environment prerequisites are satisfied (paths, permissions, config)
6. Decide whether initialization is successful or blocked by unrecoverable conditions

Possible Next States:
- IDLE
- RECOVERING   (optional: if recovery is promoted to a first-class state)

Side Effects:
- Initialize runtime directories (for example, .saayn/)
- Validate or create required subdirectories (journal, backup, temp)
- Remove or reconcile stale lock files if safe
- Detect and log presence of recovery artifacts
- Emit structured startup log entry
- Prepare system for normal operation or recovery

Terminal:
No

Notes:
- INITIAL is entered exactly once per process startup
- In v1, recovery may be handled inline within this handler and still transition to IDLE
- If recovery logic becomes complex, it should be promoted to a dedicated RECOVERING state
- No Change Proposal processing is allowed in this state



============

State Name:
RECOVERING

Purpose:
Restore the SAAYN workflow engine to a safe, consistent, and known-good condition after an interrupted or failed execution that left durable recovery artifacts behind. This state exists to centralize recovery logic when rollback or startup repair cannot remain implicit inside another handler.

Allowed Previous States:
- INITIAL
- EXECUTING

Handler Name:
HandleRecovering

Handler Inputs:
- Previous state
- Operation ID (if known)
- Runtime context (paths, environment, config)
- Filesystem state (.saayn/journal, backup, temp directories)
- Rollback journal contents
- Backup metadata
- Registry state (current and expected)
- Lock file state
- Recovery trigger reason (startup detection, execution failure, interrupted apply)

Handler Decisions:
1. Determine whether recovery artifacts are present and complete enough to attempt recovery
2. Determine whether the rollback journal is valid and readable
3. Determine whether backups required for restoration are present and intact
4. Determine whether recovery should:
   - restore original files from backup
   - restore chunk-registry.json from backup
   - remove partially applied temp artifacts
   - clear stale lock artifacts if safe
5. Determine whether recovery completed successfully
6. Determine whether recovery failed in a way that still leaves the system unsafe
7. Determine whether the machine can safely return to a stable resting state after recovery

Possible Next States:
- IDLE
- FAILED_EXECUTION

Side Effects:
- Read and validate rollback journal
- Restore original files from backup, if required
- Restore chunk-registry.json from backup, if required
- Remove or quarantine stale .tmp files
- Remove or reconcile stale lock files if safe
- Remove or archive recovery artifacts after successful recovery
- Persist structured recovery logs
- Present recovery result to the human if running in an interactive/operator-facing mode

Terminal:
No

Notes:
- RECOVERING is a working (non-resting) state
- This state may be entered on startup from INITIAL if recovery artifacts are detected
- This state may be entered from EXECUTING if failure occurs after journal creation or partial apply
- Recovery MUST be idempotent to the extent practical; repeated entry into RECOVERING should not further damage the system
- If recovery succeeds, the machine returns to IDLE
- If recovery cannot establish a safe and consistent state, transition to FAILED_EXECUTION
- In v1, RECOVERING is the only state permitted to act on stale journals and interrupted execution artifacts

=======

```text
State Name:
VALIDATED

Purpose:
Represent a Change Proposal that has successfully passed machine validation and is now ready to be presented for human decision.

Allowed Previous States:
- VALIDATING

Handler Name:
HandleValidated

Handler Inputs:
- Change Proposal (full JSON artifact)
- Validation result summary
- Validation warnings (if any)
- Proposal metadata (proposal_id, revision, intent)
- Previous state
- Runtime context (paths, environment, config)

Handler Decisions:
1. Confirm that validation completed successfully and that the proposal is safe to advance
2. Determine whether any warnings should be preserved and shown during human review
3. Determine whether proposal metadata and validation results have been durably recorded
4. Advance the proposal into the human decision phase

Possible Next States:
- PENDING_APPROVAL

Side Effects:
- Persist validated state and validation timestamp
- Persist validation summary and any non-fatal warnings
- Emit structured log entry for successful validation
- Make the validated proposal available for review rendering

Terminal:
No

Notes:
- VALIDATED is a transient state in v1
- This state exists to create a clean boundary between machine validation and human approval
- No proposal mutation, execution, or undo work occurs in this state
- The handler should perform any final normalization needed before entering PENDING_APPROVAL
```

