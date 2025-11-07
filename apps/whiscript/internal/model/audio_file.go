package model

import "time"

// AudioFile represents an audio file entity
type AudioFile struct {
	ID               int64     `db:"id" json:"id"`
	ProjectID        int64     `db:"project_id" json:"project_id"`
	Name             string    `db:"name" json:"name"`
	OriginalFilename string    `db:"original_filename" json:"original_filename"`
	FilePath         string    `db:"file_path" json:"file_path"`
	FileSize         int64     `db:"file_size" json:"file_size"`
	MimeType         string    `db:"mime_type" json:"mime_type"`
	Duration         *float64  `db:"duration" json:"duration,omitempty"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// AudioFileCreateInput represents input for creating an audio file
type AudioFileCreateInput struct {
	ProjectID        int64
	Name             string
	OriginalFilename string
	FilePath         string
	FileSize         int64
	MimeType         string
	Duration         *float64
}
