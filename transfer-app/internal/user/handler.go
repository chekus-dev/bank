package user

import (
	"encoding/json"
	"net/http"
	"transfer-app/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	u, err := h.service.GetByID(userID)
	if err != nil {
		response.NotFound(w, "user not found")
		return
	}
	response.Success(w, "profile fetched", ToResponse(u))
}

func (h *Handler) SetPin(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req UpdatePinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if len(req.Pin) != 4 && len(req.Pin) != 6 {
		response.BadRequest(w, "pin must be 4 or 6 digits")
		return
	}
	if req.Pin != req.ConfirmPin {
		response.BadRequest(w, "pins do not match")
		return
	}
	if err := h.service.UpdatePin(userID, req.Pin); err != nil {
		response.InternalServerError(w, "could not set pin")
		return
	}
	response.Success(w, "transaction pin set successfully", nil)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if len(req.NewPassword) < 8 {
		response.BadRequest(w, "new password must be at least 8 characters")
		return
	}
	if err := h.service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "password changed successfully", nil)
}
