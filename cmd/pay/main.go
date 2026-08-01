// Command pay drives one hardcoded purchase end to end: create a Prava session,
// have a human approve it, use the issued card to complete checkout against the
// merchant over UCP, then report the real checkout outcome back to Prava.
package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/prava/hackathon-agent/internal/payments"
	"github.com/prava/hackathon-agent/internal/reputation"
	"github.com/prava/hackathon-agent/internal/risk"
	"github.com/prava/hackathon-agent/internal/ucp"
)

// Hardcoded test purchase: Sennheiser HD 600, headphone-zone, as seen live via UCP search_catalog/get_product.
const (
	productDescription = "Sennheiser - HD 600"
	productVariantID   = "gid://shopify/ProductVariant/39651754606780"
	unitPrice          = "22990.00"
	unitPriceRupees    = 22990
	currency           = "INR"
	merchantName       = "Headphone Zone"
	merchantURL        = "https://headphone-zone.myshopify.com"
	merchantMCPURL     = "https://headphone-zone.myshopify.com/api/ucp/mcp"
	merchantCountry    = "IN"
	testUserID         = "test_user_1"
	testUserEmail      = "test@example.com"

	// Wallet whose SwarmPay reputation gates this purchase. Override with
	// WALLET_ADDRESS to exercise the other two decision paths — see README.
	defaultWalletAddress = "0x9cab350e3485e3981b0e729d3cfdb90992a56e9c" // real C-tier wallet -> blocked
)

