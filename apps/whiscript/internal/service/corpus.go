package service

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/whiscript/internal/model"
	"github.com/yourusername/whiscript/internal/repository"
)

// CorpusService handles business logic for corpus files
type CorpusService struct {
	repo       *repository.CorpusRepository
	uploadPath string
}

// NewCorpusService creates a new corpus service
func NewCorpusService(repo *repository.CorpusRepository, uploadPath string) *CorpusService {
	return &CorpusService{
		repo:       repo,
		uploadPath: uploadPath,
	}
}

// ListFilesByProjectID retrieves all corpus files for a project
func (s *CorpusService) ListFilesByProjectID(projectID int64) ([]*model.CorpusFile, error) {
	if projectID < 1 {
		return nil, fmt.Errorf("invalid project id")
	}
	return s.repo.ListFilesByProjectID(projectID)
}

// GetFileByID retrieves a corpus file by ID
func (s *CorpusService) GetFileByID(id int64) (*model.CorpusFile, error) {
	if id < 1 {
		return nil, fmt.Errorf("invalid corpus file id")
	}
	return s.repo.GetFileByID(id)
}

// GetSegmentsByCorpusFileID retrieves all segments for a corpus file
func (s *CorpusService) GetSegmentsByCorpusFileID(corpusFileID int64) ([]*model.CorpusSegment, error) {
	if corpusFileID < 1 {
		return nil, fmt.Errorf("invalid corpus file id")
	}
	return s.repo.ListSegmentsByCorpusFileID(corpusFileID)
}

// GetSegmentsWithGaps retrieves segments with gap information
func (s *CorpusService) GetSegmentsWithGaps(corpusFileID int64, gapThreshold float64) ([]*model.SegmentWithGap, error) {
	segments, err := s.GetSegmentsByCorpusFileID(corpusFileID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.SegmentWithGap, len(segments))
	for i, segment := range segments {
		segWithGap := &model.SegmentWithGap{
			Segment: segment,
		}

		// Calculate gap after this segment (if there's a next segment)
		if i < len(segments)-1 {
			nextSegment := segments[i+1]
			gap := nextSegment.StartTime - segment.EndTime
			if gap >= gapThreshold {
				segWithGap.GapAfter = &gap
			}
		}

		result[i] = segWithGap
	}

	return result, nil
}

// Upload handles corpus JSON file upload
func (s *CorpusService) Upload(projectID int64, audioFileID *int64, file *multipart.FileHeader) (*model.CorpusFile, error) {
	if projectID < 1 {
		return nil, fmt.Errorf("invalid project id")
	}

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".json") {
		return nil, fmt.Errorf("invalid file type: only JSON files are allowed")
	}

	// Validate file size (max 50MB)
	if file.Size > 50*1024*1024 {
		return nil, fmt.Errorf("file too large: maximum size is 50MB")
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read file content
	content, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Parse JSON to validate and extract segments
	var whisperXOutput model.WhisperXOutput
	if err := json.Unmarshal(content, &whisperXOutput); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	if len(whisperXOutput.Segments) == 0 {
		return nil, fmt.Errorf("no segments found in JSON file")
	}

	// Generate unique filename for storage
	ext := filepath.Ext(file.Filename)
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%d%s", projectID, timestamp, ext)
	filePath := filepath.Join(s.uploadPath, "corpus", filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Save JSON file
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create corpus file record
	corpusFileInput := &model.CorpusFileCreateInput{
		ProjectID:        projectID,
		AudioFileID:      audioFileID,
		Name:             strings.TrimSuffix(file.Filename, ext),
		OriginalFilename: file.Filename,
		FilePath:         filePath,
		FileSize:         file.Size,
		SegmentCount:     len(whisperXOutput.Segments),
	}

	corpusFile, err := s.repo.CreateFile(corpusFileInput)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to create corpus file record: %w", err)
	}

	// Create segment records in batch
	segmentInputs := make([]*model.CorpusSegmentCreateInput, len(whisperXOutput.Segments))
	for i, segment := range whisperXOutput.Segments {
		segmentInputs[i] = &model.CorpusSegmentCreateInput{
			CorpusFileID: corpusFile.ID,
			SegmentIndex: i,
			StartTime:    segment.Start,
			EndTime:      segment.End,
			Text:         segment.Text,
			Speaker:      segment.Speaker,
		}
	}

	if err := s.repo.CreateSegmentsBatch(segmentInputs); err != nil {
		// Clean up corpus file record and physical file
		s.repo.DeleteFile(corpusFile.ID)
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create segments: %w", err)
	}

	return corpusFile, nil
}

// DeleteFile deletes a corpus file by ID
func (s *CorpusService) DeleteFile(id int64) error {
	if id < 1 {
		return fmt.Errorf("invalid corpus file id")
	}

	// Get corpus file to retrieve file path
	corpusFile, err := s.repo.GetFileByID(id)
	if err != nil {
		return err
	}

	// Delete from database (cascade deletes segments)
	if err := s.repo.DeleteFile(id); err != nil {
		return err
	}

	// Delete physical file
	if err := os.Remove(corpusFile.FilePath); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to delete physical file %s: %v\n", corpusFile.FilePath, err)
	}

	return nil
}

// UpdateSegmentText updates the text of a corpus segment
func (s *CorpusService) UpdateSegmentText(segmentID int64, text string) error {
	if segmentID < 1 {
		return fmt.Errorf("invalid segment id")
	}

	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("segment text cannot be empty")
	}

	return s.repo.UpdateSegmentText(segmentID, text)
}
