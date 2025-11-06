package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/whiscript/internal/model"
)

// CorpusRepository handles database operations for corpus files and segments
type CorpusRepository struct {
	db *sqlx.DB
}

// NewCorpusRepository creates a new corpus repository
func NewCorpusRepository(db *sqlx.DB) *CorpusRepository {
	return &CorpusRepository{db: db}
}

// ListFilesByProjectID retrieves all corpus files for a project
func (r *CorpusRepository) ListFilesByProjectID(projectID int64) ([]*model.CorpusFile, error) {
	var corpusFiles []*model.CorpusFile
	query := "SELECT * FROM corpus_files WHERE project_id = ? ORDER BY created_at DESC"
	if err := r.db.Select(&corpusFiles, query, projectID); err != nil {
		return nil, fmt.Errorf("failed to select corpus files: %w", err)
	}
	return corpusFiles, nil
}

// GetFileByID retrieves a corpus file by ID
func (r *CorpusRepository) GetFileByID(id int64) (*model.CorpusFile, error) {
	var corpusFile model.CorpusFile
	query := "SELECT * FROM corpus_files WHERE id = ?"
	if err := r.db.Get(&corpusFile, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("corpus file not found")
		}
		return nil, fmt.Errorf("failed to get corpus file: %w", err)
	}
	return &corpusFile, nil
}

// CreateFile creates a new corpus file
func (r *CorpusRepository) CreateFile(input *model.CorpusFileCreateInput) (*model.CorpusFile, error) {
	now := time.Now()
	query := `INSERT INTO corpus_files (project_id, audio_file_id, name, original_filename, file_path, file_size, segment_count, created_at, updated_at)
	          VALUES (:project_id, :audio_file_id, :name, :original_filename, :file_path, :file_size, :segment_count, :created_at, :updated_at)`

	args := map[string]interface{}{
		"project_id":        input.ProjectID,
		"audio_file_id":     input.AudioFileID,
		"name":              input.Name,
		"original_filename": input.OriginalFilename,
		"file_path":         input.FilePath,
		"file_size":         input.FileSize,
		"segment_count":     input.SegmentCount,
		"created_at":        now,
		"updated_at":        now,
	}

	result, err := r.db.NamedExec(query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to create corpus file: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetFileByID(id)
}

// DeleteFile deletes a corpus file by ID (cascade deletes segments)
func (r *CorpusRepository) DeleteFile(id int64) error {
	query := "DELETE FROM corpus_files WHERE id = ?"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete corpus file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("corpus file not found")
	}

	return nil
}

