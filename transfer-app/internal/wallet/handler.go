package wallet

import (
	"encoding/json"
	"net/http"
	"transfer-app/pkg/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	wallet, err := h.service.GetWallet(userID)
	if err != nil {
		response.NotFound(w, "wallet not found")
		return
	}
	response.Success(w, "wallet fetched", wallet)
}

func (h *Handler) FundWallet(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req FundWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Amount <= 0 {
		response.BadRequest(w, "amount must be greater than 0")
		return
	}
	if req.Email == "" {
		response.BadRequest(w, "email is required")
		return
	}
	res, err := h.service.InitiateFunding(userID, req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "payment link generated", res)
}

func (h *Handler) ListBanks(w http.ResponseWriter, r *http.Request) {
	banks, err := h.service.ListBanks()
	if err != nil {
		response.InternalServerError(w, "could not fetch banks")
		return
	}
	response.Success(w, "banks fetched", banks)
}

func (h *Handler) ResolveAccount(w http.ResponseWriter, r *http.Request) {
	var req ResolveAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.AccountNumber == "" || req.BankCode == "" {
		response.BadRequest(w, "account_number and bank_code are required")
		return
	}
	res, err := h.service.ResolveAccount(req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "account resolved", res)
}

func (h *Handler) AddBeneficiary(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req AddBeneficiaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	ben, err := h.service.AddBeneficiary(userID, req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Created(w, "beneficiary added", ben)
}

func (h *Handler) ListBeneficiaries(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	list, err := h.service.ListBeneficiaries(userID)
	if err != nil {
		response.InternalServerError(w, "could not fetch beneficiaries")
		return
	}
	response.Success(w, "beneficiaries fetched", list)
}

func (h *Handler) DeleteBeneficiary(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	benID := chi.URLParam(r, "id")
	if err := h.service.DeleteBeneficiary(userID, benID); err != nil {
		response.NotFound(w, "beneficiary not found")
		return
	}
	response.Success(w, "beneficiary deleted", nil)
}
