CREATE EXTENSION IF NOT EXISTS "vector";
CREATE TABLE wiki_records (
    id SERIAL PRIMARY KEY,
    content_id TEXT,
    page_title TEXT,
    section_title TEXT,
    breadcrumb TEXT,
    text TEXT,
    embedding VECTOR(4096),
    metadata JSONB
);
-- Uncomment it after https://github.com/pgvector/pgvector/issues/461
-- CREATE INDEX wiki_records_ivfflat_index ON wiki_records USING ivfflat (embedding vector_cosine_ops) WITH (lists = 10);
