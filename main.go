package main

// SAAYN:CHUNK_START:main-entry-v1-m1a2i3n4
// BUSINESS_PURPOSE: The final entry point for the saayn-agent. 
// It initializes the CLI and hands off control to the Cobra command router.
// SPEC_LINK: SpecBook v1.7 Chapter 0
import (
	"saayn/cmd"
)

func main() {
	// Ignition: Execute the root command and all registered subcommands.
	cmd.Execute()
}
// SAAYN:CHUNK_END:main-entry-v1-m1a2i3n4
