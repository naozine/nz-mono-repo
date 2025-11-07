-- +goose Up
-- +goose StatementBegin
CREATE TABLE audio_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    duration REAL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX idx_audio_files_project_id ON audio_files(project_id);
CREATE INDEX idx_audio_files_created_at ON audio_files(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_audio_files_created_at;
DROP INDEX IF EXISTS idx_audio_files_project_id;
DROP TABLE IF EXISTS audio_files;
-- +goose StatementEnd
