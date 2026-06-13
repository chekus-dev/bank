package transfer

import (
	"encoding/json"
	"net/http"
	"strconv"
	"transfer-app/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) P2PTransfer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req P2PTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.ReceiverPhone == "" || req.Amount <= 0 || req.Pin == "" {
		response.BadRequest(w, "receiver_phone, amount, and pin are required")
		return
	}
	res, err := h.service.P2PTransfer(userID, req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Created(w, "transfer successful", res)
}

func (h *Handler) BankTransfer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req BankTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.AccountNumber == "" || req.BankCode == "" || req.Amount <= 0 || req.Pin == "" {
		response.BadRequest(w, "account_number, bank_code, amount, and pin are required")
		return
	}
	res, err := h.service.BankTransfer(userID, req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Created(w, "bank transfer initiated", res)
}

func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	txns, total, err := h.service.GetTransactions(userID, page, perPage)
	if err != nil {
		response.InternalServerError(w, "could not fetch transactions")
		return
	}
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}
	response.WithPagination(w, "transactions fetched", txns, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	ref := r.URL.Query().Get("reference")
	if ref == "" {
		response.BadRequest(w, "reference is required")
		return
	}
	txn, err := h.service.GetTransaction(ref)
	if err != nil {
		response.NotFound(w, "transaction not found")
		return
	}
	response.Success(w, "transaction fetched", txn)
}
