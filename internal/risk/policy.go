// Package risk turns a reputation score into a spend limit and decision.
package risk

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
}

// Evaluate applies the SwarmPay trust gate: score -> spend limit -> decision
// for a specific purchase amount (rupees). known distinguishes a wallet with
// no reputation history (never zero, never an error — always NeutralScore)
// from a wallet that genuinely scored low.
//
//	70+            -> 10,000 limit, auto-approve
//	unknown/50     -> 500 limit, auto-approve under it, human review above
//	below 30       -> 0 limit, blocked
//	30-69 (known)  -> 500 limit, auto-approve under it, human review above
func Evaluate(known bool, score int, amount int) Policy {
	effectiveScore := score
	if !known {
		effectiveScore = NeutralScore
	}

	switch {
	case effectiveScore >= 70:
		return decide(effectiveScore, 10000, amount)
	case effectiveScore < 30:
		return Policy{
			Score:      effectiveScore,
			SpendLimit: 0,
			Decision:   DecisionBlock,
			Reason:     "score below 30 — wallet blocked",
		}
	default: // 30-69, including the unknown-wallet neutral default of 50
		return decide(effectiveScore, 500, amount)
	}
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
