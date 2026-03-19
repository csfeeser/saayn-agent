Here is the complete, integrated, and semantically audited **SAAYN Agent SpecBook (v1.9)**. This version consolidates all prior modules, enforces strict primitive naming, and locks down the transactional and protocol-level boundaries.

---

# SAAYN Agent SpecBook (v1.9)
**Project Name:** `saayn-agent` | **Binary Name:** `saayn`  
**Motto:** The UPC Barcode System for AI-Native Codebases.

---

## Chapter 1: The Six Laws of SAAYN
1.  **Zero Collateral Damage:** Never modify a single byte outside of explicitly targeted `SAAYN:CHUNK` boundaries.
2.  **The Registry is Law:** If a chunk isn't in `chunk-registry.json`, it is invisible to the agent.
3.  **Director Mode:** Humans provide intent; the CLI performs extraction, prompting, and replacement.
4.  **Sub-Second Local Execution:** Local operations (parsing, hashing, staging) must be near-instantaneous.
5.  **Protocol-Only Communication:** The agent strictly enforces and validates "raw code only" output. Zero markdown.
6.  **Transactional Atomicity:** All-or-nothing global state. Partial applies are strictly forbidden.

---

## Chapter 2: Marker Grammar & Sovereign UUIDs
**The Marker Grammar:**
Markers must occupy their own line. Inline code following a marker is a Protocol Violation.
* `<comment_prefix> SAAYN:CHUNK_START:<uuid>`
* `<comment_prefix> SAAYN:CHUNK_END:<uuid>`

**UUID Semantics:**
* **Format:** `<slug>-v<version>-<8_hex_chars>` (e.g., `db-init-v1-a3c7d2f8`).
* **Sovereignty:** A UUID represents a unique **logical unit of responsibility**, not a file position. 
* **Immutability:** Once assigned to a logic block, the UUID never changes.
* **Global Uniqueness:** A UUID must appear exactly twice (START/END) in the entire repository.
* **Integrity:** Files must be UTF-8. Binary files are prohibited. No nested chunks.

---

## Chapter 3: The Registry & Data Model
**Location:** `chunk-registry.json` (Ordered by physical appearance in source).

**Schema:**
* `uuid`: Primary Key.
* `file_path`: Relative path.
* `language_hint`: Triggers the Language Adapter.
* `content_hash`: SHA-256 of the code body (excluding markers).
* `marker_hash`: SHA-256 of the exact START/END lines.
* `version`: Auto-incrementing integer.
* `line_span`: `{ "start": int, "end": int, "confidence": "low" }`.

---

## Chapter 4: Primitive Execution Model
SAAYN executes changes using a bounded set of primitives to ensure deterministic behavior.

1.  **CHUNK_REQUEST:** Discovery only. Retrieve content and metadata. No state mutation.
2.  **CHUNK_REPLACE:** Modify existing chunk content. UUID remains invariant.
3.  **CHUNK_CREATE:** Insert a new chunk relative to a target (`after_uuid`). Requires new unique UUID.
4.  **CHUNK_DELETE:** Remove a chunk and its markers from the file and registry.
5.  **CHUNK_MOVE:** Reposition an existing chunk while preserving UUID and `content_hash`. It is NOT equivalent to DELETE + CREATE at the protocol level.

---

## Chapter 5: Change Proposal Format (CPF) v1.0
The Change Proposal is the sole artifact passed between review, approval, and execution.

**Top-Level Structure:**
* **`human`**: Editable section containing `intent` and `operations[]`.
* **`saayn`**: Tool-managed section containing `state`, `validation`, `approval`, and `execution` metadata.
* **`seal`**: Tamper-detection. A SHA-256 digest of the canonicalized `saayn` section.

**State Invalidation:** Any modification to the `human` section after validation resets the `state` to `DRAFT`. Any manual modification to `saayn` or `seal` results in a **Hard Tamper Failure**.


To finalize **Chapter 5**, here is the formal **JSON Schema** for the `Change Proposal Format (CPF) v1.0`. This schema is designed for use with standard JSON-Schema validators (Draft 7 or later) to enforce the structural integrity, primitive types, and tool-managed constraints defined in the SpecBook.

---

