// Package ucp is a minimal client for Shopify's UCP-over-MCP merchant endpoints.
package ucp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AgentProfileURL points to a hosted UCP agent profile document. Shopify fetches
// and validates this on every request before answering. Using Shopify's own
// published test fixture for now; swap for a self-hosted profile later.
const AgentProfileURL = "https://shopify.dev/ucp/agent-profiles/2026-04-08/valid-with-capabilities.json"

// Client talks MCP/JSON-RPC to a single merchant's UCP endpoint.
type Client struct {
	Name       string // human-readable merchant label
	Endpoint   string // e.g. https://headphone-zone.myshopify.com/api/ucp/mcp
	HTTPClient *http.Client
}

func NewClient(name, endpoint string) *Client {
	return &Client{
		Name:     name,
		Endpoint: endpoint,
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("mcp error %d: %s", e.Code, e.Message)
}

func agentMeta() map[string]any {
	return map[string]any{
		"ucp-agent": map[string]any{
			"profile": AgentProfileURL,
		},
	}
}

func (c *Client) call(method string, params any) (json.RawMessage, error) {
	req := rpcRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: params}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}
	return rpcResp.Result, nil
}

// Tool describes an MCP tool as returned by tools/list.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ListTools calls tools/list and returns the tools the merchant exposes.
func (c *Client) ListTools() ([]Tool, error) {
	params := map[string]any{
		"arguments": map[string]any{
			"meta": agentMeta(),
		},
	}
	raw, err := c.call("tools/list", params)
	if err != nil {
		return nil, err
	}
	var result struct {
		Tools []Tool `json:"tools"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode tools/list result: %w", err)
	}
	return result.Tools, nil
}

// Money is an amount in minor units (e.g. paise, cents) plus currency code.
type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// Product is the subset of the UCP catalog search response we care about.
type Product struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	PriceRange struct {
		Min Money `json:"min"`
		Max Money `json:"max"`
	} `json:"price_range"`
}

// SearchCatalog calls search_catalog with a free-text query and returns matching products.
func (c *Client) SearchCatalog(query string) ([]Product, error) {
	params := map[string]any{
		"name": "search_catalog",
		"arguments": map[string]any{
			"meta": agentMeta(),
			"catalog": map[string]any{
				"query": query,
			},
		},
	}
	raw, err := c.call("tools/call", params)
	if err != nil {
		return nil, err
	}

	var toolResult struct {
		IsError bool `json:"isError"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StructuredContent struct {
			Products []Product `json:"products"`
		} `json:"structuredContent"`
	}
	if err := json.Unmarshal(raw, &toolResult); err != nil {
		return nil, fmt.Errorf("decode search_catalog result: %w", err)
	}
	if toolResult.IsError {
		msg := "unknown tool error"
		if len(toolResult.Content) > 0 {
			msg = toolResult.Content[0].Text
		}
		return nil, fmt.Errorf("search_catalog failed: %s", msg)
	}
	return toolResult.StructuredContent.Products, nil
}

// ToolResult is the generic shape of an MCP tools/call response: human-readable
// content plus an optional structured payload, and whether the call errored.
type ToolResult struct {
	IsError bool
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StructuredContent json.RawMessage `json:"structuredContent"`
	Raw               json.RawMessage
}

// ErrorText returns the human-readable error message from Content, if any.
func (r *ToolResult) ErrorText() string {
	if len(r.Content) > 0 {
		return r.Content[0].Text
	}
	return "unknown tool error"
}

// CallTool invokes an arbitrary MCP tool by name with the given arguments
// (agent profile meta is injected automatically) and returns the raw tool result.
func (c *Client) CallTool(name string, arguments map[string]any) (*ToolResult, error) {
	args := map[string]any{"meta": agentMeta()}
	for k, v := range arguments {
		args[k] = v
	}
	params := map[string]any{
		"name":      name,
		"arguments": args,
	}
	raw, err := c.call("tools/call", params)
	if err != nil {
		return nil, err
	}

	var result ToolResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode %s result: %w", name, err)
	}
	result.Raw = raw
	return &result, nil
}

// CreateCheckout creates a checkout from line items and a shipping destination, and
// returns the raw tool result; the caller decides how to interpret structuredContent
// (checkout id, totals, errors).
func (c *Client) CreateCheckout(lineItems []map[string]any, destination map[string]any) (*ToolResult, error) {
	checkout := map[string]any{
		"line_items": lineItems,
	}
	if destination != nil {
		checkout["fulfillment"] = map[string]any{
			"methods": []map[string]any{
				{
					"type":         "shipping",
					"destinations": []map[string]any{destination},
				},
			},
		}
	}
	return c.CallTool("create_checkout", map[string]any{
		"checkout": checkout,
	})
}

// CompleteCheckout submits a payment instrument against an existing checkout and
// returns the raw tool result; the caller decides success/failure from the response.
func (c *Client) CompleteCheckout(checkoutID, idempotencyKey string, instruments []map[string]any) (*ToolResult, error) {
	args := map[string]any{
		"id": checkoutID,
		"checkout": map[string]any{
			"payment": map[string]any{
				"instruments": instruments,
			},
		},
	}
	// idempotency-key lives under meta alongside ucp-agent, so bypass CallTool's
	// generic meta injection and build the full argument set here instead.
	metaWithIdempotency := map[string]any{
		"ucp-agent":       agentMeta()["ucp-agent"],
		"idempotency-key": idempotencyKey,
	}
	args["meta"] = metaWithIdempotency

	params := map[string]any{
		"name":      "complete_checkout",
		"arguments": args,
	}
	raw, err := c.call("tools/call", params)
	if err != nil {
		return nil, err
	}
	var result ToolResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode complete_checkout result: %w", err)
	}
	result.Raw = raw
	return &result, nil
}
