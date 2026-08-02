// Package risk turns a reputation score into a spend limit and decision.
package risk

import (
	"fmt"
	"strings"
)

// Decision is what the policy decided for a purchase amount against a limit.
type Decision string

const (
	DecisionApprove     Decision = "auto_approve"
	DecisionHumanReview Decision = "human_review"
	DecisionBlock       Decision = "blocked"
)

// NeutralScore is applied to a wallet with no history — deliberately distinct
// from a real low score. An unknown wallet is unproven, not untrusted.
const NeutralScore = 50

// Policy is the outcome of scoring a wallet: the spend limit it's allowed and
// whether a purchase of a given amount is auto-approved, needs a human, or is
// blocked outright.
type Policy struct {
	Score      int // 0-100 normalized score actually used (NeutralScore if unknown)
	SpendLimit int // rupees
	Decision   Decision
	Reason     string
	Policy     *ProcurementPolicy `json:"policy,omitempty"` // the constraint actually applied, if any
}

// ProcurementPolicy is a real, per-agent purchasing constraint — a company's
// own rule about what an agent is allowed to buy, independent of its trust
// score. Checked before the score-derived limit; a category block or a
// policy cap tighter than the score would allow always wins, since a
// procurement rule is a hard company decision, not a trust signal to weigh.
type ProcurementPolicy struct {
	AgentLabel      string   `json:"agent_label"`
	CategoryCap     int      `json:"category_cap_rupees"`        // rupees; 0 means no policy cap configured
	BlockedKeywords []string `json:"blocked_keywords,omitempty"` // case-insensitive substrings in product description that are never allowed
}

// procurementPolicies is the one configured policy per demo agent — a real
// constraint the decision engine checks, not just a label. Keyed by wallet
// address so EvaluatePolicy can look it up the same way the reputation
// client looks up a score.
var procurementPolicies = map[string]ProcurementPolicy{
	"0xca4b519063ff7f8154fcb768259bc9059df07237": {
		AgentLabel:      "Procurement Agent — Alpha",
		CategoryCap:     2000,
		BlockedKeywords: []string{"headphone"},
	},
}

// PolicyFor returns the configured procurement policy for a wallet, if any.
func PolicyFor(walletAddress string) (ProcurementPolicy, bool) {
	p, ok := procurementPolicies[walletAddress]
	return p, ok
}

// Evaluate applies the SwarmPay trust gate: score -> spend limit -> decision
// for a specific purchase amount (rupees), then applies any configured
// procurement policy for that wallet as a hard constraint on top. known
// distinguishes a wallet with no reputation history (never zero, never an
// error — always NeutralScore) from a wallet that genuinely scored low.
//
//	70+            -> 10,000 limit, auto-approve
//	unknown/50     -> 500 limit, auto-approve under it, human review above
//	below 30       -> 0 limit, blocked
//	30-69 (known)  -> 500 limit, auto-approve under it, human review above
//
// productDescription is used only to check blocked-keyword procurement
// rules; pass "" when no policy check against product content is needed.
func Evaluate(known bool, score int, amount int, walletAddress, productDescription string) Policy {
	effectiveScore := score
	if !known {
		effectiveScore = NeutralScore
	}

	var base Policy
	switch {
	case effectiveScore >= 70:
		base = decide(effectiveScore, 10000, amount)
	case effectiveScore < 30:
		base = Policy{
			Score:      effectiveScore,
			SpendLimit: 0,
			Decision:   DecisionBlock,
			Reason:     "score below 30 — wallet blocked",
		}
	default: // 30-69, including the unknown-wallet neutral default of 50
		base = decide(effectiveScore, 500, amount)
	}

	policy, hasPolicy := PolicyFor(walletAddress)
	if !hasPolicy || base.Decision == DecisionBlock {
		return base
	}
	base.Policy = &policy

	for _, kw := range policy.BlockedKeywords {
		if containsFold(productDescription, kw) {
			base.Decision = DecisionBlock
			base.SpendLimit = 0
			base.Reason = fmt.Sprintf("procurement policy for %s blocks purchases matching %q", policy.AgentLabel, kw)
			return base
		}
	}

	if policy.CategoryCap > 0 && amount > policy.CategoryCap {
		base.Decision = DecisionHumanReview
		base.Reason = fmt.Sprintf("amount exceeds %s's procurement policy cap of Rs.%d — needs manager approval", policy.AgentLabel, policy.CategoryCap)
		if policy.CategoryCap < base.SpendLimit {
			base.SpendLimit = policy.CategoryCap
		}
	}

	return base
}

func containsFold(s, substr string) bool {
	return len(substr) > 0 && strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func decide(score, limit, amount int) Policy {
	if amount <= limit {
		return Policy{
			Score:      score,
			SpendLimit: limit,
			Decision:   DecisionApprove,
			Reason:     "amount within spend limit",
		}
	}
	return Policy{
		Score:      score,
		SpendLimit: limit,
		Decision:   DecisionHumanReview,
		Reason:     "amount exceeds spend limit — needs human review",
	}
}
