// decide.go is the CLI entrypoint for the agent decision loop: given TASK and
// BUDGET_RUPEES, it calls internal/decide.Run (shared with cmd/server), prints
// the reasoning trail, then hands the chosen product to internal/purchase.Run
// — the same trust-gate + payment flow cmd/pay uses.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/prava/hackathon-agent/internal/decide"
	"github.com/prava/hackathon-agent/internal/purchase"
)

// runDecide is the agent decision loop entrypoint: TASK describes what the
// agent is trying to buy, BUDGET_RUPEES caps what it's allowed to spend.
func runDecide() {
	task := os.Getenv("TASK")
	if task == "" {
		log.Fatal("TASK is not set — describe what the agent should buy, e.g. TASK=\"buy a cheap razor\"")
	}
	budget := envInt("BUDGET_RUPEES", 1000)

	fmt.Printf("=== Agent decision loop ===\ntask:   %q\nbudget: Rs.%d\n\n", task, budget)

	result := decide.Run(task, budget)

	for _, step := range result.Steps {
		if !step.Configured {
			fmt.Printf("[%s] not configured, skipping\n", step.MerchantName)
			continue
		}
		fmt.Printf("[%s] searching for %q (task keyword match: %v)...\n", step.MerchantName, step.SearchTerm, step.TaskMatch)
		if !step.Reachable {
			fmt.Printf("  unreachable or search failed: %s\n", step.Error)
			continue
		}
		if step.FoundTitle == "" {
			fmt.Printf("  no results\n")
			continue
		}
		if step.Confirmed {
			fmt.Printf("  confirmed: %s — Rs.%d\n", step.FoundTitle, step.FoundPriceRupees)
		} else {
			fmt.Printf("  top result: %s — Rs.%d (hardcoded variant not found in top results, using cached price for checkout)\n", step.FoundTitle, step.FoundPriceRupees)
		}
		if step.Rejected {
			fmt.Printf("rejected: %s\n", step.RejectReason)
		}
	}
	fmt.Println()

	if result.Chosen == nil {
		log.Fatal(result.Reasoning)
	}

	fmt.Printf("decision: %s\nreasoning: %s\n\n", result.Chosen.Description, result.Reasoning)

	walletAddress := os.Getenv("WALLET_ADDRESS")
	if walletAddress == "" {
		walletAddress = "0x9cab350e3485e3981b0e729d3cfdb90992a56e9c" // same default as cmd/pay
	}

	fmt.Println("=== Handing off to trust-gate + payment flow ===")
	purchase.Run(*result.Chosen, walletAddress)
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	var n int
	if _, err := fmt.Sscanf(raw, "%d", &n); err != nil {
		return fallback
	}
	return n
}
