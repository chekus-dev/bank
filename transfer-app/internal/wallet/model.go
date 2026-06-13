package wallet

import "time"

type Wallet struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Balance       float64   `json:"balance"`
	LedgerBalance float64   `json:"ledger_balance"`
	Currency      string    `json:"currency"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type FundWalletRequest struct {
	Amount   float64 `json:"amount"`
	Email    string  `json:"email"`
	Callback string  `json:"callback_url"`
}

type FundWalletResponse struct {
	AuthorizationURL string  `json:"authorization_url"`
	Reference        string  `json:"reference"`
	Amount           float64 `json:"amount"`
}

type BankListResponse struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type ResolveAccountRequest struct {
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
}

type ResolveAccountResponse struct {
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	BankCode      string `json:"bank_code"`
}

type AddBeneficiaryRequest struct {
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
}

type Beneficiary struct {
	ID            string    `json:"id"`
	AccountName   string    `json:"account_name"`
	AccountNumber string    `json:"account_number"`
	BankName      string    `json:"bank_name"`
	RecipientCode string    `json:"recipient_code"`
	CreatedAt     time.Time `json:"created_at"`
}
