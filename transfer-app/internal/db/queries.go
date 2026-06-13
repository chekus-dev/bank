package db

const (
	// ── Users ──────────────────────────────────────────────────────────────
	CreateUserQuery = `
		INSERT INTO users (id, first_name, last_name, email, phone, password_hash, pin_hash, bvn, is_verified, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, first_name, last_name, email, phone, is_verified, is_active, created_at`

	GetUserByIDQuery = `
		SELECT id, first_name, last_name, email, phone, password_hash, pin_hash, bvn, is_verified, is_active, created_at, updated_at
		FROM users WHERE id = $1 AND is_active = true`

	GetUserByEmailQuery = `
		SELECT id, first_name, last_name, email, phone, password_hash, pin_hash, bvn, is_verified, is_active, created_at, updated_at
		FROM users WHERE email = $1`

	GetUserByPhoneQuery = `
		SELECT id, first_name, last_name, email, phone, password_hash, pin_hash, bvn, is_verified, is_active, created_at, updated_at
		FROM users WHERE phone = $1`

	UpdateUserVerifiedQuery = `UPDATE users SET is_verified = true, updated_at = NOW() WHERE id = $1`

	UpdateUserPinQuery = `UPDATE users SET pin_hash = $1, updated_at = NOW() WHERE id = $2`

	UpdateUserPasswordQuery = `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`

	// ── OTP ────────────────────────────────────────────────────────────────
	UpsertOTPQuery = `
		INSERT INTO otps (phone, pin_id, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (phone) DO UPDATE SET pin_id = EXCLUDED.pin_id, expires_at = EXCLUDED.expires_at, created_at = NOW()`

	GetOTPQuery = `SELECT pin_id, expires_at FROM otps WHERE phone = $1`

	DeleteOTPQuery = `DELETE FROM otps WHERE phone = $1`

	// ── Wallets ────────────────────────────────────────────────────────────
	CreateWalletQuery = `
		INSERT INTO wallets (id, user_id, balance, ledger_balance, currency, is_active, created_at, updated_at)
		VALUES ($1, $2, 0, 0, 'NGN', true, NOW(), NOW())
		RETURNING id, user_id, balance, ledger_balance, currency, is_active, created_at`

	GetWalletByUserIDQuery = `
		SELECT id, user_id, balance, ledger_balance, currency, is_active, created_at, updated_at
		FROM wallets WHERE user_id = $1 AND is_active = true`

	GetWalletByIDQuery = `
		SELECT id, user_id, balance, ledger_balance, currency, is_active, created_at, updated_at
		FROM wallets WHERE id = $1 AND is_active = true`

	CreditWalletQuery = `
		UPDATE wallets SET balance = balance + $1, ledger_balance = ledger_balance + $1, updated_at = NOW()
		WHERE id = $2 AND is_active = true RETURNING balance`

	DebitWalletQuery = `
		UPDATE wallets SET balance = balance - $1, ledger_balance = ledger_balance - $1, updated_at = NOW()
		WHERE id = $2 AND balance >= $1 AND is_active = true RETURNING balance`

	// ── Transactions ───────────────────────────────────────────────────────
	CreateTransactionQuery = `
		INSERT INTO transactions (id, reference, sender_wallet_id, receiver_wallet_id, amount, fee, type, status, description, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, reference, amount, fee, type, status, created_at`

	GetTransactionByReferenceQuery = `
		SELECT id, reference, sender_wallet_id, receiver_wallet_id, amount, fee, type, status, description, metadata, created_at, updated_at
		FROM transactions WHERE reference = $1`

	GetTransactionByIDQuery = `
		SELECT id, reference, sender_wallet_id, receiver_wallet_id, amount, fee, type, status, description, metadata, created_at, updated_at
		FROM transactions WHERE id = $1`

	UpdateTransactionStatusQuery = `
		UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`

	ListTransactionsByWalletQuery = `
		SELECT id, reference, sender_wallet_id, receiver_wallet_id, amount, fee, type, status, description, created_at
		FROM transactions
		WHERE sender_wallet_id = $1 OR receiver_wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	CountTransactionsByWalletQuery = `
		SELECT COUNT(*) FROM transactions
		WHERE sender_wallet_id = $1 OR receiver_wallet_id = $1`

	// ── Beneficiaries ──────────────────────────────────────────────────────
	CreateBeneficiaryQuery = `
		INSERT INTO beneficiaries (id, user_id, account_name, account_number, bank_code, bank_name, recipient_code, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id, account_name, account_number, bank_name, created_at`

	ListBeneficiariesQuery = `
		SELECT id, user_id, account_name, account_number, bank_code, bank_name, recipient_code, created_at
		FROM beneficiaries WHERE user_id = $1 ORDER BY created_at DESC`

	DeleteBeneficiaryQuery = `DELETE FROM beneficiaries WHERE id = $1 AND user_id = $2`
)
