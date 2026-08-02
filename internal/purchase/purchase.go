// Package purchase drives one purchase end to end: SwarmPay trust gate, Prava
// sandbox session, human card approval, merchant checkout over UCP, and
// reporting the real outcome back to Prava. Shared by cmd/pay (hardcoded
// product) and cmd/agent (agent-selected product) so neither duplicates it.
package purchase

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/prava/hackathon-agent/internal/payments"
	"github.com/prava/hackathon-agent/internal/reputation"
	"github.com/prava/hackathon-agent/internal/risk"
	"github.com/prava/hackathon-agent/internal/ucp"
)

// Product is everything Run needs to know about what's being bought and where.
type Product struct {
	Description      string `json:"description"`
	VariantID        string `json:"variant_id"`
	UnitPriceDecimal string `json:"unit_price_decimal"` // e.g. "22990.00" — Prava's decimal-string amount format
	UnitPriceRupees  int    `json:"unit_price_rupees"`  // e.g. 22990 — whole rupees, for the risk policy
	Currency         string `json:"currency"`
	MerchantName     string `json:"merchant_name"`
	MerchantURL      string `json:"merchant_url"`
	MerchantMCPURL   string `json:"merchant_mcp_url"`
	MerchantCountry  string `json:"merchant_country"`
}

// Result is the final, screenshottable outcome of a purchase run.
type Result struct {
	PravaOrderID      string
	PravaSessionID    string
	MerchantOrderID   string
	TxnStatus         string
	VisaConfirmation  string
	CheckoutSucceeded bool
	FailureReason     string
}

const (
	testUserID    = "test_user_1"
	testUserEmail = "test@example.com"
)

// TrustGateResult is the outcome of the SwarmPay check alone, independent of
// whether a Prava session followed it.
type TrustGateResult struct {
	WalletAddress   string        `json:"wallet_address"`
	Known           bool          `json:"known"`
	RawScore        int           `json:"raw_score"`
	Tier            string        `json:"tier"`
	NormalizedScore int           `json:"normalized_score"`
	SpendLimit      int           `json:"spend_limit_rupees"`
	Decision        risk.Decision `json:"decision"`
	Reason          string        `json:"reason"`
	// Policy is the configured procurement policy applied for this wallet,
	// if any — omitted entirely when the wallet has no policy configured.
	Policy *risk.ProcurementPolicy `json:"policy,omitempty"`
}

