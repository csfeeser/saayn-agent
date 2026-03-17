# SAAYN Agent (`saayn`)

> **The UPC Barcode System for AI-Native Codebases.**

`saayn` is a lightweight, sovereign CLI tool designed to orchestrate the relationship between human intent and AI execution. It enforces the **Code Chunking** architecture—treating your source files as collections of immutable, machine-readable "barcodes" to prevent AI hallucinations and feature drift.

### The Philosophy: "Day 1 vs. Day 2"

- **Day 1 (Synthesis):** Use a frontier model (Gemini Pro, Claude 3.5) to build your foundation. 
- **Day 2 (Evolution):** Use `saayn` to maintain it. 

By partitioning your code into UUID-bound "chunks," `saayn` allows even small, local models (running on your own hardware) to perform surgical, production-grade edits with the precision of a much larger model.

### Key Features

- **Zero Collateral Damage:** Edits are strictly bounded by `CHUNK_START` and `CHUNK_END` markers.
- **Planner/Coder Architecture:** One model finds the target; another model executes the edit. 
- **Sovereign First:** Designed to run against local inference servers (vLLM, Ollama) on your own hardware. 
- **Atomic Swaps:** File writes are transactional. If a build fails, the swap is rejected.

### Installation

```bash
# Clone the repo
git clone [https://github.com/your-username/saayn-agent](https://github.com/your-username/saayn-agent)
cd saayn-agent

# Build the binary
go build -o saayn main.go
sudo mv saayn /usr/local/bin/
```

### Configuration (12-Factor Style)

Create a `.env` file in your project root (it will be auto-ignored by `saayn init`):

```bash
SAAYN_INFERENCE_URL="http://your-a100-ip:8000/v1"
SAAYN_PLANNER_MODEL="llama-3-8b"
SAAYN_CODER_MODEL="qwen-2.5-coder-32b"
```

### Quick Start

1. **Initialize your project:**
   `saayn init` (This creates your `chunk-registry.json` and scans for existing markers).

2. **Add a new feature:**
   `saayn edit --intent "Add a scroll-to-top button to the history table"`

3. **Heal your boundaries:**
   `saayn heal` (Fixes any broken or missing markers).

### License

Licensed under the **Functional Source License (FSL-1.1-MIT-2.0)**. 

This project is free for individuals and all non-competing use. To protect the project's sovereignty, commercial use that competes with the `saayn` tool is restricted for 2 years, after which the code automatically reverts to the **Apache 2.0** license.
