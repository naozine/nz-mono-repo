package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/whiscript/internal/model"
)

// AudioFileRepository handles database operations for audio files
type AudioFileRepository struct {
	db *sqlx.DB
}

// NewAudioFileRepository creates a new audio file repository
func NewAudioFileRepository(db *sqlx.DB) *AudioFileRepository {
	return &AudioFileRepository{db: db}
}

// ListByProjectID retrieves all audio files for a project
func (r *AudioFileRepository) ListByProjectID(projectID int64) ([]*model.AudioFile, error) {
	var audioFiles []*model.AudioFile
	query := "SELECT * FROM audio_files WHERE project_id = ? ORDER BY created_at DESC"
	if err := r.db.Select(&audioFiles, query, projectID); err != nil {
		return nil, fmt.Errorf("failed to select audio files: %w", err)
	}
	return audioFiles, nil
}

// GetByID retrieves an audio file by ID
func (r *AudioFileRepository) GetByID(id int64) (*model.AudioFile, error) {
	var audioFile model.AudioFile
	query := "SELECT * FROM audio_files WHERE id = ?"
	if err := r.db.Get(&audioFile, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audio file not found")
		}
		return nil, fmt.Errorf("failed to get audio file: %w", err)
	}
	return &audioFile, nil
}

// Create creates a new audio file
func (r *AudioFileRepository) Create(input *model.AudioFileCreateInput) (*model.AudioFile, error) {
	now := time.Now()
	query := `INSERT INTO audio_files (project_id, name, original_filename, file_path, file_size, mime_type, duration, created_at, updated_at)
	          VALUES (:project_id, :name, :original_filename, :file_path, :file_size, :mime_type, :duration, :created_at, :updated_at)`

	args := map[string]interface{}{
		"project_id":        input.ProjectID,
		"name":              input.Name,
		"original_filename": input.OriginalFilename,
		"file_path":         input.FilePath,
		"file_size":         input.FileSize,
		"mime_type":         input.MimeType,
		"duration":          input.Duration,
		"created_at":        now,
		"updated_at":        now,
	}

	result, err := r.db.NamedExec(query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio file: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetByID(id)
}

// Delete deletes an audio file by ID
func (r *AudioFileRepository) Delete(id int64) error {
	query := "DELETE FROM audio_files WHERE id = ?"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete audio file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("audio file not found")
	}

	return nil
}
