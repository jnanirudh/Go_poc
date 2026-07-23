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

	"github.com/jnanirudh/fund-platform/data-service/internal/config"
	"github.com/jnanirudh/fund-platform/data-service/internal/grpchandler"
	"github.com/jnanirudh/fund-platform/data-service/internal/repository"
	"github.com/jnanirudh/fund-platform/data-service/internal/service"
	gen "github.com/jnanirudh/fund-platform/gen"
)

func main() {
	cfg := config.Load()

	// ── Database ──────────────────────────────────────────────────────────────
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
	repo    := repository.NewTransactionRepository(pool)
	svc     := service.NewTransactionService(repo)
	handler := grpchandler.NewTransactionHandler(svc)

	// ── gRPC server ───────────────────────────────────────────────────────────
	grpcServer := grpc.NewServer()
	gen.RegisterTransactionServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("✓ data-service gRPC server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down data-service gracefully...")
	grpcServer.GracefulStop()
	log.Println("data-service stopped")
}
