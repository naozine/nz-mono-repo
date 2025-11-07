-- +goose Up
-- +goose StatementBegin
ALTER TABLE corpus_segments ADD COLUMN words_json TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite doesn't support DROP COLUMN directly, need to recreate table
CREATE TABLE corpus_segments_backup (
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

INSERT INTO corpus_segments_backup (id, corpus_file_id, segment_index, start_time, end_time, text, speaker, created_at, updated_at)
SELECT id, corpus_file_id, segment_index, start_time, end_time, text, speaker, created_at, updated_at
FROM corpus_segments;

DROP TABLE corpus_segments;

ALTER TABLE corpus_segments_backup RENAME TO corpus_segments;

CREATE INDEX idx_corpus_segments_corpus_file_id ON corpus_segments(corpus_file_id);
CREATE INDEX idx_corpus_segments_time ON corpus_segments(start_time, end_time);
CREATE INDEX idx_corpus_segments_text ON corpus_segments(text);
CREATE UNIQUE INDEX idx_corpus_segments_file_index ON corpus_segments(corpus_file_id, segment_index);
-- +goose StatementEnd
