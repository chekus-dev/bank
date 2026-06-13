package transfer

import (
	"database/sql"
	"fmt"
	"transfer-app/config"
	"transfer-app/internal/db"
	"transfer-app/internal/notification"
	"transfer-app/internal/user"
	"transfer-app/internal/wallet"
	"transfer-app/pkg/paystack"

	"github.com/google/uuid"
)

type Service struct {
	walletSvc  *wallet.Service
	userSvc    *user.Service
	notifSvc   *notification.Service
	paystackCl *paystack.Client
}

func NewService(ws *wallet.Service, us *user.Service, ns *notification.Service, ps *paystack.Client) *Service {
	return &Service{walletSvc: ws, userSvc: us, notifSvc: ns, paystackCl: ps}
}

// ── P2P Transfer ─────────────────────────────────────────────────────────────

func (s *Service) P2PTransfer(senderID string, req P2PTransferRequest) (*TransferResponse, error) {
	if req.Amount < config.App.MinTransferAmt {
		return nil, fmt.Errorf("minimum transfer is ₦%.2f", config.App.MinTransferAmt)
	}
	if req.Amount > config.App.MaxTransferAmt {
		return nil, fmt.Errorf("maximum transfer is ₦%.2f", config.App.MaxTransferAmt)
	}

	sender, err := s.userSvc.GetByID(senderID)
	if err != nil {
		return nil, fmt.Errorf("sender not found")
	}
	if !s.userSvc.VerifyPin(sender, req.Pin) {
		return nil, fmt.Errorf("invalid transaction PIN")
	}

	receiver, err := s.userSvc.GetByPhone(req.ReceiverPhone)
	if err != nil {
		return nil, fmt.Errorf("receiver not found")
	}
	if receiver.ID == senderID {
		return nil, fmt.Errorf("cannot transfer to yourself")
	}

	senderWallet, err := s.walletSvc.GetWallet(senderID)
	if err != nil {
		return nil, fmt.Errorf("sender wallet not found")
	}
	receiverWallet, err := s.walletSvc.GetWallet(receiver.ID)
	if err != nil {
		return nil, fmt.Errorf("receiver wallet not found")
	}

	fee := config.App.TransferFee
	total := req.Amount + fee
	if senderWallet.Balance < total {
		return nil, fmt.Errorf("insufficient balance. Need ₦%.2f (amount + ₦%.2f fee)", total, fee)
	}

	tx, err := db.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.walletSvc.DebitWallet(senderWallet.ID, total, tx); err != nil {
		return nil, fmt.Errorf("debit sender: %w", err)
	}
	if err := s.walletSvc.CreditWallet(receiver.ID, req.Amount, tx); err != nil {
		return nil, fmt.Errorf("credit receiver: %w", err)
	}

	ref := "P2P-" + uuid.New().String()
	desc := req.Description
	if desc == "" {
		desc = fmt.Sprintf("Transfer from %s %s", sender.FirstName, sender.LastName)
	}

	txnID := uuid.New().String()
	var txn Transaction
	err = tx.QueryRow(db.CreateTransactionQuery,
		txnID, ref, senderWallet.ID, receiverWallet.ID,
		int64(req.Amount*100), int64(fee*100),
		TypeP2P, StatusSuccess, desc, "{}",
	).Scan(&txn.ID, &txn.Reference, &txn.Amount, &txn.Fee, &txn.Type, &txn.Status, &txn.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("record transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// Notifications (non-blocking)
	go s.notifSvc.SendTransferNotification(sender, receiver, req.Amount, fee, ref)

	return &TransferResponse{
		Reference:   ref,
		Amount:      req.Amount,
		Fee:         fee,
		Status:      StatusSuccess,
		Description: desc,
		CreatedAt:   txn.CreatedAt,
	}, nil
}

// ── Bank Transfer ─────────────────────────────────────────────────────────────

func (s *Service) BankTransfer(senderID string, req BankTransferRequest) (*TransferResponse, error) {
	if req.Amount < config.App.MinTransferAmt {
		return nil, fmt.Errorf("minimum transfer is ₦%.2f", config.App.MinTransferAmt)
	}

	sender, err := s.userSvc.GetByID(senderID)
	if err != nil {
		return nil, fmt.Errorf("sender not found")
	}
	if !s.userSvc.VerifyPin(sender, req.Pin) {
		return nil, fmt.Errorf("invalid transaction PIN")
	}

	senderWallet, err := s.walletSvc.GetWallet(senderID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found")
	}

	fee := config.App.TransferFee
	total := req.Amount + fee
	if senderWallet.Balance < total {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Create Paystack recipient
	recRes, err := s.paystackCl.CreateTransferRecipient(paystack.CreateRecipientRequest{
		Name:          req.AccountName,
		AccountNumber: req.AccountNumber,
		BankCode:      req.BankCode,
	})
	if err != nil {
		return nil, fmt.Errorf("create recipient: %w", err)
	}

	ref := "BANK-" + uuid.New().String()

	// Debit wallet first
	tx, err := db.BeginTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.walletSvc.DebitWallet(senderWallet.ID, total, tx); err != nil {
		return nil, fmt.Errorf("debit wallet: %w", err)
	}

	desc := req.Description
	if desc == "" {
		desc = fmt.Sprintf("Bank transfer to %s", req.AccountName)
	}

	txnID := uuid.New().String()
	var txn Transaction
	tx.QueryRow(db.CreateTransactionQuery,
		txnID, ref, senderWallet.ID, nil,
		int64(req.Amount*100), int64(fee*100),
		TypeBank, StatusPending, desc, "{}",
	).Scan(&txn.ID, &txn.Reference, &txn.Amount, &txn.Fee, &txn.Type, &txn.Status, &txn.CreatedAt)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Initiate Paystack transfer
	_, psErr := s.paystackCl.InitiateTransfer(paystack.InitiateTransferRequest{
		Amount:    int64(req.Amount * 100),
		Recipient: recRes.Data.RecipientCode,
		Reason:    desc,
		Reference: ref,
	})
	if psErr != nil {
		// Reverse debit
		db.DB.Exec(db.CreditWalletQuery, int64(total*100), senderWallet.ID)
		db.DB.Exec(db.UpdateTransactionStatusQuery, StatusFailed, txnID)
		return nil, fmt.Errorf("transfer failed: %w", psErr)
	}

	go s.notifSvc.SendBankTransferNotification(sender, req.Amount, fee, req.AccountName, ref)

	return &TransferResponse{
		Reference:   ref,
		Amount:      req.Amount,
		Fee:         fee,
		Status:      StatusPending,
		Description: desc,
		CreatedAt:   txn.CreatedAt,
	}, nil
}

// ── Transaction History ───────────────────────────────────────────────────────

func (s *Service) GetTransactions(userID string, page, perPage int) ([]Transaction, int64, error) {
	w, err := s.walletSvc.GetWallet(userID)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := db.DB.Query(db.ListTransactionsByWalletQuery, w.ID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txns []Transaction
	for rows.Next() {
		var t Transaction
		rows.Scan(&t.ID, &t.Reference, &t.SenderWalletID, &t.ReceiverWalletID,
			&t.Amount, &t.Fee, &t.Type, &t.Status, &t.Description, &t.CreatedAt)
		t.Amount /= 100
		t.Fee /= 100
		txns = append(txns, t)
	}

	var total int64
	db.DB.QueryRow(db.CountTransactionsByWalletQuery, w.ID).Scan(&total)

	return txns, total, nil
}

func (s *Service) GetTransaction(reference string) (*Transaction, error) {
	var t Transaction
	err := db.DB.QueryRow(db.GetTransactionByReferenceQuery, reference).Scan(
		&t.ID, &t.Reference, &t.SenderWalletID, &t.ReceiverWalletID,
		&t.Amount, &t.Fee, &t.Type, &t.Status, &t.Description, &t.Metadata,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found")
	}
	t.Amount /= 100
	t.Fee /= 100
	return &t, err
}
