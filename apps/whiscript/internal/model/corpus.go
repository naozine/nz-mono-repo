package model

import (
	"encoding/json"
	"time"
)

// CorpusFile represents a corpus file entity
type CorpusFile struct {
	ID               int64     `db:"id" json:"id"`
	ProjectID        int64     `db:"project_id" json:"project_id"`
	AudioFileID      *int64    `db:"audio_file_id" json:"audio_file_id,omitempty"`
	GroupID          *int64    `db:"group_id" json:"group_id,omitempty"`
	SpeakerLabel     *string   `db:"speaker_label" json:"speaker_label,omitempty"`
	Name             string    `db:"name" json:"name"`
	OriginalFilename string    `db:"original_filename" json:"original_filename"`
	FilePath         string    `db:"file_path" json:"file_path"`
	FileSize         int64     `db:"file_size" json:"file_size"`
	SegmentCount     int       `db:"segment_count" json:"segment_count"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// CorpusFileGroup represents a group of corpus files for multi-speaker conversations
type CorpusFileGroup struct {
	ID        int64     `db:"id" json:"id"`
	ProjectID int64     `db:"project_id" json:"project_id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// CorpusSegment represents a corpus segment entity
type CorpusSegment struct {
	ID           int64     `db:"id" json:"id"`
	CorpusFileID int64     `db:"corpus_file_id" json:"corpus_file_id"`
	SegmentIndex int       `db:"segment_index" json:"segment_index"`
	StartTime    float64   `db:"start_time" json:"start_time"`
	EndTime      float64   `db:"end_time" json:"end_time"`
	Text         string    `db:"text" json:"text"`
	Speaker      *string   `db:"speaker" json:"speaker,omitempty"`
	WordsJSON    *string   `db:"words_json" json:"words_json,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// GetWords parses and returns the words array from JSON
func (c *CorpusSegment) GetWords() ([]Word, error) {
	if c.WordsJSON == nil || *c.WordsJSON == "" {
		return nil, nil
	}

	var words []Word
	if err := json.Unmarshal([]byte(*c.WordsJSON), &words); err != nil {
		return nil, err
	}
	return words, nil
}

// HasWords returns true if the segment has word-level timing data
func (c *CorpusSegment) HasWords() bool {
	return c.WordsJSON != nil && *c.WordsJSON != ""
}

// CorpusFileCreateInput represents input for creating a corpus file
type CorpusFileCreateInput struct {
	ProjectID        int64
	AudioFileID      *int64
	Name             string
	OriginalFilename string
	FilePath         string
	FileSize         int64
	SegmentCount     int
}

// CorpusSegmentCreateInput represents input for creating a corpus segment
type CorpusSegmentCreateInput struct {
	CorpusFileID int64
	SegmentIndex int
	StartTime    float64
	EndTime      float64
	Text         string
	Speaker      *string
	WordsJSON    *string
}

// Word represents a word timing information
type Word struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// WhisperXSegment represents a segment in WhisperX JSON format
type WhisperXSegment struct {
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
	Text    string  `json:"text"`
	Speaker *string `json:"speaker,omitempty"`
	Words   []Word  `json:"words,omitempty"`
}

// WhisperXOutput represents the WhisperX JSON format
type WhisperXOutput struct {
	Segments []WhisperXSegment `json:"segments"`
}

// SegmentWithGap represents a segment with optional gap information
type SegmentWithGap struct {
	Segment      *CorpusSegment
	GapAfter     *float64 // Gap duration in seconds after this segment
	SpeakerLabel *string  // Speaker label (for grouped segments)
	AudioFileID  *int64   // Audio file ID (for grouped segments)
}

// CorpusFileGroupCreateInput represents input for creating a corpus file group
type CorpusFileGroupCreateInput struct {
	ProjectID int64
	Name      string
}

// GroupedCorpusFile represents a corpus file with its group info
type GroupedCorpusFile struct {
	File  *CorpusFile
	Group *CorpusFileGroup
}

// MergedSegment represents a segment with speaker information for grouped display
type MergedSegment struct {
	Segment      *CorpusSegment
	SpeakerLabel string
	AudioFileID  *int64
}
