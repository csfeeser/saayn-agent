package cmd

// SAAYN:CHUNK_START:edit-imports-v1-a1b2c3d4
// BUSINESS_PURPOSE: CLI command imports. Connects the Cobra CLI to the internal AI and Transaction packages.
// SPEC_LINK: SpecBook v1.7 Chapter 2 & 5
import (
	"fmt"
	"os"

	"github.com/sfeeser/saayn-agent/internal/ai"
	"github.com/sfeeser/saayn-agent/internal/transaction"
	"github.com/spf13/cobra"
)

// SAAYN:CHUNK_END:edit-imports-v1-a1b2c3d4

// SAAYN:CHUNK_START:edit-command-definition-v1-e5f6g7h8
// BUSINESS_PURPOSE: Defines the 'edit' command flags and the main execution loop.
// SPEC_LINK: SpecBook v1.7 Chapter 2 & 9
var (
	intent      string
	autoApprove bool
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Perform a surgical AI edit on targeted code chunks",
	Run: func(cmd *cobra.Command, args []string) {
		runEdit()
	},
}

func init() {
	editCmd.Flags().StringVarP(&intent, "intent", "i", "", "The natural language intent for the edit (required)")
	editCmd.MarkFlagRequired("intent")
	editCmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip manual confirmation of the AI plan")
	rootCmd.AddCommand(editCmd)
}

// SAAYN:CHUNK_END:edit-command-definition-v1-e5f6g7h8

// SAAYN:CHUNK_START:edit-execution-loop-v1.8-i9j0k1l2
// BUSINESS_PURPOSE: The Master Loop. Orchestrates the v1.8 State Machine from IDLE to EXECUTED.
// SPEC_LINK: SpecBook v1.8 Chapter 6 (The State Machine)
func runEdit() {
	// 1. Establish Environment (v1.8 Preparation)
	workDir, _ := os.Getwd()
	reg := loadRegistry() // Load the Inventory first so the Engine can verify it
	opID := fmt.Sprintf("op-%d", os.Getpid())

	// 2. Setup Engine & Acquire Lock (Steps 1-2)
	// Pass the Registry (reg) and Path (workDir) into the new v1.8 Constructor
	engine := transaction.NewEngine(reg, workDir, opID)
	if err := engine.AcquireLock(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer engine.ReleaseLock()

	// 3. Plan (Steps 3-4)
	planner := ai.NewPlanner(os.Getenv("SAAYN_PLANNER_MODEL"), os.Getenv("SAAYN_INFERENCE_URL"))

	fmt.Println("🤖 Planning edit...")
	plan, err := planner.Plan(reg, intent)
	if err != nil {
		fmt.Printf("❌ Planning failed: %v\n", err)
		return
	}

	// 4. Human Approval Gate (Chapter 5)
	if !autoApprove {
		confirmPlan(plan) // Helper to print justifications and wait for Y/N
	}

	// 5. Synthesis & Staging (Steps 5-10)
	coder := ai.NewCoder(os.Getenv("SAAYN_CODER_MODEL"), os.Getenv("SAAYN_INFERENCE_URL"))
	var actions []transaction.FileAction

	for _, item := range plan {
		chunk := findChunk(reg, item.UUID)
		fmt.Printf("✍️  Coding chunk: %s...\n", chunk.UUID)

		newCode, err := coder.Generate(chunk, intent)
		if err != nil {
			fmt.Printf("❌ Generation failed for %s: %v\n", chunk.UUID, err)
			return
		}

		// Stage .tmp file (Step 8) - Now supported by the Engine's v1.8 Bridge
		action, err := engine.StageChunk(chunk, newCode)
		if err != nil {
			fmt.Printf("❌ Staging failed: %v\n", err)
			return
		}
		actions = append(actions, action)
	}

	// 6. The Atomic Commit (Steps 11-16)
	fmt.Println("💾 Committing changes to filesystem...")
	// Commit now triggers the v1.8 'StateApproved' -> 'StateExecuting' flow
	if err := engine.Commit(actions); err != nil {
		fmt.Printf("❌ Critical Error during commit: %v\n", err)
		fmt.Println("⚠️  Run 'saayn verify' to check codebase integrity.")
		return
	}

	fmt.Println("✅ Edit successful. Codebase synchronized.")
}

// SAAYN:CHUNK_END:edit-execution-loop-v1.8-i9j0k1l2
