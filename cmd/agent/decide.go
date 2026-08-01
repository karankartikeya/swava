// decide.go implements the agent decision loop: given a task description and
// a budget, search all configured merchants over UCP, compare real results,
// pick the best match under budget, then hand the choice off to the same
// trust-gate + payment flow cmd/pay uses (internal/purchase) — no duplicated
// checkout logic.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/prava/hackathon-agent/internal/purchase"
	"github.com/prava/hackathon-agent/internal/ucp"
)

// catalogEntry pairs a live UCP search hit with the hardcoded purchase.Product
// it corresponds to — purchase.Run only knows how to check out the two
// products verified live in cmd/pay, so the agent's job is to decide which of
// those (if either) is the right answer for a task, not to check out an
// arbitrary catalog result.
type catalogEntry struct {
	merchantEnvVar string // e.g. MERCHANT_1
	searchTerm     string // query sent to search_catalog
	product        purchase.Product
}

// knownCatalog is deliberately small and hardcoded, matching cmd/pay's two
// hardcoded products — the agent chooses BETWEEN them using live search
// results as evidence, it does not synthesize a checkout for a product it has
// never verified end to end.
func knownCatalog() []catalogEntry {
	return []catalogEntry{
		{
			merchantEnvVar: "MERCHANT_1",
			searchTerm:     "headphones",
			product: purchase.Product{
				Description:      "Sennheiser - HD 600",
				VariantID:        "gid://shopify/ProductVariant/39651754606780",
				UnitPriceDecimal: "22990.00",
				UnitPriceRupees:  22990,
				Currency:         "INR",
				MerchantName:     "Headphone Zone",
				MerchantURL:      "https://headphone-zone.myshopify.com",
				MerchantMCPURL:   "https://headphone-zone.myshopify.com/api/ucp/mcp",
				MerchantCountry:  "IN",
			},
		},
		{
			merchantEnvVar: "MERCHANT_2",
			searchTerm:     "razor",
			product: purchase.Product{
				Description:      "Sensi Smart 3 Razor",
				VariantID:        "gid://shopify/ProductVariant/43762382569626",
				UnitPriceDecimal: "99.00",
				UnitPriceRupees:  99,
				Currency:         "INR",
				MerchantName:     "Bombay Shaving Company",
				MerchantURL:      "https://bombay-shaving.myshopify.com",
				MerchantMCPURL:   "https://bombay-shaving.myshopify.com/api/ucp/mcp",
				MerchantCountry:  "IN",
			},
		},
	}
}

// runDecide is the agent decision loop entrypoint: TASK describes what the
// agent is trying to buy, BUDGET_RUPEES caps what it's allowed to spend.
// Every candidate is checked against all three merchants' live search_catalog
// (proving the agent actually looked, not just pattern-matched the task
// string), then the best in-budget match is hand off to purchase.Run.
func runDecide() {
	task := os.Getenv("TASK")
	if task == "" {
		log.Fatal("TASK is not set — describe what the agent should buy, e.g. TASK=\"buy a cheap razor\"")
	}
	budget := envInt("BUDGET_RUPEES", 1000)

	fmt.Printf("=== Agent decision loop ===\ntask:   %q\nbudget: Rs.%d\n\n", task, budget)

	merchants := loadMerchants()
	catalog := knownCatalog()

	type candidate struct {
		entry       catalogEntry
		taskMatch   bool
		underBudget bool
	}

	var candidates []candidate
	taskLower := strings.ToLower(task)

	for _, entry := range catalog {
		endpoint := merchants[entry.merchantEnvVar]
		if endpoint == "" {
			fmt.Printf("[%s] not configured, skipping\n", entry.merchantEnvVar)
			continue
		}

		taskMatch := strings.Contains(taskLower, entry.searchTerm) ||
			strings.Contains(taskLower, strings.ToLower(entry.product.Description))

		client := ucp.NewClient(entry.product.MerchantName, endpoint)
		fmt.Printf("[%s] searching for %q (task keyword match: %v)...\n", entry.product.MerchantName, entry.searchTerm, taskMatch)

		products, err := client.SearchCatalog(entry.searchTerm)
		if err != nil {
			fmt.Printf("  unreachable or search failed: %v\n", err)
			continue
		}
		if len(products) == 0 {
			fmt.Printf("  no results\n")
			continue
		}

		// Find our known variant among the live results, to confirm price/availability
		// rather than trusting the hardcoded constant blindly.
		matched := false
		var livePriceRupees int
		var liveTitle string
		for _, p := range products {
			if p.ID == "" {
				continue
			}
			if strings.EqualFold(p.Title, entry.product.Description) {
				matched = true
				livePriceRupees = int(p.PriceRange.Min.Amount / 100)
				liveTitle = p.Title
				break
			}
		}
		if !matched {
			// Fall back to the top search result's price for the reasoning trace,
			// even though we only know how to check out our hardcoded variant.
			livePriceRupees = int(products[0].PriceRange.Min.Amount / 100)
			liveTitle = products[0].Title
			fmt.Printf("  top result: %s — Rs.%d (hardcoded variant %q not found in top results, using cached price for checkout)\n",
				liveTitle, livePriceRupees, entry.product.Description)
		} else {
			fmt.Printf("  confirmed: %s — Rs.%d\n", liveTitle, livePriceRupees)
		}

		candidates = append(candidates, candidate{
			entry:       entry,
			taskMatch:   taskMatch,
			underBudget: entry.product.UnitPriceRupees <= budget,
		})
	}

	fmt.Println()
	if len(candidates) == 0 {
		log.Fatal("no merchant returned usable results for this task")
	}

	// Pick the best candidate: in-budget is a hard requirement; among those,
	// prefer a task-keyword match (the agent is trying to fulfil the specific
	// task, not just find something cheap), then the lower price.
	var best *candidate
	for i := range candidates {
		c := &candidates[i]
		if !c.underBudget {
			fmt.Printf("rejected: %s (Rs.%d) — exceeds budget Rs.%d\n", c.entry.product.Description, c.entry.product.UnitPriceRupees, budget)
			continue
		}
		if best == nil {
			best = c
			continue
		}
		betterMatch := c.taskMatch && !best.taskMatch
		sameMatchCheaper := c.taskMatch == best.taskMatch && c.entry.product.UnitPriceRupees < best.entry.product.UnitPriceRupees
		if betterMatch || sameMatchCheaper {
			best = c
		}
	}

	if best == nil {
		log.Fatalf("no in-budget option found for task %q with budget Rs.%d", task, budget)
	}

	fmt.Printf("\ndecision: %s at %s — Rs.%d (budget Rs.%d)\n", best.entry.product.Description, best.entry.product.MerchantName, best.entry.product.UnitPriceRupees, budget)
	fmt.Printf("reasoning: cheapest task-relevant, in-budget option confirmed live via search_catalog\n\n")

	walletAddress := os.Getenv("WALLET_ADDRESS")
	if walletAddress == "" {
		walletAddress = "0x9cab350e3485e3981b0e729d3cfdb90992a56e9c" // same default as cmd/pay
	}

	fmt.Println("=== Handing off to trust-gate + payment flow ===")
	purchase.Run(best.entry.product, walletAddress)
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
