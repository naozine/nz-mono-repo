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

// GetAllWordsFromGroup retrieves all words from all corpus files in a group with speaker info
func (s *CorpusService) GetAllWordsFromGroup(groupID int64) ([]*model.WordWithSpeaker, error) {
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

	// Collect all words with speaker labels
	var allWords []*model.WordWithSpeaker
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
			words, err := seg.GetWords()
			if err != nil {
				continue // Skip segments with invalid word JSON
			}

			for _, word := range words {
				allWords = append(allWords, &model.WordWithSpeaker{
					Word:         word,
					Speaker:      speakerLabel,
					CorpusFileID: file.ID,
					AudioFileID:  file.AudioFileID,
				})
			}
		}
	}

	// Sort by start time
	sort.Slice(allWords, func(i, j int) bool {
		return allWords[i].Start < allWords[j].Start
	})

	return allWords, nil
}

// RefinedSegmentInput represents input for creating refined segments
type RefinedSegmentInput struct {
	CorpusFileID int64
	Speaker      string
	AudioFileID  *int64
	StartTime    float64
	EndTime      float64
	Text         string
	Words        []model.Word
}

// RefineSegmentsByWords re-segments corpus files based on word-level timing and speaker changes
func (s *CorpusService) RefineSegmentsByWords(groupID int64, minDuration float64, maxDuration float64, gapThreshold float64) ([]*RefinedSegmentInput, error) {
	if groupID < 1 {
		return nil, fmt.Errorf("invalid group id")
	}

	// Get all words from group
	allWords, err := s.GetAllWordsFromGroup(groupID)
	if err != nil {
		return nil, err
	}

	if len(allWords) == 0 {
		return nil, fmt.Errorf("no words found in group")
	}

	// Filter out words with abnormal duration
	validWords := filterWordsByDuration(allWords, minDuration, maxDuration)

	if len(validWords) == 0 {
		return nil, fmt.Errorf("no valid words after filtering")
	}

	// Re-segment based on speaker changes and gaps
	segments := splitIntoSegments(validWords, gapThreshold)

	return segments, nil
}

// filterWordsByDuration filters words with abnormal durations
func filterWordsByDuration(words []*model.WordWithSpeaker, minDuration float64, maxDuration float64) []*model.WordWithSpeaker {
	filtered := make([]*model.WordWithSpeaker, 0, len(words))

	for _, word := range words {
		duration := word.Duration()
		if duration >= minDuration && duration <= maxDuration {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// splitIntoSegments splits words into segments based on speaker changes and gaps
func splitIntoSegments(words []*model.WordWithSpeaker, gapThreshold float64) []*RefinedSegmentInput {
	if len(words) == 0 {
		return nil
	}

	segments := make([]*RefinedSegmentInput, 0)
	currentSegment := &RefinedSegmentInput{
		CorpusFileID: words[0].CorpusFileID,
		Speaker:      words[0].Speaker,
		AudioFileID:  words[0].AudioFileID,
		StartTime:    words[0].Start,
		Words:        []model.Word{words[0].Word},
	}

	for i := 1; i < len(words); i++ {
		word := words[i]
		prevWord := words[i-1]

		// Calculate gap between previous word and current word
		gap := word.Start - prevWord.End

		// Determine if we should start a new segment
		shouldStartNewSegment := false

		// Condition 1: Speaker changed
		if word.Speaker != prevWord.Speaker {
			shouldStartNewSegment = true
		}

		// Condition 2: Gap exceeds threshold
		if gap > gapThreshold {
			shouldStartNewSegment = true
		}

		if shouldStartNewSegment {
			// Finalize current segment
			currentSegment.EndTime = prevWord.End
			currentSegment.Text = buildTextFromWords(currentSegment.Words)
			segments = append(segments, currentSegment)

			// Start new segment
			currentSegment = &RefinedSegmentInput{
				CorpusFileID: word.CorpusFileID,
				Speaker:      word.Speaker,
				AudioFileID:  word.AudioFileID,
				StartTime:    word.Start,
				Words:        []model.Word{word.Word},
			}
		} else {
			// Add word to current segment
			currentSegment.Words = append(currentSegment.Words, word.Word)
		}
	}

	// Finalize last segment
	if len(currentSegment.Words) > 0 {
		lastWord := words[len(words)-1]
		currentSegment.EndTime = lastWord.End
		currentSegment.Text = buildTextFromWords(currentSegment.Words)
		segments = append(segments, currentSegment)
	}

	return segments
}

// buildTextFromWords concatenates words to form segment text
func buildTextFromWords(words []model.Word) string {
	if len(words) == 0 {
		return ""
	}

	text := ""
	for _, word := range words {
		text += word.Word
	}

	return text
}
