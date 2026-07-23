package grpchandler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gen "github.com/jnanirudh/fund-platform/gen"
	"github.com/jnanirudh/fund-platform/search-service/internal/service"
)

// TransactionSearchHandler implements the TransactionSearchServiceServer gRPC interface.
type TransactionSearchHandler struct {
	gen.UnimplementedTransactionSearchServiceServer
	svc *service.SearchService
}

// NewTransactionSearchHandler returns a new TransactionSearchHandler wired to the given service.
func NewTransactionSearchHandler(svc *service.SearchService) *TransactionSearchHandler {
	return &TransactionSearchHandler{svc: svc}
}

// SearchTransactions executes a search and returns matching transactions.
func (h *TransactionSearchHandler) SearchTransactions(ctx context.Context, req *gen.SearchTransactionsRequest) (*gen.SearchTransactionsResponse, error) {
	transactions, err := h.svc.Search(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return &gen.SearchTransactionsResponse{
		Transactions: transactions,
		Total:        int32(len(transactions)),
	}, nil
}
