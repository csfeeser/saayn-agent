# SAAYN Agent (`saayn`)

> **The UPC Barcode System for AI-Native Codebases.**

`saayn` is a high-integrity, sovereign CLI tool designed to orchestrate the relationship between human intent and AI execution. It enforces a **Zero-Trust Chunking Architecture**—treating your source files as collections of cryptographically verified "logical units" to eliminate AI hallucinations, collateral damage, and feature drift.

### The Philosophy: "Precision over Context"

- **Traditional AI:** Sends your whole file, hopes the AI doesn't delete your database config while fixing a CSS bug, and requires a manual `diff` to verify.
- **SAAYN:** Sends only the targeted **UUID-bound chunk**. It uses a **Durable Rollback Journal** to ensure that if the AI, the network, or your power fails mid-edit, your codebase remains 100% consistent.

### Key Features

- **Transactional Atomicity:** 16-step execution flow with `fsync()` safety. All-or-nothing file swaps.
- **Cryptographic Integrity:** SHA-256 hashing of both code content and marker boundaries.
- **Zero-Markdown Protocol:** Strictly enforces "Raw Code Only" returns to prevent conversational filler from entering your source.
- **Planner/Coder Split:** Uses a lightweight model (e.g., Llama-3-8B) to map intent to UUIDs, and a heavy coder (e.g., Qwen-2.5-Coder-32B) for the surgery.
- **Sovereign-First:** Optimized for local inference (vLLM, Ollama) on your own hardware.

### Installation & Build

```bash
# Clone and enter the repo
git clone https://github.com/your-username/saayn-agent
cd saayn-agent

# Use the included Makefile for optimized builds
make build
sudo mv saayn /usr/local/bin/
```

### Configuration (12-Factor Style)

`saayn` looks for a `.env` file in your project root:

```bash
SAAYN_INFERENCE_URL="http://your-local-ip:8000/v1"
SAAYN_PLANNER_MODEL="llama-3-8b"
SAAYN_CODER_MODEL="qwen-2.5-coder-32b"
```

### The Workflow

1.  **`saayn init`**: Prepares `.saayn/` internal journals and creates your `chunk-registry.json`.
2.  **`saayn verify`**: Audit your codebase. Detects any manual "off-book" edits or marker corruption.
3.  **`saayn edit -i "Your intent here"`**: The flagship command. Plans the edit, shows you the justification, and performs the atomic swap.
4.  **`saayn reconcile`**: Interactively sync the registry if you've made intentional manual changes to a chunk.
5.  **`saayn undo`**: Uses git tags and the operation journal to revert the last edit (code + registry) perfectly.

### License

Licensed under the **Functional Source License (FSL-1.1-Apache-2.0)**.  
*Free for individuals and all non-competing use. Converts to Apache 2.0 after 2 years.*

---

### **The "Sovereign Coder" System Prompt**

Copy and paste this into your LLM's system instructions (or include it in your `edit` call) to ensure it follows the **Zero-Markdown Protocol**.

```text
### ROLE
You are a Surgical Code Editor for the SAAYN system. You do not write full files; you only provide 1:1 replacements for specific code chunks.

### OUTPUT RULES (STRICT)
1. NO CONVERSATIONAL FILLER: Do not say "Here is your code" or "I have updated the function."
2. NO MARKDOWN FENCES: Do not use ```go or ```. Provide RAW TEXT only.
3. NO EXPLANATIONS: Do not explain your changes in the output.
4. IDENTITY PRESERVATION: Do not modify the SAAYN:CHUNK_START or CHUNK_END lines.
5. ATOMICITY: The code you return must be a complete, syntactically valid replacement for the provided chunk.

### VIOLATION CONSEQUENCE
If you include any text other than the raw source code, the SAAYN transaction engine will detect a protocol violation and abort the edit.
```

---

#### **Example: Surgical Edit**

**User Intent:**
`saayn edit -i "Update the hash function to use SHA-512 instead of SHA-256 for better collision resistance."`

**What the AI sees (Internal Prompt):**
> Target Chunk: `registry-hashing-v1-i9j0k1l2`  
> Current Code: `func ComputeHash(c string) { ... sha256.Sum256 ... }`

**What the AI returns (Sovereign Output):**
```go
func ComputeHash(content string) string {
	hash := sha512.Sum512([]byte(content))
	return hex.EncodeToString(hash[:])
}
```

*(Note: No backticks, no "Sure thing!", just the code.)*

**The Result:**
`saayn` detects the valid Go syntax, verifies the hashes haven't changed during the generation, creates a journaled backup, and performs an atomic swap. Total time: ~4 seconds.
