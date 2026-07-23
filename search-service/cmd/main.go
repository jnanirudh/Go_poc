package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/jnanirudh/fund-platform/search-service/internal/config"
	"github.com/jnanirudh/fund-platform/search-service/internal/grpchandler"
	"github.com/jnanirudh/fund-platform/search-service/internal/repository"
	"github.com/jnanirudh/fund-platform/search-service/internal/service"
	gen "github.com/jnanirudh/fund-platform/gen"
)

func main() {
	cfg := config.Load()

	// ── Database (read-only Postgres FTS) ─────────────────────────────────────
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("✓ connected to PostgreSQL")

	// ── Dependency wiring ─────────────────────────────────────────────────────
	repo := repository.NewSearchRepository(pool)

	svc, cleanup, err := service.NewSearchService(repo, cfg.DataServiceAddr)
	if err != nil {
		log.Fatalf("failed to connect to data-service at %s: %v", cfg.DataServiceAddr, err)
	}
	defer cleanup()
	log.Printf("✓ data-service gRPC client connected to %s", cfg.DataServiceAddr)

	handler := grpchandler.NewTransactionSearchHandler(svc)

	// ── gRPC server ───────────────────────────────────────────────────────────
	grpcServer := grpc.NewServer()
	gen.RegisterTransactionSearchServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("✓ search-service gRPC server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down search-service gracefully...")
	grpcServer.GracefulStop()
	log.Println("search-service stopped")
}
