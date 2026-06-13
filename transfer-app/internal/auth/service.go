package auth

import (
	"database/sql"
	"fmt"
	"time"
	"transfer-app/config"
	"transfer-app/internal/db"
	"transfer-app/pkg/otp"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
	BVN       string `json:"bvn"`
}

type RegisterResponse struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

type SendOTPRequest struct {
	Phone string `json:"phone"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	IsVerified   bool   `json:"is_verified"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ResetPasswordRequest struct {
	Phone       string `json:"phone"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

type Service struct {
	otpClient *otp.Client
}

func NewService(otpClient *otp.Client) *Service {
	return &Service{otpClient: otpClient}
}

func (s *Service) Register(req RegisterRequest) (*RegisterResponse, error) {
	var exists bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, req.Email).Scan(&exists)
	if exists {
		return nil, fmt.Errorf("email already registered")
	}
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE phone=$1)`, req.Phone).Scan(&exists)
	if exists {
		return nil, fmt.Errorf("phone already registered")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	userID := uuid.New().String()
	_, err = db.DB.Exec(
		`INSERT INTO users (id, first_name, last_name, email, phone, password_hash, pin_hash, bvn, is_verified, is_active, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,'',  $7, false, true, NOW(), NOW())`,
		userID, req.FirstName, req.LastName, req.Email, req.Phone, string(hash), req.BVN,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	walletID := uuid.New().String()
	db.DB.Exec(db.CreateWalletQuery, walletID, userID)
	_ = s.SendOTP(req.Phone)
	return &RegisterResponse{UserID: userID, Message: "verify your phone to activate account"}, nil
}

func (s *Service) SendOTP(phone string) error {
	res, err := s.otpClient.SendOTP(phone)
	if err != nil {
		return err
	}
	expiry := time.Now().Add(time.Duration(config.App.OTPExpiryMins) * time.Minute)
	_, err = db.DB.Exec(db.UpsertOTPQuery, phone, res.PinID, expiry)
	return err
}

func (s *Service) VerifyOTP(phone, pin string) error {
	var pinID string
	var expiresAt time.Time
	err := db.DB.QueryRow(db.GetOTPQuery, phone).Scan(&pinID, &expiresAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("OTP not found")
	}
	if err != nil {
		return err
	}
	if time.Now().After(expiresAt) {
		return fmt.Errorf("OTP expired")
	}
	if _, err := s.otpClient.VerifyOTP(pinID, pin); err != nil {
		return err
	}
	db.DB.Exec(db.DeleteOTPQuery, phone)
	db.DB.Exec(`UPDATE users SET is_verified=true, updated_at=NOW() WHERE phone=$1`, phone)
	return nil
}

func (s *Service) Login(req LoginRequest) (*LoginResponse, error) {
	var id, passwordHash string
	var isVerified, isActive bool
	err := db.DB.QueryRow(
		`SELECT id, password_hash, is_verified, is_active FROM users WHERE email=$1`, req.Email,
	).Scan(&id, &passwordHash, &isVerified, &isActive)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !isActive {
		return nil, fmt.Errorf("account deactivated")
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	accessToken, err := generateAccessToken(id)
	if err != nil {
		return nil, err
	}
	refreshToken, err := generateRefreshToken(id)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    config.App.JWTExpiryHours * 3600,
		IsVerified:   isVerified,
	}, nil
}

func (s *Service) RefreshToken(tokenStr string) (*LoginResponse, error) {
	claims, err := parseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	userID := claims["user_id"].(string)
	accessToken, _ := generateAccessToken(userID)
	refreshToken, _ := generateRefreshToken(userID)
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    config.App.JWTExpiryHours * 3600,
	}, nil
}

func (s *Service) ResetPassword(req ResetPasswordRequest) error {
	if err := s.VerifyOTP(req.Phone, req.OTP); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.DB.Exec(`UPDATE users SET password_hash=$1, updated_at=NOW() WHERE phone=$2`, string(hash), req.Phone)
	return err
}

func generateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"type":    "access",
		"exp":     time.Now().Add(time.Duration(config.App.JWTExpiryHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.App.JWTSecret))
}

func generateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"exp":     time.Now().Add(time.Duration(config.App.RefreshTokenExpiry) * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.App.JWTSecret))
}

func parseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(config.App.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func ValidateAccessToken(tokenStr string) (string, error) {
	claims, err := parseToken(tokenStr)
	if err != nil {
		return "", err
	}
	if claims["type"] != "access" {
		return "", fmt.Errorf("not an access token")
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user_id in token")
	}
	return userID, nil
}
