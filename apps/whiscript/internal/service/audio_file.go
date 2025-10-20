package service

import (
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

// AudioFileService handles business logic for audio files
type AudioFileService struct {
	repo       *repository.AudioFileRepository
	uploadPath string
}

// NewAudioFileService creates a new audio file service
func NewAudioFileService(repo *repository.AudioFileRepository, uploadPath string) *AudioFileService {
	return &AudioFileService{
		repo:       repo,
		uploadPath: uploadPath,
	}
}

// ListByProjectID retrieves all audio files for a project
func (s *AudioFileService) ListByProjectID(projectID int64) ([]*model.AudioFile, error) {
	if projectID < 1 {
		return nil, fmt.Errorf("invalid project id")
	}
	return s.repo.ListByProjectID(projectID)
}

// GetByID retrieves an audio file by ID
func (s *AudioFileService) GetByID(id int64) (*model.AudioFile, error) {
	if id < 1 {
		return nil, fmt.Errorf("invalid audio file id")
	}
	return s.repo.GetByID(id)
}

// Upload handles audio file upload
func (s *AudioFileService) Upload(projectID int64, file *multipart.FileHeader) (*model.AudioFile, error) {
	if projectID < 1 {
		return nil, fmt.Errorf("invalid project id")
	}

	// Validate file type
	mimeType := file.Header.Get("Content-Type")
	if !isValidAudioType(mimeType) {
		return nil, fmt.Errorf("invalid file type: only mp3 and wav files are allowed")
	}

	// Validate file size (max 100MB)
	if file.Size > 100*1024*1024 {
		return nil, fmt.Errorf("file too large: maximum size is 100MB")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%d%s", projectID, timestamp, ext)
	filePath := filepath.Join(s.uploadPath, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(s.uploadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create database record
	input := &model.AudioFileCreateInput{
		ProjectID:        projectID,
		Name:             strings.TrimSuffix(file.Filename, ext),
		OriginalFilename: file.Filename,
		FilePath:         filePath,
		FileSize:         file.Size,
		MimeType:         mimeType,
	}

	audioFile, err := s.repo.Create(input)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, err
	}

	return audioFile, nil
}

// Delete deletes an audio file by ID
func (s *AudioFileService) Delete(id int64) error {
	if id < 1 {
		return fmt.Errorf("invalid audio file id")
	}

	// Get audio file to retrieve file path
	audioFile, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	// Delete physical file
	if err := os.Remove(audioFile.FilePath); err != nil {
		// Log error but don't fail the operation
		// The database record is already deleted
		fmt.Printf("Warning: failed to delete physical file %s: %v\n", audioFile.FilePath, err)
	}

	return nil
}

// isValidAudioType checks if the MIME type is valid
func isValidAudioType(mimeType string) bool {
	validTypes := map[string]bool{
		"audio/mpeg":  true, // mp3
		"audio/mp3":   true, // mp3
		"audio/wav":   true, // wav
		"audio/wave":  true, // wav
		"audio/x-wav": true, // wav
	}
	return validTypes[mimeType]
}
