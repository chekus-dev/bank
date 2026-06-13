package paystack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"transfer-app/config"
)

type Client struct {
	SecretKey  string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient() *Client {
	return &Client{
		SecretKey: config.App.PaystackSecretKey,
		BaseURL:   config.App.PaystackBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("paystack error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ── Initialize Transaction ──────────────────────────────────────────────────

type InitializeRequest struct {
	Email     string                 `json:"email"`
	Amount    int64                  `json:"amount"` // in kobo
	Reference string                 `json:"reference"`
	Callback  string                 `json:"callback_url,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Channels  []string               `json:"channels,omitempty"`
}

type InitializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

func (c *Client) InitializeTransaction(req InitializeRequest) (*InitializeResponse, error) {
	body, err := c.doRequest("POST", "/transaction/initialize", req)
	if err != nil {
		return nil, err
	}
	var result InitializeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal initialize response: %w", err)
	}
	return &result, nil
}

// ── Verify Transaction ──────────────────────────────────────────────────────

type VerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID              int64  `json:"id"`
		Status          string `json:"status"`
		Reference       string `json:"reference"`
		Amount          int64  `json:"amount"`
		GatewayResponse string `json:"gateway_response"`
		PaidAt          string `json:"paid_at"`
		Channel         string `json:"channel"`
		Currency        string `json:"currency"`
		Customer        struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
		} `json:"customer"`
	} `json:"data"`
}

func (c *Client) VerifyTransaction(reference string) (*VerifyResponse, error) {
	body, err := c.doRequest("GET", "/transaction/verify/"+reference, nil)
	if err != nil {
		return nil, err
	}
	var result VerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal verify response: %w", err)
	}
	return &result, nil
}

// ── Resolve Account Number ──────────────────────────────────────────────────

type ResolveAccountResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AccountNumber string `json:"account_number"`
		AccountName   string `json:"account_name"`
		BankID        int    `json:"bank_id"`
	} `json:"data"`
}

func (c *Client) ResolveAccountNumber(accountNumber, bankCode string) (*ResolveAccountResponse, error) {
	endpoint := fmt.Sprintf("/bank/resolve?account_number=%s&bank_code=%s", accountNumber, bankCode)
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result ResolveAccountResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal resolve response: %w", err)
	}
	return &result, nil
}

// ── List Banks ──────────────────────────────────────────────────────────────

type Bank struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	LongCode string `json:"longcode"`
	Country  string `json:"country"`
	Currency string `json:"currency"`
}

type ListBanksResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    []Bank `json:"data"`
}

func (c *Client) ListBanks() (*ListBanksResponse, error) {
	body, err := c.doRequest("GET", "/bank?country=nigeria&use_cursor=false&perPage=100", nil)
	if err != nil {
		return nil, err
	}
	var result ListBanksResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal banks response: %w", err)
	}
	return &result, nil
}

// ── Create Transfer Recipient ───────────────────────────────────────────────

type CreateRecipientRequest struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	Currency      string `json:"currency"`
}

type CreateRecipientResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		RecipientCode string `json:"recipient_code"`
		Type          string `json:"type"`
		Name          string `json:"name"`
		AccountNumber string `json:"account_number"`
		BankCode      string `json:"bank_code"`
	} `json:"data"`
}

func (c *Client) CreateTransferRecipient(req CreateRecipientRequest) (*CreateRecipientResponse, error) {
	if req.Currency == "" {
		req.Currency = "NGN"
	}
	if req.Type == "" {
		req.Type = "nuban"
	}
	body, err := c.doRequest("POST", "/transferrecipient", req)
	if err != nil {
		return nil, err
	}
	var result CreateRecipientResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal recipient response: %w", err)
	}
	return &result, nil
}

// ── Initiate Transfer ───────────────────────────────────────────────────────

type InitiateTransferRequest struct {
	Source    string `json:"source"`
	Amount    int64  `json:"amount"` // in kobo
	Recipient string `json:"recipient"`
	Reason    string `json:"reason"`
	Reference string `json:"reference"`
}

type InitiateTransferResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		TransferCode string `json:"transfer_code"`
		Status       string `json:"status"`
		Amount       int64  `json:"amount"`
		Reference    string `json:"reference"`
	} `json:"data"`
}

func (c *Client) InitiateTransfer(req InitiateTransferRequest) (*InitiateTransferResponse, error) {
	if req.Source == "" {
		req.Source = "balance"
	}
	body, err := c.doRequest("POST", "/transfer", req)
	if err != nil {
		return nil, err
	}
	var result InitiateTransferResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal transfer response: %w", err)
	}
	return &result, nil
}
