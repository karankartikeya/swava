// Command server exposes the agent decision loop, SwarmPay trust gate, and
// Prava sandbox session creation over HTTP for the frontend. It is a thin
// layer over internal/decide and internal/purchase — no business logic
// lives here that isn't already in those packages (and thus already
// exercised by cmd/agent and cmd/pay).
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/prava/hackathon-agent/internal/decide"
	"github.com/prava/hackathon-agent/internal/purchase"
	"github.com/prava/hackathon-agent/internal/reputation"
	"github.com/prava/hackathon-agent/internal/risk"
)

// walletRoleNote is included on every /api/decide and /api/purchase response.
// Explaining out loud which side gets scored (the buying agent, never the
// merchant) was the most confusing part of this project when demoed verbally
// — every response says it the same way so the frontend never has to
// re-explain it inconsistently on different pages.
const walletRoleNote = "This wallet is the deployed agent's persistent economic identity — its own on-chain reputation, built from its own transaction history. It is not the end user submitting the task, and it is not the merchant being purchased from. The merchant is never scored."

// demoWallets is the one source of truth for the three wallets the frontend
// offers — labeled here so the frontend never hardcodes what each represents.
var demoWallets = []DemoWallet{
	{Address: "0xaaaa000000000000000000000000000000aaaa", Label: "Established", AgentRole: "Procurement Agent — Established", Description: "Seeded test fixture, score 900/1000 (AAA) — long, clean transaction history."},
	{Address: "0xca4b519063ff7f8154fcb768259bc9059df07237", Label: "New / Moderate", AgentRole: "Procurement Agent — Alpha", Description: "Real indexed wallet, score 600/1000 (AA) — some history, not yet fully trusted."},
	{Address: "0x9cab350e3485e3981b0e729d3cfdb90992a56e9c", Label: "Known-bad", AgentRole: "Procurement Agent — Flagged", Description: "Real indexed wallet, score 220/1000 (C) — poor history, blocked outright."},
}

type DemoWallet struct {
	Address     string `json:"address"`
	Label       string `json:"label"`
	AgentRole   string `json:"agent_role"`
	Description string `json:"description"`
}

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/wallets", withCORS(handleWallets))
	mux.HandleFunc("/api/decide", withCORS(handleDecide))
	mux.HandleFunc("/api/trust-gate", withCORS(handleTrustGate))
	mux.HandleFunc("/api/purchase", withCORS(handlePurchase))
	mux.HandleFunc("/api/agent-profile", withCORS(handleAgentProfile))

	log.Printf("server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// allowedOrigins is the explicit CORS allowlist: the deployed frontend plus
// local Vite dev. A wildcard ("*") can't be paired with credentialed
// requests and can't express "these two specific origins" — echoing back
// the request's Origin only when it matches this list is the standard
// pattern for a small, known set of frontends.
var allowedOrigins = map[string]bool{
	"https://swava.vercel.app": true,
	"http://localhost:5173":    true,
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func handleWallets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}
	writeJSON(w, http.StatusOK, demoWallets)
}

// --- GET /api/agent-profile?address=... ---

type agentProfileResponse struct {
	Address      string                   `json:"address"`
	AgentRole    string                   `json:"agent_role"`
	Known        bool                     `json:"known"`
	RawScore     int                      `json:"raw_score"`
	Tier         string                   `json:"tier"`
	Policy       *risk.ProcurementPolicy  `json:"policy,omitempty"`
	Transactions []reputation.Transaction `json:"transactions"`
}

// handleAgentProfile builds a real profile for one of the demo agents: role
// label, its own wallet identity, its trust score, and its actual indexed
// transaction history — sparse if the wallet's real history is sparse, never
// padded with invented rows.
func handleAgentProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}
	address := r.URL.Query().Get("address")
	if address == "" {
		writeError(w, http.StatusBadRequest, "address is required")
		return
	}

	agentRole := "Procurement Agent"
	for _, w := range demoWallets {
		if w.Address == address {
			agentRole = w.AgentRole
			break
		}
	}

	swarmpayURL := os.Getenv("SWARMPAY_API_URL")
	if swarmpayURL == "" {
		swarmpayURL = "http://localhost:8080"
	}
	repClient := reputation.NewClient(swarmpayURL, os.Getenv("SWARMPAY_API_KEY"))

	score, err := repClient.GetScore(address)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	txs, err := repClient.GetHistory(address)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	var policy *risk.ProcurementPolicy
	if p, ok := risk.PolicyFor(address); ok {
		policy = &p
	}

	writeJSON(w, http.StatusOK, agentProfileResponse{
		Address:      address,
		AgentRole:    agentRole,
		Known:        score.Known,
		RawScore:     score.RawScore,
		Tier:         score.Tier,
		Policy:       policy,
		Transactions: txs,
	})
}

// --- POST /api/decide ---

type decideRequest struct {
	Task         string `json:"task"`
	BudgetRupees int    `json:"budget_rupees"`
}

type decideResponse struct {
	decide.Result
	WalletRole string `json:"wallet_role"`
}

func handleDecide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var req decideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Task == "" {
		writeError(w, http.StatusBadRequest, "task is required")
		return
	}
	if req.BudgetRupees <= 0 {
		req.BudgetRupees = 1000
	}

	result := decide.Run(req.Task, req.BudgetRupees)
	writeJSON(w, http.StatusOK, decideResponse{Result: result, WalletRole: walletRoleNote})
}

// --- POST /api/trust-gate ---

type trustGateRequest struct {
	WalletAddress string `json:"wallet_address"`
	AmountRupees  int    `json:"amount_rupees"`
}

func handleTrustGate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var req trustGateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.WalletAddress == "" {
		writeError(w, http.StatusBadRequest, "wallet_address is required")
		return
	}

	gate, err := purchase.EvaluateTrustGate(req.WalletAddress, req.AmountRupees, "")
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, gate)
}

// --- POST /api/purchase ---

type purchaseRequest struct {
	WalletAddress string           `json:"wallet_address"`
	Product       purchase.Product `json:"product"`
}

type purchaseResponse struct {
	purchase.SessionResult
	WalletRole string `json:"wallet_role"`
	// SandboxNote is set whenever the flow reached Prava (i.e. was not
	// blocked) — the frontend surfaces this as a finding, not a failure.
	SandboxNote string `json:"sandbox_note,omitempty"`
}

const sandboxStopNote = "A real session and card have been issued by Prava. Completing the purchase requires a human to open the session URL in a browser and pass Visa's passkey/biometric verification — there is no server-side or API path around this step. Prava's documented supported completion path for agents, Browser Harness, is production-only (real cards, no sandbox) and has no direct API — it's CLI/hosted-agent only. This is where an automated sandbox run stops."

func handlePurchase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var req purchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.WalletAddress == "" {
		writeError(w, http.StatusBadRequest, "wallet_address is required")
		return
	}
	if req.Product.VariantID == "" {
		writeError(w, http.StatusBadRequest, "product is required")
		return
	}

	session, err := purchase.CreateSandboxSession(req.Product, req.WalletAddress)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	resp := purchaseResponse{SessionResult: session, WalletRole: walletRoleNote}
	if session.TrustGate.Decision != risk.DecisionBlock {
		resp.SandboxNote = sandboxStopNote
	}
	writeJSON(w, http.StatusOK, resp)
}
