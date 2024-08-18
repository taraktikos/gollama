package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/pgvector/pgvector-go"
	"github.com/taraktikos/gollama/gen/db"
	gollamav1 "github.com/taraktikos/gollama/gen/gollama/v1"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type GollamaServer struct {
	queries *db.Queries
	llm     *ollama.LLM
}

func (s *GollamaServer) Search(ctx context.Context, req *connect.Request[gollamav1.SearchRequest]) (*connect.Response[gollamav1.SearchResponse], error) {
	slog.Info("search request", slog.String("text", req.Msg.Text), slog.Any("headers", req.Header()))

	embeddings, err := s.llm.CreateEmbedding(ctx, []string{req.Msg.Text})
	if err != nil {
		log.Fatal(err)
	}

	records, err := s.queries.GetMostSimilarRecord(ctx, pgvector.NewVector(embeddings[0]))
	if err != nil {
		return nil, fmt.Errorf("query similar record: %w", err)
	}

	if len(records) == 0 {
		return nil, errors.New("no similar record")
	}

	return connect.NewResponse(&gollamav1.SearchResponse{
		Text: records[0].Text.String,
	}), nil
}

func (s *GollamaServer) GenerateFromSinglePrompt(ctx context.Context, req *connect.Request[gollamav1.GenerateFromSinglePromptRequest]) (*connect.Response[gollamav1.GenerateFromSinglePromptResponse], error) {
	slog.Info("generate from single prompt request", slog.String("prompt", req.Msg.Prompt), slog.Any("headers", req.Header()))

	completion, err := llms.GenerateFromSinglePrompt(ctx, s.llm, req.Msg.Prompt)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&gollamav1.GenerateFromSinglePromptResponse{
		Content: completion,
	}), nil
}
