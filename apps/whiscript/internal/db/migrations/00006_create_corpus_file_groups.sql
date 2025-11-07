-- +goose Up
-- +goose StatementBegin
CREATE TABLE corpus_file_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX idx_corpus_file_groups_project_id ON corpus_file_groups(project_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_corpus_file_groups_project_id;
DROP TABLE IF EXISTS corpus_file_groups;
-- +goose StatementEnd
