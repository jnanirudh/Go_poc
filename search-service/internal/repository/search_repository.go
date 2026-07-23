package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SearchRepository performs full-text and filter-based search queries on the
// shared PostgreSQL database. It is read-only and never writes any data.
type SearchRepository struct {
	pool *pgxpool.Pool
}

// NewSearchRepository creates a new SearchRepository backed by the given pool.
func NewSearchRepository(pool *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{pool: pool}
}

// SearchByQuery executes a Postgres full-text search and optional field filters,
// returning the matching trx_ids. At least one search parameter must be provided.
//
// Search priority:
//  1. Full-text query (fund_name, amc_name, type, user_name)
//  2. Exact user_id filter
//  3. LIKE fund_name filter
func (r *SearchRepository) SearchByQuery(ctx context.Context, query, userID, fundName string) ([]string, error) {
	if query == "" && userID == "" && fundName == "" {
		return nil, fmt.Errorf("at least one of query, user_id, or fund_name must be provided")
	}

	// Build query dynamically to only apply active filters
	var conditions []string
	var args []any
	argIdx := 1

	if query != "" {
		conditions = append(conditions, fmt.Sprintf(`
			to_tsvector('english',
				md.fund_name || ' ' || md.amc_name || ' ' || md.type || ' ' || fe.user_name
			) @@ plainto_tsquery('english', $%d)`, argIdx))
		args = append(args, query)
		argIdx++
	}

	if userID != "" {
		conditions = append(conditions, fmt.Sprintf(`fe.user_id = $%d`, argIdx))
		args = append(args, userID)
		argIdx++
	}

	if fundName != "" {
		conditions = append(conditions, fmt.Sprintf(`LOWER(md.fund_name) LIKE LOWER($%d)`, argIdx))
		args = append(args, "%"+fundName+"%")
	}

	finalQuery := `
		SELECT DISTINCT fe.trx_id
		FROM fund_entries fe
		JOIN mf_details md ON fe.trx_id = md.trx_id
		WHERE ` + strings.Join(conditions, " AND ")

	rows, err := r.pool.Query(ctx, finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("execute search query: %w", err)
	}
	defer rows.Close()

	var trxIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan trx_id: %w", err)
		}
		trxIDs = append(trxIDs, id)
	}

	return trxIDs, rows.Err()
}