func main() {
	_ = godotenv.Load()

	// 0. SwarmPay trust gate — runs before any Prava session is created.
	// A blocked wallet must never reach create-session; a human-review wallet
	// still creates a session (so a human has something to approve against)
	// but the code stops short of driving it further on its own.
	swarmpayURL := os.Getenv("SWARMPAY_API_URL")
	if swarmpayURL == "" {
		swarmpayURL = "http://localhost:8080"
	}
	walletAddress := os.Getenv("WALLET_ADDRESS")
	if walletAddress == "" {
		walletAddress = defaultWalletAddress
	}

	repClient := reputation.NewClient(swarmpayURL, os.Getenv("SWARMPAY_API_KEY"))
	repScore, err := repClient.GetScore(walletAddress)
	if err != nil {
		log.Fatalf("swarmpay reputation check failed: %v", err)
	}

	normalizedScore := repScore.ToNormalized()
	policy := risk.Evaluate(repScore.Known, normalizedScore, unitPriceRupees)

	fmt.Println("=== 0. SwarmPay trust gate ===")
	fmt.Printf("wallet:           %s\n", walletAddress)
	fmt.Printf("known:            %v\n", repScore.Known)
	fmt.Printf("raw_score:        %d (tier %s)\n", repScore.RawScore, repScore.Tier)
	fmt.Printf("score (0-100):    %d\n", policy.Score)
	fmt.Printf("spend_limit:      Rs.%d\n", policy.SpendLimit)
	fmt.Printf("purchase_amount:  Rs.%d\n", unitPriceRupees)
	fmt.Printf("decision:         %s\n", policy.Decision)
	fmt.Printf("reason:           %s\n\n", policy.Reason)

	switch policy.Decision {
	case risk.DecisionBlock:
		log.Fatalf("wallet %s is blocked by SwarmPay trust gate — stopping before payment", walletAddress)
	case risk.DecisionHumanReview:
		fmt.Println("Purchase amount exceeds this wallet's auto-approve limit — human review required.")
		fmt.Println("Press Enter to approve and continue, or Ctrl+C to abort...")
		bufio.NewReader(os.Stdin).ReadString('\n')
	case risk.DecisionApprove:
		fmt.Println("Auto-approved under spend limit — continuing.")
	}

	baseURL := os.Getenv("PRAVA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://sandbox.api.prava.space"
	}
	secretKey := os.Getenv("PRAVA_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("PRAVA_SECRET_KEY is not set")
	}

	client := payments.NewClient(baseURL, secretKey)

	// 1. Create session
	sessionReq := payments.CreateSessionRequest{
		UserID:      testUserID,
		UserEmail:   testUserEmail,
		TotalAmount: unitPrice,
		Currency:    currency,
		PurchaseContext: []payments.PurchaseContext{
			{
				MerchantDetails: payments.MerchantDetails{
					Name:            merchantName,
					URL:             merchantURL,
					CountryCodeISO2: merchantCountry,
				},
				ProductDetails: []payments.ProductDetails{
					{
						Description: productDescription,
						UnitPrice:   unitPrice,
						Quantity:    1,
					},
				},
			},
		},
		IntegrationType: "full_checkout",
	}

	fmt.Println("=== 1. create-session ===")
	session, res, err := client.CreateSession(sessionReq)
	logResult("create-session", res, err)
	if err != nil {
		log.Fatalf("create-session failed: %v", err)
	}

	fmt.Printf("\nsession_id: %s\norder_id:   %s\nexpires_at: %s\n", session.SessionID, session.OrderID, session.ExpiresAt)
	fmt.Printf("\nOpen this URL in a browser and complete card entry (sandbox OTP: 456789):\n\n  %s\n\n", session.IframeURL)
	fmt.Println("Press Enter once you've approved the payment...")
	bufio.NewReader(os.Stdin).ReadString('\n')

	// 2. Poll for payment result until a token is issued (or terminal failure).
	fmt.Println("\n=== 2. poll get-payment-result ===")
	var result *payments.PaymentResultResponse
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		r, res, err := client.GetPaymentResult(session.SessionID)
		logResult("get-payment-result", res, err)
		if err != nil {
			log.Fatalf("get-payment-result failed: %v", err)
		}
		result = r
		fmt.Printf("status: %s\n", result.Status)

		if result.Status == "awaiting_result" || result.Status == "completed" {
			break
		}
		if result.Status == "failed" {
			log.Fatal("payment session failed before a token was issued")
		}
		time.Sleep(3 * time.Second)
	}
	if result == nil || len(result.Transactions) == 0 || len(result.Transactions[0].LineItems) == 0 {
		log.Fatal("timed out waiting for payment result / no line items returned")
	}

	lineItem := result.Transactions[0].LineItems[0]
	if lineItem.Token == "" {
		log.Fatal("no payment token issued — approval likely incomplete")
	}
	fmt.Printf("token issued: %s (exp %s/%s)\n", lineItem.Token, lineItem.ExpiryMonth, lineItem.ExpiryYear)

	// 3. Spend the issued card at the merchant over UCP: create a checkout, then
	// submit the card as a payment instrument to complete it. Whatever happens here —
	// success or failure — is the real result we report back to Prava in step 4.
	fmt.Println("\n=== 3. merchant checkout (UCP) ===")
	merchantClient := ucp.NewClient(merchantName, merchantMCPURL)

	checkoutResult, err := merchantClient.CreateCheckout(
		[]map[string]any{
			{
				"quantity": 1,
				"item":     map[string]any{"id": productVariantID},
			},
		},
		map[string]any{
			"first_name":       "Test",
			"last_name":        "Buyer",
			"street_address":   "221B Baker Street",
			"address_locality": "Mumbai",
			"address_region":   "MH",
			"postal_code":      "400001",
			"address_country":  "IN",
			"phone_number":     "+919999999999",
		},
	)
	logToolResult("create_checkout", checkoutResult, err)
	if err != nil {
		log.Fatalf("create_checkout failed: %v", err)
	}

	// create_checkout can return isError:true alongside a valid, still-usable checkout
	// (e.g. "requires_escalation" for a missing shipping address) — the id is what matters.
	var checkout struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(checkoutResult.StructuredContent, &checkout); err != nil {
		log.Fatalf("decode create_checkout structuredContent: %v", err)
	}
	if checkout.ID == "" {
		log.Fatalf("create_checkout did not return a checkout id: %s", checkoutResult.ErrorText())
	}
	fmt.Printf("checkout_id: %s (status: %s)\n", checkout.ID, checkout.Status)

	expiryMonth, expiryYear := 0, 0
	fmt.Sscanf(lineItem.ExpiryMonth, "%d", &expiryMonth)
	fmt.Sscanf(lineItem.ExpiryYear, "%d", &expiryYear)

	completeResult, err := merchantClient.CompleteCheckout(checkout.ID, idempotencyKey(), []map[string]any{
		{
			"id":         "prava-" + lineItem.TxnRefID,
			"handler_id": "shopify.card", // from create_checkout's advertised payment_handlers (dev.shopify.card)
			"type":       "card",
			"credential": map[string]any{
				"token": lineItem.Token,
				"type":  "merchant.token",
			},
			"display": map[string]any{
				"brand":        "visa",
				"last_digits":  lastDigits(lineItem.Token),
				"expiry_month": expiryMonth,
				"expiry_year":  expiryYear,
			},
			"selected": true,
		},
	})
	// A merchant/network rejection here is expected and is itself the real result —
	// do not fall back to a hardcoded outcome; whatever happened is what gets reported.
	var checkoutSucceeded bool
	var orderID, failureReason string
	if err != nil {
		checkoutSucceeded = false
		failureReason = err.Error()
		logToolResultErr("complete_checkout", err)
	} else {
		logToolResult("complete_checkout", completeResult, nil)
		if completeResult.IsError {
			checkoutSucceeded = false
			failureReason = completeResult.ErrorText()
		} else {
			var completed struct {
				Checkout struct {
					OrderID string `json:"order_id"`
				} `json:"checkout"`
			}
			if err := json.Unmarshal(completeResult.StructuredContent, &completed); err != nil {
				checkoutSucceeded = false
				failureReason = fmt.Sprintf("decode complete_checkout structuredContent: %v", err)
			} else if completed.Checkout.OrderID == "" {
				checkoutSucceeded = false
				failureReason = "complete_checkout returned no order_id"
			} else {
				checkoutSucceeded = true
				orderID = completed.Checkout.OrderID
			}
		}
	}

	if checkoutSucceeded {
		fmt.Printf("merchant order placed: %s\n", orderID)
	} else {
		fmt.Printf("merchant checkout failed: %s\n", failureReason)
	}

	// 4. Report the real checkout outcome back to Prava — never hardcoded.
	fmt.Println("\n=== 4. report-status ===")
	txnStatus := "DECLINED"
	if checkoutSucceeded {
		txnStatus = "APPROVED"
	}
	reportReq := payments.ReportStatusRequest{
		TxnRefID:  lineItem.TxnRefID,
		TxnStatus: txnStatus,
	}
	if !checkoutSucceeded {
		reportReq.ResponseCode = "05" // generic decline; merchant rejected the instrument
	}
	reportResp, res, err := client.ReportStatus(session.SessionID, reportReq)
	logResult("report-status", res, err)
	if err != nil {
		log.Fatalf("report-status failed: %v", err)
	}

	// 5. Final result, screenshottable.
	fmt.Println("\n=== TRANSACTION COMPLETE ===")
	fmt.Printf("order_id:          %s\n", session.OrderID)
	fmt.Printf("session_id:        %s\n", session.SessionID)
	fmt.Printf("merchant_order_id: %s\n", orderID)
	fmt.Printf("txn_status:        %s\n", reportResp.TxnStatus)
	fmt.Printf("visa_confirmation: %s\n", reportResp.VisaConfirmation)
	fmt.Printf("amount:            %s %s\n", unitPrice, currency)
	fmt.Printf("product:           %s\n", productDescription)
	fmt.Printf("merchant:          %s\n", merchantName)
	if !checkoutSucceeded {
		fmt.Printf("merchant checkout failure reason: %s\n", failureReason)
	}
}

