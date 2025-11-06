package service

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
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
		// Serialize words to JSON if present
		var wordsJSON *string
		if len(segment.Words) > 0 {
			wordsBytes, err := json.Marshal(segment.Words)
			if err != nil {
				s.repo.DeleteFile(corpusFile.ID)
				os.Remove(filePath)
				return nil, fmt.Errorf("failed to serialize words: %w", err)
			}
			wordsJSONStr := string(wordsBytes)
			wordsJSON = &wordsJSONStr
		}

		segmentInputs[i] = &model.CorpusSegmentCreateInput{
			CorpusFileID: corpusFile.ID,
			SegmentIndex: i,
			StartTime:    segment.Start,
			EndTime:      segment.End,
			Text:         segment.Text,
			Speaker:      segment.Speaker,
			WordsJSON:    wordsJSON,
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

// CreateGroup creates a new corpus file group and associates files with it
func (s *CorpusService) CreateGroup(projectID int64, name string, fileIDs []int64, speakerLabels []string) (*model.CorpusFileGroup, error) {
	if projectID < 1 {
		return nil, fmt.Errorf("invalid project id")
	}

	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("group name cannot be empty")
	}

	if len(fileIDs) != len(speakerLabels) {
		return nil, fmt.Errorf("file IDs and speaker labels must have the same length")
	}

	// Create group
	input := &model.CorpusFileGroupCreateInput{
		ProjectID: projectID,
		Name:      name,
	}

	group, err := s.repo.CreateGroup(input)
	if err != nil {
		return nil, err
	}

	// Associate files with group
	for i, fileID := range fileIDs {
		label := speakerLabels[i]
		if err := s.repo.UpdateFileGroup(fileID, &group.ID, &label); err != nil {
			return nil, fmt.Errorf("failed to associate file %d: %w", fileID, err)
		}
	}

	return group, nil
}

// GetGroupByID retrieves a corpus file group by ID
func (s *CorpusService) GetGroupByID(id int64) (*model.CorpusFileGroup, error) {
	if id < 1 {
		return nil, fmt.Errorf("invalid group id")
	}
	return s.repo.GetGroupByID(id)
}

// ListGroupsByProjectID retrieves all corpus file groups for a project
func (s *CorpusService) ListGroupsByProjectID(projectID int64) ([]*model.CorpusFileGroup, error) {
	if projectID < 1 {
		return nil, fmt.Errorf("invalid project id")
	}
	return s.repo.ListGroupsByProjectID(projectID)
}

// GetMergedSegments retrieves and merges segments from all files in a group, sorted by time
func (s *CorpusService) GetMergedSegments(groupID int64, gapThreshold float64) ([]*model.SegmentWithGap, error) {
	if groupID < 1 {
		return nil, fmt.Errorf("invalid group id")
	}

	// Get all files in the group
	files, err := s.repo.ListFilesByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files in group")
	}

	// Collect all segments with speaker labels
	var allSegments []*model.MergedSegment
	for _, file := range files {
		segments, err := s.repo.ListSegmentsByCorpusFileID(file.ID)
		if err != nil {
			return nil, err
		}

		speakerLabel := "Unknown"
		if file.SpeakerLabel != nil {
			speakerLabel = *file.SpeakerLabel
		}

		for _, seg := range segments {
			allSegments = append(allSegments, &model.MergedSegment{
				Segment:      seg,
				SpeakerLabel: speakerLabel,
				AudioFileID:  file.AudioFileID,
			})
		}
	}

	// Sort by start time
	sort.Slice(allSegments, func(i, j int) bool {
		return allSegments[i].Segment.StartTime < allSegments[j].Segment.StartTime
	})

	// Calculate gaps
	result := make([]*model.SegmentWithGap, len(allSegments))
	for i, mergedSeg := range allSegments {
		segWithGap := &model.SegmentWithGap{
			Segment:      mergedSeg.Segment,
			SpeakerLabel: &mergedSeg.SpeakerLabel,
			AudioFileID:  mergedSeg.AudioFileID,
		}

		// Calculate gap after this segment (if there's a next segment)
		if i < len(allSegments)-1 {
			nextSeg := allSegments[i+1]
			gap := nextSeg.Segment.StartTime - mergedSeg.Segment.EndTime
			if gap >= gapThreshold {
				segWithGap.GapAfter = &gap
			}
		}

		result[i] = segWithGap
	}

	return result, nil
}

// GetGroupFiles retrieves all files in a group with their audio file info
func (s *CorpusService) GetGroupFiles(groupID int64) ([]*model.CorpusFile, error) {
	if groupID < 1 {
		return nil, fmt.Errorf("invalid group id")
	}
	return s.repo.ListFilesByGroupID(groupID)
}
