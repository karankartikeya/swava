// Command pay drives one hardcoded purchase end to end through the SwarmPay
// trust gate, Prava sandbox, and merchant checkout over UCP. Which product is
// bought is selected via PRODUCT_CHOICE — see internal/purchase for the flow.
package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/prava/hackathon-agent/internal/purchase"
)

// Two hardcoded, live-verified products. headphones deliberately exceeds
// every trust tier's spend limit (forces human_review/blocked paths); shaving
// comfortably clears the 70+ tier's Rs.10,000 limit (exercises auto_approve).
var products = map[string]purchase.Product{
	"headphones": {
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
	"shaving": {
		// Sensi Smart 3 Razor — confirmed live via search_catalog/get_product
		// against bombay-shaving.myshopify.com.
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
}

// Wallet whose SwarmPay reputation gates this purchase. Override with
// WALLET_ADDRESS to exercise the other two decision paths — see .env.example.
const defaultWalletAddress = "0x9cab350e3485e3981b0e729d3cfdb90992a56e9c" // real C-tier wallet -> blocked

func main() {
	_ = godotenv.Load()

	choice := os.Getenv("PRODUCT_CHOICE")
	if choice == "" {
		choice = "headphones"
	}
	product, ok := products[choice]
	if !ok {
		log.Fatalf("unknown PRODUCT_CHOICE %q — must be one of: headphones, shaving", choice)
	}

	walletAddress := os.Getenv("WALLET_ADDRESS")
	if walletAddress == "" {
		walletAddress = defaultWalletAddress
	}

	purchase.Run(product, walletAddress)
}
