-- +goose Up
-- +goose StatementBegin
ALTER TABLE corpus_files ADD COLUMN group_id INTEGER;
ALTER TABLE corpus_files ADD COLUMN speaker_label TEXT;

CREATE INDEX idx_corpus_files_group_id ON corpus_files(group_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite doesn't support DROP COLUMN directly, need to recreate table
CREATE TABLE corpus_files_backup (
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

INSERT INTO corpus_files_backup (id, project_id, audio_file_id, name, original_filename, file_path, file_size, segment_count, created_at, updated_at)
SELECT id, project_id, audio_file_id, name, original_filename, file_path, file_size, segment_count, created_at, updated_at
FROM corpus_files;

DROP TABLE corpus_files;

ALTER TABLE corpus_files_backup RENAME TO corpus_files;

CREATE INDEX idx_corpus_files_project_id ON corpus_files(project_id);
CREATE INDEX idx_corpus_files_audio_file_id ON corpus_files(audio_file_id);
CREATE INDEX idx_corpus_files_created_at ON corpus_files(created_at);
-- +goose StatementEnd
