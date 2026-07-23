package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	gen "github.com/jnanirudh/fund-platform/gen"
)

// TransactionRepository handles all database operations for fund transactions.
type TransactionRepository struct {
	pool *pgxpool.Pool
}

// NewTransactionRepository creates a new TransactionRepository backed by the given connection pool.
func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

// Create inserts a new Transaction and its associated FundDetails in a single transaction.
func (r *TransactionRepository) Create(ctx context.Context, txn *gen.Transaction) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `
		INSERT INTO fund_entries (trx_id, user_id, user_name, user_pan_num, date_of_purchase, nav, no_of_units)
		VALUES ($1, $2, $3, $4, $5::date, $6, $7)
	`, txn.TrxId, txn.UserId, txn.UserName, txn.UserPanNum, txn.DateOfPurchase, txn.Nav, txn.NoOfUnits)
	if err != nil {
		return fmt.Errorf("insert fund_entry: %w", err)
	}

	for _, d := range txn.FundDetails {
		_, err = tx.Exec(ctx, `
			INSERT INTO mf_details (trx_id, fund_name, amc_name, type)
			VALUES ($1, $2, $3, $4)
		`, txn.TrxId, d.FundName, d.AmcName, d.Type)
		if err != nil {
			return fmt.Errorf("insert mf_detail: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetByID fetches a single Transaction (with its FundDetails) by trx_id.
func (r *TransactionRepository) GetByID(ctx context.Context, trxID string) (*gen.Transaction, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT trx_id, user_id, user_name, user_pan_num, date_of_purchase::text, nav, no_of_units
		FROM fund_entries
		WHERE trx_id = $1
	`, trxID)

	txn := &gen.Transaction{}
	if err := row.Scan(
		&txn.TrxId,
		&txn.UserId,
		&txn.UserName,
		&txn.UserPanNum,
		&txn.DateOfPurchase,
		&txn.Nav,
		&txn.NoOfUnits,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %s", trxID)
		}
		return nil, fmt.Errorf("scan fund_entry: %w", err)
	}

	details, err := r.getFundDetails(ctx, trxID)
	if err != nil {
		return nil, err
	}
	txn.FundDetails = details

	return txn, nil
}

// Update replaces the mutable fields of a Transaction and resets its FundDetails.
func (r *TransactionRepository) Update(ctx context.Context, txn *gen.Transaction) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	tag, err := tx.Exec(ctx, `
		UPDATE fund_entries
		SET user_name        = $1,
		    user_pan_num     = $2,
		    date_of_purchase = $3::date,
		    nav              = $4,
		    no_of_units      = $5
		WHERE trx_id = $6
	`, txn.UserName, txn.UserPanNum, txn.DateOfPurchase, txn.Nav, txn.NoOfUnits, txn.TrxId)
	if err != nil {
		return fmt.Errorf("update fund_entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("transaction not found: %s", txn.TrxId)
	}

	if _, err = tx.Exec(ctx, `DELETE FROM mf_details WHERE trx_id = $1`, txn.TrxId); err != nil {
		return fmt.Errorf("delete old fund details: %w", err)
	}

	for _, d := range txn.FundDetails {
		_, err = tx.Exec(ctx, `
			INSERT INTO mf_details (trx_id, fund_name, amc_name, type)
			VALUES ($1, $2, $3, $4)
		`, txn.TrxId, d.FundName, d.AmcName, d.Type)
		if err != nil {
			return fmt.Errorf("insert fund detail: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// Delete removes a Transaction by trx_id (FundDetails cascade via FK).
func (r *TransactionRepository) Delete(ctx context.Context, trxID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM fund_entries WHERE trx_id = $1`, trxID)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("transaction not found: %s", trxID)
	}
	return nil
}

// List returns all Transactions ordered by date_of_purchase descending.
func (r *TransactionRepository) List(ctx context.Context) ([]*gen.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT trx_id, user_id, user_name, user_pan_num, date_of_purchase::text, nav, no_of_units
		FROM fund_entries
		ORDER BY date_of_purchase DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*gen.Transaction
	for rows.Next() {
		t := &gen.Transaction{}
		if err := rows.Scan(
			&t.TrxId, &t.UserId, &t.UserName, &t.UserPanNum, &t.DateOfPurchase, &t.Nav, &t.NoOfUnits,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	for _, t := range transactions {
		t.FundDetails, err = r.getFundDetails(ctx, t.TrxId)
		if err != nil {
			return nil, err
		}
	}

	return transactions, nil
}

// getFundDetails fetches all FundDetail rows for a given trx_id.
func (r *TransactionRepository) getFundDetails(ctx context.Context, trxID string) ([]*gen.FundDetail, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT fund_name, amc_name, type
		FROM mf_details
		WHERE trx_id = $1
	`, trxID)
	if err != nil {
		return nil, fmt.Errorf("query fund details: %w", err)
	}
	defer rows.Close()

	var details []*gen.FundDetail
	for rows.Next() {
		d := &gen.FundDetail{}
		if err := rows.Scan(&d.FundName, &d.AmcName, &d.Type); err != nil {
			return nil, fmt.Errorf("scan fund detail: %w", err)
		}
		details = append(details, d)
	}
	return details, rows.Err()
}
