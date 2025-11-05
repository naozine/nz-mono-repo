package model

import "time"

// CorpusFile represents a corpus file entity
type CorpusFile struct {
	ID               int64     `db:"id" json:"id"`
	ProjectID        int64     `db:"project_id" json:"project_id"`
	AudioFileID      *int64    `db:"audio_file_id" json:"audio_file_id,omitempty"`
	Name             string    `db:"name" json:"name"`
	OriginalFilename string    `db:"original_filename" json:"original_filename"`
	FilePath         string    `db:"file_path" json:"file_path"`
	FileSize         int64     `db:"file_size" json:"file_size"`
	SegmentCount     int       `db:"segment_count" json:"segment_count"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
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
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
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
}

// WhisperXSegment represents a segment in WhisperX JSON format
type WhisperXSegment struct {
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
	Text    string  `json:"text"`
	Speaker *string `json:"speaker,omitempty"`
}

// WhisperXOutput represents the WhisperX JSON format
type WhisperXOutput struct {
	Segments []WhisperXSegment `json:"segments"`
}
