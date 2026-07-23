package service

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gen "github.com/jnanirudh/fund-platform/gen"
	"github.com/jnanirudh/fund-platform/search-service/internal/repository"
)

// SearchService orchestrates Postgres full-text search and result hydration
// via gRPC calls to the data-service.
type SearchService struct {
	repo                  *repository.SearchRepository
	transactionSvcClient  gen.TransactionServiceClient
}

// NewSearchService creates a SearchService and establishes a gRPC client
// connection to the data-service. The returned cleanup function must be called
// on shutdown to close the connection.
func NewSearchService(repo *repository.SearchRepository, dataServiceAddr string) (*SearchService, func(), error) {
	conn, err := grpc.NewClient(
		dataServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("dial data-service at %s: %w", dataServiceAddr, err)
	}

	cleanup := func() { conn.Close() }

	return &SearchService{
		repo:                 repo,
		transactionSvcClient: gen.NewTransactionServiceClient(conn),
	}, cleanup, nil
}

// Search performs a Postgres full-text search to find matching trx_ids, then
// hydrates each result by calling the data-service GetTransaction RPC.
func (s *SearchService) Search(ctx context.Context, req *gen.SearchTransactionsRequest) ([]*gen.Transaction, error) {
	trxIDs, err := s.repo.SearchByQuery(ctx, req.Query, req.UserId, req.FundName)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	transactions := make([]*gen.Transaction, 0, len(trxIDs))
	for _, id := range trxIDs {
		resp, err := s.transactionSvcClient.GetTransaction(ctx, &gen.GetTransactionRequest{TrxId: id})
		if err != nil {
			log.Printf("warn: could not hydrate transaction %s from data-service: %v", id, err)
			continue
		}
		transactions = append(transactions, resp.Transaction)
	}

	return transactions, nil
}
