// Package payments is a minimal client for the Prava sandbox payment API.
package payments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client talks to a single Prava environment (sandbox or production).
type Client struct {
	BaseURL    string // e.g. https://sandbox.api.prava.space
	SecretKey  string // sk_test_... or sk_live_...
	HTTPClient *http.Client
}

func NewClient(baseURL, secretKey string) *Client {
	return &Client{
		BaseURL:   baseURL,
		SecretKey: secretKey,
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// Result carries a raw API response plus the X-Response-ID header Prava support
// uses to trace a request server-side (required for any infra-side error report).
type Result struct {
	Body       []byte
	StatusCode int
	ResponseID string
}

func (c *Client) do(method, path string, body any) (Result, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return Result{}, fmt.Errorf("encode request: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return Result{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.SecretKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("read response: %w", err)
	}
	return Result{
		Body:       respBody,
		StatusCode: resp.StatusCode,
		ResponseID: resp.Header.Get("X-Response-ID"),
	}, nil
}

// --- Create Session ---

type MerchantDetails struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	CountryCodeISO2 string `json:"country_code_iso2"`
}

type ProductDetails struct {
	Description string `json:"description"`
	UnitPrice   string `json:"unit_price"`
	Quantity    int    `json:"quantity,omitempty"`
}

type PurchaseContext struct {
	MerchantDetails MerchantDetails  `json:"merchant_details"`
	ProductDetails  []ProductDetails `json:"product_details"`
}

type CreateSessionRequest struct {
	UserID          string            `json:"user_id"`
	UserEmail       string            `json:"user_email"`
	TotalAmount     string            `json:"total_amount"`
	Currency        string            `json:"currency"`
	PurchaseContext []PurchaseContext `json:"purchase_context"`
	IntegrationType string            `json:"integration_type,omitempty"`
	Description     string            `json:"description,omitempty"`
}

type CreateSessionResponse struct {
	SessionID    string `json:"session_id"`
	SessionToken string `json:"session_token"`
	IframeURL    string `json:"iframe_url"`
	OrderID      string `json:"order_id"`
	ExpiresAt    string `json:"expires_at"`
}

// CreateSession pins the order details and returns a session the cardholder must approve.
func (c *Client) CreateSession(req CreateSessionRequest) (*CreateSessionResponse, Result, error) {
	res, err := c.do(http.MethodPost, "/v1/sessions", req)
	if err != nil {
		return nil, res, err
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, res, fmt.Errorf("create session failed: status %d: %s", res.StatusCode, res.Body)
	}
	var out CreateSessionResponse
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, res, fmt.Errorf("decode create-session response: %w", err)
	}
	return &out, res, nil
}

// --- Get Payment Result ---

type LineItem struct {
	TxnRefID     string `json:"txn_ref_id"`
	MerchantName string `json:"merchant_name"`
	TotalAmount  string `json:"total_amount"`
	Status       string `json:"status"`
	Token        string `json:"token"`
	DynamicCVV   string `json:"dynamic_cvv"`
	ExpiryMonth  string `json:"expiry_month"`
	ExpiryYear   string `json:"expiry_year"`
}

type Transaction struct {
	TxnID     string     `json:"txn_id"`
	Status    string     `json:"status"`
	LineItems []LineItem `json:"line_items"`
}

type PaymentResultResponse struct {
	SessionID    string        `json:"session_id"`
	OrderID      string        `json:"order_id"`
	Status       string        `json:"status"` // pending, awaiting_result, completed, failed
	Transactions []Transaction `json:"transactions"`
}

// GetPaymentResult polls the current state of a session's payment.
func (c *Client) GetPaymentResult(sessionID string) (*PaymentResultResponse, Result, error) {
	res, err := c.do(http.MethodGet, "/v1/sessions/"+sessionID+"/payment-result", nil)
	if err != nil {
		return nil, res, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, res, fmt.Errorf("get payment result failed: status %d: %s", res.StatusCode, res.Body)
	}
	var out PaymentResultResponse
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, res, fmt.Errorf("decode payment-result response: %w", err)
	}
	return &out, res, nil
}

// --- Report Status ---

type ReportStatusRequest struct {
	TxnRefID          string `json:"txn_ref_id"`
	TxnStatus         string `json:"txn_status"` // APPROVED or DECLINED
	AuthorizationCode string `json:"authorization_code,omitempty"`
	ResponseCode      string `json:"response_code,omitempty"`
}

type ReportStatusResponse struct {
	Status           string `json:"status"`
	TxnRefID         string `json:"txn_ref_id"`
	TxnStatus        string `json:"txn_status"`
	VisaConfirmation string `json:"visa_confirmation"`
}

// ReportStatus closes the loop by telling Prava whether the downstream charge succeeded.
func (c *Client) ReportStatus(sessionID string, req ReportStatusRequest) (*ReportStatusResponse, Result, error) {
	res, err := c.do(http.MethodPost, "/v1/sessions/"+sessionID+"/report-status", req)
	if err != nil {
		return nil, res, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, res, fmt.Errorf("report-status failed: status %d: %s", res.StatusCode, res.Body)
	}
	var out ReportStatusResponse
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, res, fmt.Errorf("decode report-status response: %w", err)
	}
	return &out, res, nil
}
