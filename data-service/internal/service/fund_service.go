package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	gen "github.com/jnanirudh/fund-platform/gen"
	"github.com/jnanirudh/fund-platform/data-service/internal/repository"
)

// TransactionService contains the business logic for fund transaction operations.
type TransactionService struct {
	repo *repository.TransactionRepository
}

// NewTransactionService creates a new TransactionService with the given repository.
func NewTransactionService(repo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

// Create validates the request, generates a trx_id, and persists the transaction.
func (s *TransactionService) Create(ctx context.Context, req *gen.CreateTransactionRequest) (*gen.Transaction, error) {
	if req.UserId == "" || req.UserName == "" || req.UserPanNum == "" {
		return nil, fmt.Errorf("user_id, user_name, and user_pan_num are required")
	}
	if req.DateOfPurchase == "" {
		return nil, fmt.Errorf("date_of_purchase is required (format: YYYY-MM-DD)")
	}
	if len(req.FundDetails) == 0 {
		return nil, fmt.Errorf("at least one fund_detail is required")
	}

	txn := &gen.Transaction{
		TrxId:          uuid.NewString(),
		UserId:         req.UserId,
		UserName:       req.UserName,
		UserPanNum:     req.UserPanNum,
		FundDetails:    req.FundDetails,
		DateOfPurchase: req.DateOfPurchase,
		Nav:            req.Nav,
		NoOfUnits:      req.NoOfUnits,
	}

	if err := s.repo.Create(ctx, txn); err != nil {
		return nil, fmt.Errorf("persist transaction: %w", err)
	}
	return txn, nil
}

// Get retrieves a single transaction by its trx_id.
func (s *TransactionService) Get(ctx context.Context, trxID string) (*gen.Transaction, error) {
	if trxID == "" {
		return nil, fmt.Errorf("trx_id is required")
	}
	return s.repo.GetByID(ctx, trxID)
}

// Update applies changes to an existing transaction (mutable fields only).
// user_id is immutable once the transaction is created.
func (s *TransactionService) Update(ctx context.Context, req *gen.UpdateTransactionRequest) (*gen.Transaction, error) {
	if req.TrxId == "" {
		return nil, fmt.Errorf("trx_id is required")
	}
	if len(req.FundDetails) == 0 {
		return nil, fmt.Errorf("at least one fund_detail is required")
	}

	txn := &gen.Transaction{
		TrxId:          req.TrxId,
		UserName:       req.UserName,
		UserPanNum:     req.UserPanNum,
		FundDetails:    req.FundDetails,
		DateOfPurchase: req.DateOfPurchase,
		Nav:            req.Nav,
		NoOfUnits:      req.NoOfUnits,
	}

	if err := s.repo.Update(ctx, txn); err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}

	return s.repo.GetByID(ctx, req.TrxId)
}

// Delete removes a transaction and its associated fund details.
func (s *TransactionService) Delete(ctx context.Context, trxID string) error {
	if trxID == "" {
		return fmt.Errorf("trx_id is required")
	}
	return s.repo.Delete(ctx, trxID)
}

// List returns all transactions.
func (s *TransactionService) List(ctx context.Context) ([]*gen.Transaction, error) {
	return s.repo.List(ctx)
}