func idempotencyKey() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%x", buf)
}

func lastDigits(token string) string {
	if len(token) < 4 {
		return token
	}
	return token[len(token)-4:]
}

func logToolResult(step string, r *ucp.ToolResult, callErr error) {
	pretty := new(bytes.Buffer)
	if json.Indent(pretty, r.Raw, "", "  ") == nil {
		log.Printf("[%s] raw response:\n%s", step, pretty.String())
	} else {
		log.Printf("[%s] raw response (unparsed): %s", step, string(r.Raw))
	}
	if callErr != nil {
		log.Printf("[%s] error: %v", step, callErr)
	}
}

func logToolResultErr(step string, callErr error) {
	log.Printf("[%s] error: %v", step, callErr)
}

func logResult(step string, res payments.Result, callErr error) {
	pretty := new(bytes.Buffer)
	if json.Indent(pretty, res.Body, "", "  ") == nil {
		log.Printf("[%s] status=%d x-response-id=%s raw response:\n%s", step, res.StatusCode, res.ResponseID, pretty.String())
	} else {
		log.Printf("[%s] status=%d x-response-id=%s raw response (unparsed): %s", step, res.StatusCode, res.ResponseID, string(res.Body))
	}
	if callErr != nil {
		log.Printf("[%s] error: %v", step, callErr)
	}
}
