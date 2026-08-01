// Package reputation is a minimal client for the SwarmPay reputation API.
package reputation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client talks to a single SwarmPay reputation-api deployment.
type Client struct {
	BaseURL    string // e.g. http://localhost:8080
	APIKey     string // sent as a Bearer token when set; the API may not require it yet
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// scoreResponse mirrors reputation-api's GET /v1/score/{address} wire format.
type scoreResponse struct {
	Address    string          `json:"address"`
	Known      bool            `json:"known"`
	Score      int             `json:"score"`
	Tier       string          `json:"tier"`
	Components json.RawMessage `json:"components,omitempty"`
}

// Score is a wallet's reputation as reported by SwarmPay. Known distinguishes
// a wallet that has never been indexed (Known: false — genuinely no history)
// from one that has a real, computed score, including a real zero.
// RawScore is on SwarmPay's native 0-1000ish scale; risk.ScoreToLimit expects
// the 0-100 normalization (RawScore / 10) done by ToNormalized.
type Score struct {
	Address  string
	Known    bool
	RawScore int
	Tier     string
}

// ToNormalized maps SwarmPay's native score to the 0-100 scale internal/risk
// operates on. Observed live data tops out around 600/1000 (no wallet has
// reached the AAA band yet), so this is a linear /10, not a re-bucketing.
func (s Score) ToNormalized() int {
	return s.RawScore / 10
}

// GetScore fetches a wallet's reputation. A wallet with no indexed history
// returns Score{Known: false} with no error — that is the expected, valid
// response for a new/unseen wallet, never a failure.
func (c *Client) GetScore(address string) (Score, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/v1/score/"+address, nil)
	if err != nil {
		return Score{}, fmt.Errorf("build request: %w", err)
	}
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Score{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Score{}, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return Score{}, fmt.Errorf("get score failed: status %d: %s", resp.StatusCode, body)
	}

	var out scoreResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return Score{}, fmt.Errorf("decode score response: %w", err)
	}

	return Score{
		Address:  out.Address,
		Known:    out.Known,
		RawScore: out.Score,
		Tier:     out.Tier,
	}, nil
}