// CreateSegment creates a new corpus segment
func (r *CorpusRepository) CreateSegment(input *model.CorpusSegmentCreateInput) (*model.CorpusSegment, error) {
	now := time.Now()
	query := `INSERT INTO corpus_segments (corpus_file_id, segment_index, start_time, end_time, text, speaker, words_json, created_at, updated_at)
	          VALUES (:corpus_file_id, :segment_index, :start_time, :end_time, :text, :speaker, :words_json, :created_at, :updated_at)`

	args := map[string]interface{}{
		"corpus_file_id": input.CorpusFileID,
		"segment_index":  input.SegmentIndex,
		"start_time":     input.StartTime,
		"end_time":       input.EndTime,
		"text":           input.Text,
		"speaker":        input.Speaker,
		"words_json":     input.WordsJSON,
		"created_at":     now,
		"updated_at":     now,
	}

	result, err := r.db.NamedExec(query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to create corpus segment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetSegmentByID(id)
}

// CreateSegmentsBatch creates multiple corpus segments efficiently
func (r *CorpusRepository) CreateSegmentsBatch(inputs []*model.CorpusSegmentCreateInput) error {
	if len(inputs) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	query := `INSERT INTO corpus_segments (corpus_file_id, segment_index, start_time, end_time, text, speaker, words_json, created_at, updated_at)
	          VALUES (:corpus_file_id, :segment_index, :start_time, :end_time, :text, :speaker, :words_json, :created_at, :updated_at)`

	for _, input := range inputs {
		args := map[string]interface{}{
			"corpus_file_id": input.CorpusFileID,
			"segment_index":  input.SegmentIndex,
			"start_time":     input.StartTime,
			"end_time":       input.EndTime,
			"text":           input.Text,
			"speaker":        input.Speaker,
			"words_json":     input.WordsJSON,
			"created_at":     now,
			"updated_at":     now,
		}

		if _, err := tx.NamedExec(query, args); err != nil {
			return fmt.Errorf("failed to insert segment: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetSegmentByID retrieves a corpus segment by ID
func (r *CorpusRepository) GetSegmentByID(id int64) (*model.CorpusSegment, error) {
	var segment model.CorpusSegment
	query := "SELECT * FROM corpus_segments WHERE id = ?"
	if err := r.db.Get(&segment, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("corpus segment not found")
		}
		return nil, fmt.Errorf("failed to get corpus segment: %w", err)
	}
	return &segment, nil
}

// ListSegmentsByCorpusFileID retrieves all segments for a corpus file
func (r *CorpusRepository) ListSegmentsByCorpusFileID(corpusFileID int64) ([]*model.CorpusSegment, error) {
	var segments []*model.CorpusSegment
	query := "SELECT * FROM corpus_segments WHERE corpus_file_id = ? ORDER BY segment_index ASC"
	if err := r.db.Select(&segments, query, corpusFileID); err != nil {
		return nil, fmt.Errorf("failed to select corpus segments: %w", err)
	}
	return segments, nil
}

// UpdateSegmentText updates the text of a corpus segment
func (r *CorpusRepository) UpdateSegmentText(id int64, text string) error {
	query := "UPDATE corpus_segments SET text = ?, updated_at = ? WHERE id = ?"
	result, err := r.db.Exec(query, text, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update segment text: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("corpus segment not found")
	}

	return nil
}

// CreateGroup creates a new corpus file group
func (r *CorpusRepository) CreateGroup(input *model.CorpusFileGroupCreateInput) (*model.CorpusFileGroup, error) {
	now := time.Now()
	query := `INSERT INTO corpus_file_groups (project_id, name, created_at, updated_at)
	          VALUES (?, ?, ?, ?)`

	result, err := r.db.Exec(query, input.ProjectID, input.Name, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create corpus file group: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetGroupByID(id)
}

// GetGroupByID retrieves a corpus file group by ID
func (r *CorpusRepository) GetGroupByID(id int64) (*model.CorpusFileGroup, error) {
	var group model.CorpusFileGroup
	query := "SELECT * FROM corpus_file_groups WHERE id = ?"
	if err := r.db.Get(&group, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("corpus file group not found")
		}
		return nil, fmt.Errorf("failed to get corpus file group: %w", err)
	}
	return &group, nil
}

// ListGroupsByProjectID retrieves all corpus file groups for a project
func (r *CorpusRepository) ListGroupsByProjectID(projectID int64) ([]*model.CorpusFileGroup, error) {
	var groups []*model.CorpusFileGroup
	query := "SELECT * FROM corpus_file_groups WHERE project_id = ? ORDER BY created_at DESC"
	if err := r.db.Select(&groups, query, projectID); err != nil {
		return nil, fmt.Errorf("failed to select corpus file groups: %w", err)
	}
	return groups, nil
}

// UpdateFileGroup updates the group_id and speaker_label of a corpus file
func (r *CorpusRepository) UpdateFileGroup(fileID int64, groupID *int64, speakerLabel *string) error {
	query := "UPDATE corpus_files SET group_id = ?, speaker_label = ?, updated_at = ? WHERE id = ?"
	result, err := r.db.Exec(query, groupID, speakerLabel, time.Now(), fileID)
	if err != nil {
		return fmt.Errorf("failed to update corpus file group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("corpus file not found")
	}

	return nil
}

// ListFilesByGroupID retrieves all corpus files in a group
func (r *CorpusRepository) ListFilesByGroupID(groupID int64) ([]*model.CorpusFile, error) {
	var files []*model.CorpusFile
	query := "SELECT * FROM corpus_files WHERE group_id = ? ORDER BY speaker_label ASC"
	if err := r.db.Select(&files, query, groupID); err != nil {
		return nil, fmt.Errorf("failed to select corpus files by group: %w", err)
	}
	return files, nil
}
