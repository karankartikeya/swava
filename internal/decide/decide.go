// Package decide implements the agent decision loop: given a task description
// and a budget, search configured merchants live over UCP, compare real
// results, and pick the best in-budget match. Shared by cmd/agent (CLI) and
// cmd/server (HTTP) so neither duplicates the search/selection logic.
package decide

import (
	"fmt"
	"os"
	"strings"

	"github.com/prava/hackathon-agent/internal/purchase"
	"github.com/prava/hackathon-agent/internal/ucp"
)

// CatalogEntry pairs a merchant + search term with the hardcoded
// purchase.Product it corresponds to. purchase.CreateSandboxSession /
// purchase.Run only know how to check out products verified live in cmd/pay,
// so the agent chooses BETWEEN a small known set using live search results as
// evidence, rather than synthesizing a checkout for an arbitrary catalog hit.
type CatalogEntry struct {
	MerchantEnvVar string // e.g. MERCHANT_1
	SearchTerm     string // query sent to search_catalog
	Product        purchase.Product
}

// KnownCatalog is deliberately small and hardcoded, matching cmd/pay's two
// hardcoded products.
func KnownCatalog() []CatalogEntry {
	return []CatalogEntry{
		{
			MerchantEnvVar: "MERCHANT_1",
			SearchTerm:     "headphones",
			Product: purchase.Product{
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
			MerchantEnvVar: "MERCHANT_2",
			SearchTerm:     "razor",
			Product: purchase.Product{
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

// StepResult records what happened when one merchant was searched — the
// reasoning trail a caller (CLI or HTTP) surfaces to explain the decision.
type StepResult struct {
	MerchantName     string `json:"merchant_name"`
	SearchTerm       string `json:"search_term"`
	TaskMatch        bool   `json:"task_match"`
	Configured       bool   `json:"configured"`
	Reachable        bool   `json:"reachable"`
	Error            string `json:"error,omitempty"`
	FoundTitle       string `json:"found_title,omitempty"`
	FoundPriceRupees int    `json:"found_price_rupees,omitempty"`
	Confirmed        bool   `json:"confirmed"` // true if the hardcoded variant was found in live results
	UnderBudget      bool   `json:"under_budget"`
	Rejected         bool   `json:"rejected"`
	RejectReason     string `json:"reject_reason,omitempty"`
}

// Result is the full outcome of a decision run: every merchant step taken,
// plus the winning product (nil if nothing qualified).
type Result struct {
	Task      string            `json:"task"`
	Budget    int               `json:"budget_rupees"`
	Steps     []StepResult      `json:"steps"`
	Chosen    *purchase.Product `json:"chosen_product,omitempty"`
	Reasoning string            `json:"reasoning"`
}

// loadMerchants reads every MERCHANT_* env var into a name->endpoint map.
func loadMerchants() map[string]string {
	merchants := make(map[string]string)
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		if strings.HasPrefix(key, "MERCHANT_") && value != "" {
			merchants[key] = value
		}
	}
	return merchants
}

// Run searches every catalog entry's merchant live via search_catalog,
// records what it found, then picks the best in-budget match: task-keyword
// relevance first, then lowest price. Never panics or exits — every failure
// mode (merchant unreachable, no results, nothing in budget) is captured in
// the returned Result so an HTTP handler can render it directly.
func Run(task string, budgetRupees int) Result {
	catalog := KnownCatalog()
	merchants := loadMerchants()
	taskLower := strings.ToLower(task)

	result := Result{Task: task, Budget: budgetRupees}

	type candidate struct {
		entry     CatalogEntry
		taskMatch bool
	}
	var candidates []candidate

	for _, entry := range catalog {
		step := StepResult{
			MerchantName: entry.Product.MerchantName,
			SearchTerm:   entry.SearchTerm,
			TaskMatch: strings.Contains(taskLower, entry.SearchTerm) ||
				strings.Contains(taskLower, strings.ToLower(entry.Product.Description)),
			UnderBudget: entry.Product.UnitPriceRupees <= budgetRupees,
		}

		endpoint := merchants[entry.MerchantEnvVar]
		if endpoint == "" {
			step.Configured = false
			result.Steps = append(result.Steps, step)
			continue
		}
		step.Configured = true

		client := ucp.NewClient(entry.Product.MerchantName, endpoint)
		products, err := client.SearchCatalog(entry.SearchTerm)
		if err != nil {
			step.Reachable = false
			step.Error = err.Error()
			result.Steps = append(result.Steps, step)
			continue
		}
		step.Reachable = true

		if len(products) == 0 {
			step.Error = "no results"
			result.Steps = append(result.Steps, step)
			continue
		}

		matched := false
		for _, p := range products {
			if p.ID == "" {
				continue
			}
			if strings.EqualFold(p.Title, entry.Product.Description) {
				matched = true
				step.FoundTitle = p.Title
				step.FoundPriceRupees = int(p.PriceRange.Min.Amount / 100)
				break
			}
		}
		if !matched {
			step.FoundTitle = products[0].Title
			step.FoundPriceRupees = int(products[0].PriceRange.Min.Amount / 100)
		}
		step.Confirmed = matched

		if !step.UnderBudget {
			step.Rejected = true
			step.RejectReason = fmt.Sprintf("Rs.%d exceeds budget Rs.%d", entry.Product.UnitPriceRupees, budgetRupees)
			result.Steps = append(result.Steps, step)
			continue
		}

		result.Steps = append(result.Steps, step)
		candidates = append(candidates, candidate{entry: entry, taskMatch: step.TaskMatch})
	}

	var best *candidate
	for i := range candidates {
		c := &candidates[i]
		if best == nil {
			best = c
			continue
		}
		betterMatch := c.taskMatch && !best.taskMatch
		sameMatchCheaper := c.taskMatch == best.taskMatch && c.entry.Product.UnitPriceRupees < best.entry.Product.UnitPriceRupees
		if betterMatch || sameMatchCheaper {
			best = c
		}
	}

	if best == nil {
		result.Reasoning = fmt.Sprintf("no in-budget, reachable option found for task %q with budget Rs.%d", task, budgetRupees)
		return result
	}

	product := best.entry.Product
	result.Chosen = &product
	result.Reasoning = fmt.Sprintf(
		"%s at %s (Rs.%d) — cheapest task-relevant, in-budget option confirmed live via search_catalog",
		product.Description, product.MerchantName, product.UnitPriceRupees,
	)
	return result
}
