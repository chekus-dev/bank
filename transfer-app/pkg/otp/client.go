package otp

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
	APIKey     string
	BaseURL    string
	SenderID   string
	HTTPClient *http.Client
}

func NewClient() *Client {
	return &Client{
		APIKey:   config.App.TermiiAPIKey,
		BaseURL:  config.App.TermiiBaseURL,
		SenderID: config.App.TermiiSenderID,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ── Send OTP ────────────────────────────────────────────────────────────────

type SendOTPRequest struct {
	APIKey      string `json:"api_key"`
	MessageType string `json:"message_type"`
	To          string `json:"to"`
	From        string `json:"from"`
	Channel     string `json:"channel"`
	PinType     string `json:"pin_type"`
	PinAttempts int    `json:"pin_attempts"`
	PinTimeToLive int  `json:"pin_time_to_live"`
	PinLength   int    `json:"pin_length"`
	PinPlaceholder string `json:"pin_placeholder"`
	MessageText string `json:"message_text"`
}

type SendOTPResponse struct {
	PinID   string `json:"pinId"`
	To      string `json:"to"`
	SmsStatus string `json:"smsStatus"`
}

func (c *Client) SendOTP(phone string) (*SendOTPResponse, error) {
	reqBody := SendOTPRequest{
		APIKey:         c.APIKey,
		MessageType:    "NUMERIC",
		To:             phone,
		From:           c.SenderID,
		Channel:        "dnd",
		PinType:        "NUMERIC",
		PinAttempts:    3,
		PinTimeToLive:  config.App.OTPExpiryMins,
		PinLength:      6,
		PinPlaceholder: "< 1234 >",
		MessageText:    "Your NairaTransfer verification code is < 1234 >. Valid for " + fmt.Sprintf("%d", config.App.OTPExpiryMins) + " minutes. Do not share.",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal otp request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/sms/otp/send", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create otp request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send otp: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read otp response: %w", err)
	}

	var result SendOTPResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal otp response: %w", err)
	}
	return &result, nil
}

// ── Verify OTP ──────────────────────────────────────────────────────────────

type VerifyOTPRequest struct {
	APIKey string `json:"api_key"`
	PinID  string `json:"pin_id"`
	Pin    string `json:"pin"`
}

type VerifyOTPResponse struct {
	PinID    string `json:"pinId"`
	Verified string `json:"verified"`
	Msisdn   string `json:"msisdn"`
}

func (c *Client) VerifyOTP(pinID, pin string) (*VerifyOTPResponse, error) {
	reqBody := VerifyOTPRequest{
		APIKey: c.APIKey,
		PinID:  pinID,
		Pin:    pin,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal verify request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/sms/otp/verify", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create verify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("verify otp: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read verify response: %w", err)
	}

	var result VerifyOTPResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal verify response: %w", err)
	}

	if result.Verified != "True" {
		return nil, fmt.Errorf("invalid or expired OTP")
	}

	return &result, nil
}

// ── Send Plain SMS ──────────────────────────────────────────────────────────

type SendSMSRequest struct {
	To      string `json:"to"`
	From    string `json:"from"`
	SMS     string `json:"sms"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	APIKey  string `json:"api_key"`
}

func (c *Client) SendSMS(phone, message string) error {
	reqBody := SendSMSRequest{
		To:      phone,
		From:    c.SenderID,
		SMS:     message,
		Type:    "plain",
		Channel: "dnd",
		APIKey:  c.APIKey,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal sms request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/sms/send", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("create sms request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send sms: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sms error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
