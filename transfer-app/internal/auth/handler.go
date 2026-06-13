package auth

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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validateRegister(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	res, err := h.service.Register(req)
	if err != nil {
		if err.Error() == "email already registered" || err.Error() == "phone already registered" {
			response.Conflict(w, err.Error())
			return
		}
		response.InternalServerError(w, "registration failed")
		return
	}
	response.Created(w, "registration successful, please verify your phone number", res)
}

func (h *Handler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req SendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Phone == "" {
		response.BadRequest(w, "phone number is required")
		return
	}
	if err := h.service.SendOTP(req.Phone); err != nil {
		response.InternalServerError(w, "could not send OTP")
		return
	}
	response.Success(w, "OTP sent successfully", nil)
}

func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Phone == "" || req.OTP == "" {
		response.BadRequest(w, "phone and otp are required")
		return
	}
	if err := h.service.VerifyOTP(req.Phone, req.OTP); err != nil {
		response.BadRequest(w, "invalid or expired OTP")
		return
	}
	response.Success(w, "phone verified successfully", nil)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		response.BadRequest(w, "email and password are required")
		return
	}
	res, err := h.service.Login(req)
	if err != nil {
		response.Unauthorized(w, "invalid credentials")
		return
	}
	response.Success(w, "login successful", res)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.RefreshToken == "" {
		response.BadRequest(w, "refresh token is required")
		return
	}
	res, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(w, "invalid or expired refresh token")
		return
	}
	response.Success(w, "token refreshed", res)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct{ Phone string `json:"phone"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	// Always return success to prevent enumeration
	_ = h.service.SendOTP(req.Phone)
	response.Success(w, "if your phone is registered, an OTP has been sent", nil)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Phone == "" || req.OTP == "" || req.NewPassword == "" {
		response.BadRequest(w, "phone, otp, and new_password are required")
		return
	}
	if err := h.service.ResetPassword(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "password reset successful", nil)
}
