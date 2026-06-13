package user

import (
	"database/sql"
	"fmt"
	"transfer-app/internal/db"

	"golang.org/x/crypto/bcrypt"
)

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) GetByID(userID string) (*User, error) {
	u := &User{}
	err := db.DB.QueryRow(db.GetUserByIDQuery, userID).Scan(
		&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Phone,
		&u.PasswordHash, &u.PinHash, &u.BVN, &u.IsVerified, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (s *Service) GetByEmail(email string) (*User, error) {
	u := &User{}
	err := db.DB.QueryRow(db.GetUserByEmailQuery, email).Scan(
		&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Phone,
		&u.PasswordHash, &u.PinHash, &u.BVN, &u.IsVerified, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (s *Service) GetByPhone(phone string) (*User, error) {
	u := &User{}
	err := db.DB.QueryRow(db.GetUserByPhoneQuery, phone).Scan(
		&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Phone,
		&u.PasswordHash, &u.PinHash, &u.BVN, &u.IsVerified, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (s *Service) UpdatePin(userID, pin string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash pin: %w", err)
	}
	_, err = db.DB.Exec(db.UpdateUserPinQuery, string(hash), userID)
	return err
}

func (s *Service) VerifyPin(user *User, pin string) bool {
	if user.PinHash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.PinHash), []byte(pin)) == nil
}

func (s *Service) ChangePassword(userID, oldPassword, newPassword string) error {
	u, err := s.GetByID(userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)) != nil {
		return fmt.Errorf("incorrect current password")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = db.DB.Exec(db.UpdateUserPasswordQuery, string(hash), userID)
	return err
}

func ToResponse(u *User) UserResponse {
	return UserResponse{
		ID:         u.ID,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Email:      u.Email,
		Phone:      u.Phone,
		IsVerified: u.IsVerified,
		CreatedAt:  u.CreatedAt,
	}
}
