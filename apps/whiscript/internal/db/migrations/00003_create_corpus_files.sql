-- +goose Up
-- +goose StatementBegin
CREATE TABLE corpus_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    audio_file_id INTEGER,
    name TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    segment_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (audio_file_id) REFERENCES audio_files(id) ON DELETE SET NULL
);

CREATE INDEX idx_corpus_files_project_id ON corpus_files(project_id);
CREATE INDEX idx_corpus_files_audio_file_id ON corpus_files(audio_file_id);
CREATE INDEX idx_corpus_files_created_at ON corpus_files(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_corpus_files_created_at;
DROP INDEX IF EXISTS idx_corpus_files_audio_file_id;
DROP INDEX IF EXISTS idx_corpus_files_project_id;
DROP TABLE IF EXISTS corpus_files;
-- +goose StatementEnd
