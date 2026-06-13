package wallet

import (
	"database/sql"
	"fmt"
	"transfer-app/internal/db"
	"transfer-app/pkg/paystack"

	"github.com/google/uuid"
)

type Service struct {
	paystack *paystack.Client
}

func NewService(ps *paystack.Client) *Service {
	return &Service{paystack: ps}
}

func (s *Service) GetWallet(userID string) (*Wallet, error) {
	w := &Wallet{}
	err := db.DB.QueryRow(db.GetWalletByUserIDQuery, userID).Scan(
		&w.ID, &w.UserID, &w.Balance, &w.LedgerBalance, &w.Currency, &w.IsActive, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("wallet not found")
	}
	return w, err
}

func (s *Service) GetWalletByID(walletID string) (*Wallet, error) {
	w := &Wallet{}
	err := db.DB.QueryRow(db.GetWalletByIDQuery, walletID).Scan(
		&w.ID, &w.UserID, &w.Balance, &w.LedgerBalance, &w.Currency, &w.IsActive, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("wallet not found")
	}
	return w, err
}

func (s *Service) InitiateFunding(userID string, req FundWalletRequest) (*FundWalletResponse, error) {
	if req.Amount < 100 {
		return nil, fmt.Errorf("minimum funding amount is ₦100")
	}
	ref := "FUND-" + uuid.New().String()
	amountKobo := int64(req.Amount * 100)

	res, err := s.paystack.InitializeTransaction(paystack.InitializeRequest{
		Email:     req.Email,
		Amount:    amountKobo,
		Reference: ref,
		Callback:  req.Callback,
		Metadata: map[string]interface{}{
			"user_id": userID,
			"type":    "wallet_funding",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initiate payment: %w", err)
	}

	return &FundWalletResponse{
		AuthorizationURL: res.Data.AuthorizationURL,
		Reference:        ref,
		Amount:           req.Amount,
	}, nil
}

func (s *Service) CreditWallet(userID string, amount float64, tx *sql.Tx) error {
	w, err := s.GetWallet(userID)
	if err != nil {
		return err
	}
	amountKobo := int64(amount * 100)
	if tx != nil {
		_, err = tx.Exec(db.CreditWalletQuery, amountKobo, w.ID)
	} else {
		_, err = db.DB.Exec(db.CreditWalletQuery, amountKobo, w.ID)
	}
	return err
}

func (s *Service) DebitWallet(walletID string, amount float64, tx *sql.Tx) error {
	amountKobo := int64(amount * 100)
	var newBalance float64
	var err error
	if tx != nil {
		err = tx.QueryRow(db.DebitWalletQuery, amountKobo, walletID).Scan(&newBalance)
	} else {
		err = db.DB.QueryRow(db.DebitWalletQuery, amountKobo, walletID).Scan(&newBalance)
	}
	if err == sql.ErrNoRows {
		return fmt.Errorf("insufficient balance")
	}
	return err
}

func (s *Service) ListBanks() ([]BankListResponse, error) {
	res, err := s.paystack.ListBanks()
	if err != nil {
		return nil, err
	}
	var banks []BankListResponse
	for _, b := range res.Data {
		banks = append(banks, BankListResponse{Name: b.Name, Code: b.Code})
	}
	return banks, nil
}

func (s *Service) ResolveAccount(req ResolveAccountRequest) (*ResolveAccountResponse, error) {
	res, err := s.paystack.ResolveAccountNumber(req.AccountNumber, req.BankCode)
	if err != nil {
		return nil, fmt.Errorf("could not resolve account: %w", err)
	}
	return &ResolveAccountResponse{
		AccountNumber: res.Data.AccountNumber,
		AccountName:   res.Data.AccountName,
		BankCode:      req.BankCode,
	}, nil
}

func (s *Service) AddBeneficiary(userID string, req AddBeneficiaryRequest) (*Beneficiary, error) {
	resolved, err := s.ResolveAccount(ResolveAccountRequest{
		AccountNumber: req.AccountNumber,
		BankCode:      req.BankCode,
	})
	if err != nil {
		return nil, err
	}

	// Get bank name
	var bankName string
	banks, _ := s.ListBanks()
	for _, b := range banks {
		if b.Code == req.BankCode {
			bankName = b.Name
			break
		}
	}

	// Create Paystack recipient
	recRes, err := s.paystack.CreateTransferRecipient(paystack.CreateRecipientRequest{
		Name:          resolved.AccountName,
		AccountNumber: req.AccountNumber,
		BankCode:      req.BankCode,
	})
	if err != nil {
		return nil, fmt.Errorf("create recipient: %w", err)
	}

	benID := uuid.New().String()
	var ben Beneficiary
	err = db.DB.QueryRow(db.CreateBeneficiaryQuery,
		benID, userID, resolved.AccountName, req.AccountNumber, req.BankCode, bankName, recRes.Data.RecipientCode,
	).Scan(&ben.ID, &ben.AccountName, &ben.AccountNumber, &ben.BankName, &ben.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("save beneficiary: %w", err)
	}
	ben.RecipientCode = recRes.Data.RecipientCode
	return &ben, nil
}

func (s *Service) ListBeneficiaries(userID string) ([]Beneficiary, error) {
	rows, err := db.DB.Query(db.ListBeneficiariesQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Beneficiary
	for rows.Next() {
		var b Beneficiary
		rows.Scan(&b.ID, nil, &b.AccountName, &b.AccountNumber, nil, &b.BankName, &b.RecipientCode, &b.CreatedAt)
		list = append(list, b)
	}
	return list, nil
}

func (s *Service) DeleteBeneficiary(userID, benID string) error {
	res, err := db.DB.Exec(db.DeleteBeneficiaryQuery, benID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("beneficiary not found")
	}
	return nil
}
