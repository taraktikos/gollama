package main

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
	"github.com/taraktikos/gollama/gen/db"
	"github.com/tmc/langchaingo/llms/ollama"
)

func importData(ctx context.Context, llm *ollama.LLM, queries *db.Queries, filename string) error {
	zipFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open zip file: %w", err)
	}
	defer zipFile.Close()

	zipStat, err := zipFile.Stat()
	if err != nil {
		return fmt.Errorf("zip stat: %w", err)
	}

	zipReader, err := zip.NewReader(zipFile, zipStat.Size())
	if err != nil {
		return fmt.Errorf("zip reader: %w", err)
	}

	var csvFile io.ReadCloser
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, ".csv") {
			csvFile, err = file.Open()
			if err != nil {
				return fmt.Errorf("open csv file: %w", err)
			}
			defer csvFile.Close()
			break
		}
	}

	if csvFile == nil {
		return errors.New("CSV file not found in the ZIP archive")
	}

	csvReader := csv.NewReader(csvFile)

	line := 1
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read csv record on line %d: %w", line, err)
		}
		if record[0] == "content_id" {
			continue
		}

		embeddings, err := llm.CreateEmbedding(ctx, []string{record[4]})
		if err != nil {
			log.Fatal(err)
		}

		params := db.CreateWikiRecordParams{
			ContentID:    pgtype.Text{String: record[0], Valid: true},
			PageTitle:    pgtype.Text{String: record[1], Valid: true},
			SectionTitle: pgtype.Text{String: record[2], Valid: true},
			Breadcrumb:   pgtype.Text{String: record[3], Valid: true},
			Text:         pgtype.Text{String: record[4], Valid: true},
			Embedding:    pgvector.NewVector(embeddings[0]),
		}

		_, err = queries.CreateWikiRecord(ctx, params)
		if err != nil {
			return fmt.Errorf("create db record: %w", err)
		}

		slog.Info("create db record", slog.Int("line", line))
		line++
	}

	return nil
}
