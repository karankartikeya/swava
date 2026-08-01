// Command agent connects to Shopify UCP merchant endpoints and lists products.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/prava/hackathon-agent/internal/ucp"
)

const searchQuery = "headphones"

func main() {
	_ = godotenv.Load()

	if os.Getenv("TASK") != "" {
		runDecide()
		return
	}

	merchants := loadMerchants()
	if len(merchants) == 0 {
		log.Fatal("no MERCHANT_* environment variables set")
	}

	for name, endpoint := range merchants {
		fmt.Printf("\n=== %s (%s) ===\n", name, endpoint)

		client := ucp.NewClient(name, endpoint)

		tools, err := client.ListTools()
		if err != nil {
			log.Printf("[%s] unreachable or discovery failed: %v", name, err)
			continue
		}
		fmt.Printf("tools available: %s\n", toolNames(tools))

		products, err := client.SearchCatalog(searchQuery)
		if err != nil {
			log.Printf("[%s] search_catalog failed: %v", name, err)
			continue
		}
		if len(products) == 0 {
			fmt.Printf("no products found for %q\n", searchQuery)
			continue
		}
		for _, p := range products {
			fmt.Printf("  %-40s %10.2f %s   %s\n",
				p.Title,
				float64(p.PriceRange.Min.Amount)/100,
				p.PriceRange.Min.Currency,
				p.ID,
			)
		}
	}
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

func toolNames(tools []ucp.Tool) string {
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = t.Name
	}
	return strings.Join(names, ", ")
}
