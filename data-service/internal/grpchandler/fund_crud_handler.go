package grpchandler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gen "github.com/jnanirudh/fund-platform/gen"
	"github.com/jnanirudh/fund-platform/data-service/internal/service"
)

// TransactionHandler implements the TransactionServiceServer gRPC interface.
type TransactionHandler struct {
	gen.UnimplementedTransactionServiceServer
	svc *service.TransactionService
}

// NewTransactionHandler returns a new TransactionHandler wired to the given service.
func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

// CreateTransaction creates a new fund transaction.
func (h *TransactionHandler) CreateTransaction(ctx context.Context, req *gen.CreateTransactionRequest) (*gen.TransactionResponse, error) {
	txn, err := h.svc.Create(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return &gen.TransactionResponse{Transaction: txn}, nil
}

// GetTransaction retrieves a single transaction by trx_id.
func (h *TransactionHandler) GetTransaction(ctx context.Context, req *gen.GetTransactionRequest) (*gen.TransactionResponse, error) {
	txn, err := h.svc.Get(ctx, req.TrxId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}
	return &gen.TransactionResponse{Transaction: txn}, nil
}

// UpdateTransaction updates an existing transaction.
func (h *TransactionHandler) UpdateTransaction(ctx context.Context, req *gen.UpdateTransactionRequest) (*gen.TransactionResponse, error) {
	txn, err := h.svc.Update(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &gen.TransactionResponse{Transaction: txn}, nil
}

// DeleteTransaction removes a transaction by trx_id.
func (h *TransactionHandler) DeleteTransaction(ctx context.Context, req *gen.DeleteTransactionRequest) (*gen.DeleteTransactionResponse, error) {
	if err := h.svc.Delete(ctx, req.TrxId); err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}
	return &gen.DeleteTransactionResponse{Success: true, Message: "transaction deleted successfully"}, nil
}

// ListTransactions returns all transactions.
func (h *TransactionHandler) ListTransactions(ctx context.Context, req *gen.ListTransactionsRequest) (*gen.ListTransactionsResponse, error) {
	transactions, err := h.svc.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &gen.ListTransactionsResponse{Transactions: transactions}, nil
}