## Chapter 5.1: The Change Proposal JSON Schema (v1.0)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://saayn.org/schema/change-proposal-v1.json",
  "title": "SAAYN Change Proposal Format",
  "description": "The strict structural definition for SAAYN Change Proposals.",
  "type": "object",
  "required": ["format_version", "proposal_id", "human", "saayn", "seal"],
  "additionalProperties": false,
  "properties": {
    "format_version": {
      "type": "string",
      "enum": ["1.0"],
      "description": "Must be exactly 1.0 for this specification."
    },
    "proposal_id": {
      "type": "string",
      "format": "uuid",
      "description": "A unique, stable identifier for the lifecycle of this change."
    },
    "human": {
      "type": "object",
      "required": ["intent", "operations"],
      "additionalProperties": false,
      "properties": {
        "intent": {
          "type": "string",
          "minLength": 1,
          "description": "A natural language description of the change's purpose."
        },
        "operations": {
          "type": "array",
          "minItems": 1,
          "maxItems": 50,
          "items": {
            "oneOf": [
              { "$ref": "#/definitions/CHUNK_REPLACE" },
              { "$ref": "#/definitions/CHUNK_CREATE" },
              { "$ref": "#/definitions/CHUNK_DELETE" },
              { "$ref": "#/definitions/CHUNK_MOVE" },
              { "$ref": "#/definitions/CHUNK_REQUEST" }
            ]
          }
        }
      }
    },
    "saayn": {
      "type": "object",
      "required": ["state", "validation", "approval", "execution"],
      "additionalProperties": false,
      "properties": {
        "state": {
          "type": "string",
          "enum": [
            "DRAFT", "VALIDATING", "VALIDATED", "FAILED_VALIDATION", 
            "PENDING_APPROVAL", "APPROVED", "REJECTED", 
            "EXECUTING", "EXECUTED", "FAILED_EXECUTION", 
            "UNDOING", "UNDONE", "RECOVERING"
          ]
        },
        "validation": {
          "type": "object",
          "required": ["status", "timestamp"],
          "properties": {
            "status": { "type": "string", "enum": ["VALIDATED", "FAILED", "null"] },
            "timestamp": { "type": ["string", "null"], "format": "date-time" },
            "errors": { "type": "array", "items": { "type": "string" } }
          }
        },
        "approval": {
          "type": "object",
          "required": ["approved_by", "approved_at"],
          "properties": {
            "approved_by": { "type": ["string", "null"] },
            "approved_at": { "type": ["string", "null"], "format": "date-time" }
          }
        },
        "execution": {
          "type": "object",
          "required": ["operation_id", "executed_at"],
          "properties": {
            "operation_id": { "type": ["string", "null"] },
            "executed_at": { "type": ["string", "null"], "format": "date-time" }
          }
        }
      }
    },
    "seal": {
      "type": "object",
      "required": ["algorithm", "scope", "digest"],
      "additionalProperties": false,
      "properties": {
        "algorithm": { "type": "string", "enum": ["sha256"] },
        "scope": { "type": "string", "enum": ["saayn"] },
        "digest": { 
          "type": "string", 
          "pattern": "^[a-f0-9]{64}$",
          "description": "Hex-encoded SHA-256 hash of canonicalized saayn section."
        }
      }
    }
  },
  "definitions": {
    "CHUNK_REPLACE": {
      "type": "object",
      "required": ["type", "uuid", "replacement_code"],
      "properties": {
        "type": { "const": "CHUNK_REPLACE" },
        "uuid": { "type": "string" },
        "replacement_code": { "type": "string" }
      }
    },
    "CHUNK_CREATE": {
      "type": "object",
      "required": ["type", "after_uuid", "new_uuid", "replacement_code"],
      "properties": {
        "type": { "const": "CHUNK_CREATE" },
        "after_uuid": { "type": "string" },
        "new_uuid": { "type": "string" },
        "replacement_code": { "type": "string" }
      }
    },
    "CHUNK_DELETE": {
      "type": "object",
      "required": ["type", "uuid"],
      "properties": {
        "type": { "const": "CHUNK_DELETE" },
        "uuid": { "type": "string" }
      }
    },
    "CHUNK_MOVE": {
      "type": "object",
      "required": ["type", "uuid", "after_uuid"],
      "properties": {
        "type": { "const": "CHUNK_MOVE" },
        "uuid": { "type": "string" },
        "after_uuid": { "type": "string" }
      }
    },
    "CHUNK_REQUEST": {
      "type": "object",
      "required": ["type", "uuid"],
      "properties": {
        "type": { "const": "CHUNK_REQUEST" },
        "uuid": { "type": "string" }
      }
    }
  }
}
```

---

### Key Enforcement Points:
* **Operational Integrity:** The `operations` array uses `oneOf` to ensure that every operation matches exactly one primitive structure (e.g., `CHUNK_CREATE` must have `after_uuid`, but `CHUNK_REPLACE` must not).
* **Tamper Detection:** The `seal.digest` is regex-validated to be a valid 64-character hex string.
* **Ownership Guardrails:** While the schema doesn't "lock" the file, it provides the machine-readable definitions needed for the Agent to identify if a human has illegally modified a `saayn` enum or field.
* **Lifecycle Stability:** Use of standard ISO 8601 date-time formats for `approved_at` and `executed_at` ensures cross-platform consistency.

---

## Chapter 6: The State Machine & Handler Contracts


### 6.1 State: INITIAL
* **Purpose:** Establish starting conditions and recovery detection.
* **Handler Decisions:** Detect clean startup vs. interrupted restart; check for stale locks/journals.
* **Next States:** `IDLE`, `RECOVERING`.

### 6.2 State: IDLE (Resting State)
* **Purpose:** Stable resting state ready for new input.
* **Allowed Previous:** `FAILED_VALIDATION`, `REJECTED`, `UNDONE`, `INITIAL`.
* **Next States:** `VALIDATING`, `PENDING_APPROVAL`, `EXECUTING`, `UNDOING`, `IDLE`.

### 6.3 State: VALIDATING
* **Purpose:** Machine-validate a Change Proposal before review or execution.
* **Handler Decisions:** Structural validity; UUID existence; Syntax validation via Language Adapter; No-op detection.
* **Side Effects:** Persist validation result/errors; emit structured log.
* **Next States:** `VALIDATED`, `FAILED_VALIDATION`.

### 6.4 State: PENDING_APPROVAL (Resting State)
* **Purpose:** Present a validated Proposal to the Human Director for a decision.
* **Next States:** `APPROVED`, `REJECTED`, `PENDING_APPROVAL` (wait).

### 6.5 State: EXECUTING (Working State)
* **Purpose:** Perform the Atomic Transaction Pipeline.
* **Handler Decisions:** Pre-flight Drift Check; Stage -> Journal -> Backup -> Rename -> Verify.
* **Next States:** `EXECUTED`, `FAILED_EXECUTION`.

---

## Chapter 7: Transactional Integrity & Registry Mutation
To ensure global atomicity, the registry is updated **only after** successful file application.

### 7.1 The Transaction Pipeline
1.  **Drift Detection:** If `disk_hash != registry_hash`, abort with `DRIFT_ERROR`.
2.  **Stage:** Write `.tmp` files.
3.  **Flush:** Call `fsync()` on all `.tmp` files.
4.  **Journal:** Write and `fsync()` the `saayn_journal.json`.
5.  **Backup:** Move original files to `.saayn/backup/`.
6.  **Apply:** Atomic `rename()` of `.tmp` to real files.
7.  **Registry Mutation:** Update entry/hashes in `chunk-registry.json`.
8.  **Final Sync:** `fsync()` registry and cleanup journal.

---

## Chapter 8: The Language Adapter Contract
Adapters provide:
* **`CommentPrefix()`**: Language-specific delimiters.
* **`SyntaxCheck(code)`**: Mandatory Level 1 parse to prevent build-breaking hallucinations.
* **`Format(code)`**: (Optional) Invoke canonical formatters.

---

## Chapter 9: Zero-Markdown & Protocol Enforcement
SAAYN treats the LLM as a raw logic provider. 
* **Protocol Exception:** Output containing markdown fences (```) or conversational filler results in `FAILED_VALIDATION`. 
* **No Auto-Cleaning:** The agent shall not attempt to strip markdown; the model must comply with the Zero-Markdown Protocol.

