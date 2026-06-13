package webhook

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"transfer-app/config"
	"transfer-app/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Paystack(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.BadRequest(w, "could not read body")
		return
	}

	// Verify signature
	sig := r.Header.Get("X-Paystack-Signature")
	if !verifyPaystackSignature(body, sig, config.App.PaystackWebhookSecret) {
		log.Printf("❌ invalid paystack webhook signature")
		response.Unauthorized(w, "invalid signature")
		return
	}

	var event PaystackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		response.BadRequest(w, "invalid event payload")
		return
	}

	// Respond immediately, process async
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))

	go func() {
		if err := h.service.HandlePaystackEvent(event); err != nil {
			log.Printf("❌ webhook processing error: %v", err)
		}
	}()
}

func verifyPaystackSignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
