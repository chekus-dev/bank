package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"transfer-app/pkg/response"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Unauthorized(w, "authorization header required")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(w, "invalid authorization format")
			return
		}
		userID, err := ValidateAccessToken(parts[1])
		if err != nil {
			response.Unauthorized(w, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateRegister(req RegisterRequest) error {
	if req.FirstName == "" || req.LastName == "" {
		return fmt.Errorf("first and last name are required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if len(req.Phone) < 11 {
		return fmt.Errorf("valid Nigerian phone number required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(req.BVN) != 11 {
		return fmt.Errorf("BVN must be 11 digits")
	}
	return nil
}