---

## Chapter 10: Sovereign Licensing & Command Reference
**License:** Functional Source License (FSL-1.1-Apache-2.0).
* **`saayn init`**: Setup repo.
* **`saayn plan`**: Generate a Proposal (DRAFT).
* **`saayn review`**: Validate and present Proposal for approval.
* **`saayn edit`**: Execute approved Proposal.
* **`saayn undo`**: Rollback last execution.

---

## Chapter 11: AI ↔ SAAYN Primitive Protocol
The AI interacts with SAAYN via a strict **Request-Verify-Mutate** loop.

### 11.1 Operation Logic Rules
* **Lexical Array Order:** Operations are processed in the order they appear in the JSON.
* **No-Op Detection:** If `replacement_code` produces identical `content_hash`, reject as `NO_OP`.
* **Illegal Combinations:**
    * `CHUNK_DELETE` followed by `CHUNK_REPLACE` on the same UUID = **Invalid**.
    * Duplicate `CHUNK_REPLACE` on the same UUID = **Invalid** (Amorphous Intent).

---

## Chapter 12: CLI → State Machine Mapping
| Command | Entry State | Exit State (Success) | Mutation |
| :--- | :--- | :--- | :--- |
| `saayn plan` | `IDLE` | `DRAFT` | Writes `human` section. |
| `saayn review`| `DRAFT` | `PENDING_APPROVAL` | Runs `VALIDATING` -> `VALIDATED`. |
| `saayn approve`| `PENDING_APPROVAL` | `APPROVED` | Updates `saayn.approval`. |
| `saayn edit` | `APPROVED` | `EXECUTED` | Mutates Source + Registry. |

---

## Chapter 13: Canonicalization & Sealing (The Seal)
The `saayn` section must be canonicalized before hashing:
1.  **Key Ordering:** Lexicographical (A-Z).
2.  **Whitespace:** Minified JSON (zero whitespace).
3.  **Encoding:** UTF-8 with **no BOM**.

---

## Chapter 14: The Error Model
| Error Class | Exit Code | Description |
| :--- | :--- | :--- |
| `VALIDATION_ERROR` | 10 | Schema or Syntax failure. |
| `TAMPER_ERROR` | 20 | Seal mismatch (Manual metadata edit). |
| `DRIFT_ERROR` | 30 | Disk changed since registry update. |
| `EXECUTION_ERROR` | 40 | Filesystem write or `fsync` failure. |

