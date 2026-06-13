package transfer

import "time"

type Transaction struct {
	ID               string    `json:"id"`
	Reference        string    `json:"reference"`
	SenderWalletID   string    `json:"sender_wallet_id"`
	ReceiverWalletID string    `json:"receiver_wallet_id"`
	Amount           float64   `json:"amount"`
	Fee              float64   `json:"fee"`
	Type             string    `json:"type"`
	Status           string    `json:"status"`
	Description      string    `json:"description"`
	Metadata         string    `json:"metadata,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type P2PTransferRequest struct {
	ReceiverPhone string  `json:"receiver_phone"`
	Amount        float64 `json:"amount"`
	Pin           string  `json:"pin"`
	Description   string  `json:"description"`
}

type BankTransferRequest struct {
	AccountNumber string  `json:"account_number"`
	BankCode      string  `json:"bank_code"`
	AccountName   string  `json:"account_name"`
	Amount        float64 `json:"amount"`
	Pin           string  `json:"pin"`
	Description   string  `json:"description"`
	SaveBeneficiary bool  `json:"save_beneficiary"`
}

type TransferResponse struct {
	Reference   string    `json:"reference"`
	Amount      float64   `json:"amount"`
	Fee         float64   `json:"fee"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListTransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

const (
	StatusPending  = "pending"
	StatusSuccess  = "success"
	StatusFailed   = "failed"
	StatusReversed = "reversed"

	TypeP2P      = "p2p"
	TypeBank     = "bank_transfer"
	TypeFunding  = "funding"
	TypeReversal = "reversal"
)
