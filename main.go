package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sethvargo/go-envconfig"

	"github.com/taraktikos/gollama/gen/db"
	"github.com/taraktikos/gollama/gen/gollama/v1/gollamav1connect"
	"github.com/tmc/langchaingo/llms/ollama"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	ctx := context.Background()

	logger := slog.New(devslog.NewHandler(os.Stdout, &devslog.Options{
		HandlerOptions: &slog.HandlerOptions{AddSource: false, Level: slog.LevelDebug},
	}))

	slog.SetDefault(logger)

	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		slog.Error("failed to parse config", slog.String("err", err.Error()))
		os.Exit(1)
	}

	if cfg.ImportFilePath != "" {
		if err := runImport(ctx, cfg); err != nil {
			slog.Error("run import failed", slog.Any("err", err))
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := runServer(ctx, cfg); err != nil {
		slog.Error("run server failed", slog.Any("err", err))
		os.Exit(1)
	}

}

func runImport(ctx context.Context, cfg Config) error {
	dbpool, err := pgxpool.New(ctx, dsn(cfg))
	if err != nil {
		return fmt.Errorf("new pgxpool: %w", err)
	}
	defer dbpool.Close()

	queries := db.New(dbpool)

	llm, err := ollama.New(ollama.WithModel(cfg.OllamaModel))
	if err != nil {
		return fmt.Errorf("new ollama: %w", err)
	}

	slog.Info("import data started")
	if err := importData(ctx, llm, queries, cfg.ImportFilePath); err != nil {
		return fmt.Errorf("import data: %w", err)
	}
	slog.Info("import data finished")

	return nil
}

func runServer(ctx context.Context, cfg Config) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	start := time.Now()

	dbpool, err := pgxpool.New(ctx, dsn(cfg))
	if err != nil {
		return fmt.Errorf("new pgxpool: %w", err)
	}
	defer dbpool.Close()

	llm, err := ollama.New(ollama.WithModel(cfg.OllamaModel))
	if err != nil {
		return fmt.Errorf("new ollama: %w", err)
	}

	srv := &GollamaServer{
		queries: db.New(dbpool),
		llm:     llm,
	}
	mux := http.NewServeMux()
	path, handler := gollamav1connect.NewGollamaServiceHandler(srv)

	mux.Handle(path, handler)

	addr := net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port)
	server := &http.Server{Addr: addr, Handler: h2c.NewHandler(mux, &http2.Server{})}
	go func() {
		<-ctx.Done()
		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("shutdown server", slog.Any("err", err))
		}
	}()

	slog.Info("listening http", slog.String("addr", addr), slog.Duration("time", time.Since(start)))
	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("listening http server: %w", err)
	}

	return nil
}

func dsn(cfg Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		net.JoinHostPort(cfg.Postgres.Host, cfg.Postgres.Port),
		cfg.Postgres.Database,
		cfg.Postgres.SSLMode,
	)
}
