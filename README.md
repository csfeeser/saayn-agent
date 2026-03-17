# SAAYN Agent (`saayn`)

> **The UPC Barcode System for AI-Native Codebases.**

`saayn` is a lightweight, sovereign CLI tool designed to orchestrate the relationship between human intent and AI execution. It enforces the **Code Chunking** architecture—treating your source files as collections of immutable, machine-readable "barcodes" to prevent AI hallucinations and feature drift.

---

## 🛠 The Philosophy: "Day 1 vs. Day 2"

- **Day 1 (Synthesis):** Use a frontier model (Gemini Pro, Claude 3.5) to build your foundation. 
- **Day 2 (Evolution):** Use `saayn` to maintain it. 

By partitioning your code into UUID-bound "chunks," `saayn` allows even small, local models (running on your own hardware) to perform surgical, production-grade edits with the precision of a much larger model.

## 🚀 Key Features

- **Zero Collateral Damage:** Edits are strictly bounded by `CHUNK_START` and `CHUNK_END` markers.
- **Planner/Coder Architecture:** One model finds the target; another model executes the edit. 
- **Sovereign First:** Designed to run against local inference servers (vLLM, Ollama) on your own hardware. 
- **Atomic Swaps:** File writes are transactional. If a build fails, the swap is rejected.

## 📦 Installation

```bash
# Clone the repo
git clone [https://github.com/your-username/saayn-agent](https://github.com/your-username/saayn-agent)
cd saayn-agent

# Build the binary
go build -o saayn main.go
sudo mv saayn /usr/local/bin/
