package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"transfer-app/internal/db"
	"transfer-app/internal/notification"
	"transfer-app/internal/transfer"
	"transfer-app/internal/user"
	"transfer-app/internal/wallet"
	"transfer-app/pkg/paystack"

	"github.com/google/uuid"
)

type Service struct {
	paystackCl *paystack.Client
	walletSvc  *wallet.Service
	userSvc    *user.Service
	notifSvc   *notification.Service
}

func NewService(ps *paystack.Client, ws *wallet.Service, us *user.Service, ns *notification.Service) *Service {
	return &Service{paystackCl: ps, walletSvc: ws, userSvc: us, notifSvc: ns}
}

type PaystackEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type ChargeData struct {
	Reference string `json:"reference"`
	Status    string `json:"status"`
	Amount    int64  `json:"amount"`
	Metadata  struct {
		UserID string `json:"user_id"`
		Type   string `json:"type"`
	} `json:"metadata"`
	Customer struct {
		Email string `json:"email"`
	} `json:"customer"`
}

type TransferData struct {
	Reference    string `json:"reference"`
	Status       string `json:"status"`
	Amount       int64  `json:"amount"`
	TransferCode string `json:"transfer_code"`
}

func (s *Service) HandlePaystackEvent(event PaystackEvent) error {
	switch event.Event {
	case "charge.success":
		return s.handleChargeSuccess(event.Data)
	case "transfer.success":
		return s.handleTransferSuccess(event.Data)
	case "transfer.failed":
		return s.handleTransferFailed(event.Data)
	case "transfer.reversed":
		return s.handleTransferReversed(event.Data)
	default:
		log.Printf("ℹ️  unhandled paystack event: %s", event.Event)
		return nil
	}
}

func (s *Service) handleChargeSuccess(data json.RawMessage) error {
	var charge ChargeData
	if err := json.Unmarshal(data, &charge); err != nil {
		return fmt.Errorf("unmarshal charge: %w", err)
	}

	if charge.Status != "success" {
		return nil
	}
	if charge.Metadata.Type != "wallet_funding" {
		return nil
	}

	userID := charge.Metadata.UserID
	amount := float64(charge.Amount) / 100

	// Check idempotency
	var exists bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM transactions WHERE reference=$1)`, charge.Reference).Scan(&exists)
	if exists {
		log.Printf("ℹ️  duplicate webhook for ref %s, skipping", charge.Reference)
		return nil
	}

	w, err := s.walletSvc.GetWallet(userID)
	if err != nil {
		return fmt.Errorf("get wallet for user %s: %w", userID, err)
	}

	tx, err := db.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.walletSvc.CreditWallet(userID, amount, tx); err != nil {
		return fmt.Errorf("credit wallet: %w", err)
	}

	txnID := uuid.New().String()
	tx.Exec(db.CreateTransactionQuery,
		txnID, charge.Reference, nil, w.ID,
		charge.Amount, 0, transfer.TypeFunding, transfer.StatusSuccess,
		"Wallet funding via Paystack", "{}",
	)

	if err := tx.Commit(); err != nil {
		return err
	}

	u, _ := s.userSvc.GetByID(userID)
	if u != nil {
		go s.notifSvc.SendFundingSuccessNotification(u, amount, charge.Reference)
	}

	log.Printf("✅ wallet funded: user=%s amount=₦%.2f ref=%s", userID, amount, charge.Reference)
	return nil
}

func (s *Service) handleTransferSuccess(data json.RawMessage) error {
	var t TransferData
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	db.DB.Exec(db.UpdateTransactionStatusQuery, transfer.StatusSuccess, t.Reference)
	log.Printf("✅ bank transfer success: ref=%s", t.Reference)
	return nil
}

func (s *Service) handleTransferFailed(data json.RawMessage) error {
	var t TransferData
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	// Get transaction and reverse debit
	var txnID, walletID string
	var amountKobo, feeKobo int64
	err := db.DB.QueryRow(
		`SELECT id, sender_wallet_id, amount, fee FROM transactions WHERE reference=$1`, t.Reference,
	).Scan(&txnID, &walletID, &amountKobo, &feeKobo)
	if err != nil {
		return err
	}
	total := amountKobo + feeKobo
	db.DB.Exec(db.CreditWalletQuery, total, walletID)
	db.DB.Exec(db.UpdateTransactionStatusQuery, transfer.StatusFailed, txnID)
	log.Printf("❌ bank transfer failed, reversed: ref=%s", t.Reference)
	return nil
}

func (s *Service) handleTransferReversed(data json.RawMessage) error {
	var t TransferData
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	db.DB.Exec(db.UpdateTransactionStatusQuery, transfer.StatusReversed, t.Reference)
	log.Printf("🔄 transfer reversed: ref=%s", t.Reference)
	return nil
}