// SessionResult is what CreateSandboxSession returns: the trust-gate outcome,
// plus — only if the gate didn't block — the real Prava session and iframe
// URL a human needs to open to carry the flow further. No blocking I/O of any
// kind; safe to call from an HTTP handler.
type SessionResult struct {
	TrustGate TrustGateResult `json:"trust_gate"`

	// The following are zero-valued when TrustGate.Decision == DecisionBlock,
	// since a blocked wallet must never reach Prava at all.
	PravaSessionID string `json:"prava_session_id,omitempty"`
	PravaOrderID   string `json:"prava_order_id,omitempty"`
	IframeURL      string `json:"iframe_url,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
}

// EvaluateTrustGate runs just the SwarmPay reputation check + risk policy for
// a wallet and purchase amount — no Prava call, no side effects. Shared by
// CreateSandboxSession and the standalone POST /api/trust-gate endpoint so
// there is exactly one code path that talks to SwarmPay. productDescription
// is used only to check any configured procurement-policy keyword block for
// this wallet; pass "" when there's no specific product in play yet.
func EvaluateTrustGate(walletAddress string, amountRupees int, productDescription string) (TrustGateResult, error) {
	swarmpayURL := os.Getenv("SWARMPAY_API_URL")
	if swarmpayURL == "" {
		swarmpayURL = "http://localhost:8080"
	}

	repClient := reputation.NewClient(swarmpayURL, os.Getenv("SWARMPAY_API_KEY"))
	repScore, err := repClient.GetScore(walletAddress)
	if err != nil {
		return TrustGateResult{}, fmt.Errorf("swarmpay reputation check failed: %w", err)
	}

	normalizedScore := repScore.ToNormalized()
	policy := risk.Evaluate(repScore.Known, normalizedScore, amountRupees, walletAddress, productDescription)

	return TrustGateResult{
		WalletAddress:   walletAddress,
		Known:           repScore.Known,
		RawScore:        repScore.RawScore,
		Tier:            repScore.Tier,
		NormalizedScore: policy.Score,
		SpendLimit:      policy.SpendLimit,
		Decision:        policy.Decision,
		Reason:          policy.Reason,
		Policy:          policy.Policy,
	}, nil
}

// CreateSandboxSession runs the trust gate, then — only if the decision isn't
// DecisionBlock — creates a real Prava sandbox session and returns its
// iframe_url. It deliberately stops there: completing the sandbox purchase
// requires a human in a real browser to pass Visa's FIDO/passkey step (no
// server-side path exists — confirmed live, see docs/checkout-flow-status.md),
// so this function never blocks on stdin or polls for completion. Intended
// for cmd/server's POST /api/purchase; the CLI's blocking Run above still
// carries a session all the way through for local use.
func CreateSandboxSession(product Product, walletAddress string) (SessionResult, error) {
	gate, err := EvaluateTrustGate(walletAddress, product.UnitPriceRupees, product.Description)
	if err != nil {
		return SessionResult{}, err
	}

	if gate.Decision == risk.DecisionBlock {
		return SessionResult{TrustGate: gate}, nil
	}

	baseURL := os.Getenv("PRAVA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://sandbox.api.prava.space"
	}
	secretKey := os.Getenv("PRAVA_SECRET_KEY")
	if secretKey == "" {
		return SessionResult{}, fmt.Errorf("PRAVA_SECRET_KEY is not set")
	}

	client := payments.NewClient(baseURL, secretKey)

	sessionReq := payments.CreateSessionRequest{
		UserID:      testUserID,
		UserEmail:   testUserEmail,
		TotalAmount: product.UnitPriceDecimal,
		Currency:    product.Currency,
		PurchaseContext: []payments.PurchaseContext{
			{
				MerchantDetails: payments.MerchantDetails{
					Name:            product.MerchantName,
					URL:             product.MerchantURL,
					CountryCodeISO2: product.MerchantCountry,
				},
				ProductDetails: []payments.ProductDetails{
					{
						Description: product.Description,
						UnitPrice:   product.UnitPriceDecimal,
						Quantity:    1,
					},
				},
			},
		},
		IntegrationType: "full_checkout",
	}

	session, _, err := client.CreateSession(sessionReq)
	if err != nil {
		return SessionResult{}, fmt.Errorf("create-session failed: %w", err)
	}

	return SessionResult{
		TrustGate:      gate,
		PravaSessionID: session.SessionID,
		PravaOrderID:   session.OrderID,
		IframeURL:      session.IframeURL,
		ExpiresAt:      session.ExpiresAt,
	}, nil
}

// Run executes the full flow for one product against one wallet: SwarmPay
// trust gate first (a blocked wallet never reaches Prava), then create-session,
// human card approval, merchant checkout, and report-status with the real
// outcome. Fatal errors (log.Fatal) are used deliberately, matching the rest
// of this hackathon scaffold — this is a one-shot CLI flow, not a server.
func Run(product Product, walletAddress string) Result {
	// 0. SwarmPay trust gate.
	swarmpayURL := os.Getenv("SWARMPAY_API_URL")
	if swarmpayURL == "" {
		swarmpayURL = "http://localhost:8080"
	}

	repClient := reputation.NewClient(swarmpayURL, os.Getenv("SWARMPAY_API_KEY"))
	repScore, err := repClient.GetScore(walletAddress)
	if err != nil {
		log.Fatalf("swarmpay reputation check failed: %v", err)
	}

	normalizedScore := repScore.ToNormalized()
	policy := risk.Evaluate(repScore.Known, normalizedScore, product.UnitPriceRupees, walletAddress, product.Description)

	fmt.Println("=== 0. SwarmPay trust gate ===")
	fmt.Printf("wallet:           %s\n", walletAddress)
	fmt.Printf("known:            %v\n", repScore.Known)
	fmt.Printf("raw_score:        %d (tier %s)\n", repScore.RawScore, repScore.Tier)
	fmt.Printf("score (0-100):    %d\n", policy.Score)
	fmt.Printf("spend_limit:      Rs.%d\n", policy.SpendLimit)
	fmt.Printf("purchase_amount:  Rs.%d\n", product.UnitPriceRupees)
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
		TotalAmount: product.UnitPriceDecimal,
		Currency:    product.Currency,
		PurchaseContext: []payments.PurchaseContext{
			{
				MerchantDetails: payments.MerchantDetails{
					Name:            product.MerchantName,
					URL:             product.MerchantURL,
					CountryCodeISO2: product.MerchantCountry,
				},
				ProductDetails: []payments.ProductDetails{
					{
						Description: product.Description,
						UnitPrice:   product.UnitPriceDecimal,
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
	merchantClient := ucp.NewClient(product.MerchantName, product.MerchantMCPURL)

	checkoutResult, err := merchantClient.CreateCheckout(
		[]map[string]any{
			{
				"quantity": 1,
				"item":     map[string]any{"id": product.VariantID},
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
	fmt.Printf("amount:            %s %s\n", product.UnitPriceDecimal, product.Currency)
	fmt.Printf("product:           %s\n", product.Description)
	fmt.Printf("merchant:          %s\n", product.MerchantName)
	if !checkoutSucceeded {
		fmt.Printf("merchant checkout failure reason: %s\n", failureReason)
	}

	return Result{
		PravaOrderID:      session.OrderID,
		PravaSessionID:    session.SessionID,
		MerchantOrderID:   orderID,
		TxnStatus:         reportResp.TxnStatus,
		VisaConfirmation:  reportResp.VisaConfirmation,
		CheckoutSucceeded: checkoutSucceeded,
		FailureReason:     failureReason,
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
