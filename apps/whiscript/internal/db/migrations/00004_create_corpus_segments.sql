-- +goose Up
-- +goose StatementBegin
CREATE TABLE corpus_segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    corpus_file_id INTEGER NOT NULL,
    segment_index INTEGER NOT NULL,
    start_time REAL NOT NULL,
    end_time REAL NOT NULL,
    text TEXT NOT NULL,
    speaker TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (corpus_file_id) REFERENCES corpus_files(id) ON DELETE CASCADE
);

CREATE INDEX idx_corpus_segments_corpus_file_id ON corpus_segments(corpus_file_id);
CREATE INDEX idx_corpus_segments_time ON corpus_segments(start_time, end_time);
CREATE INDEX idx_corpus_segments_text ON corpus_segments(text);
CREATE UNIQUE INDEX idx_corpus_segments_file_index ON corpus_segments(corpus_file_id, segment_index);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_corpus_segments_file_index;
DROP INDEX IF EXISTS idx_corpus_segments_text;
DROP INDEX IF EXISTS idx_corpus_segments_time;
DROP INDEX IF EXISTS idx_corpus_segments_corpus_file_id;
DROP TABLE IF EXISTS corpus_segments;
-- +goose StatementEnd
