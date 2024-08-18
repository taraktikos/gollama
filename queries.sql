-- name: GetWikiRecordsCount :one
SELECT count(*) FROM wiki_records;

-- name: CreateWikiRecord :one
INSERT INTO wiki_records (
  content_id, page_title, section_title, breadcrumb, text, embedding, metadata
) VALUES (
  @content_id, @page_title, @section_title, @breadcrumb, @text, @embedding, @metadata
)
RETURNING *;

-- name: GetMostSimilarRecord :many
SELECT * FROM wiki_records ORDER BY embedding <=> @embedding LIMIT 1;
