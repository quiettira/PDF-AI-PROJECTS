package models

import "time"

type PdfFile struct {
	ID               int       `json:"id" db:"id"`
	Filename         string    `json:"filename" db:"filename"`
	OriginalFilename string    `json:"original_filename" db:"original_filename"`
	Filepath         string    `json:"filepath" db:"filepath"`
	Filesize         int64     `json:"filesize" db:"filesize"`
	UploadTime       time.Time `json:"upload_time" db:"upload_time"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	LatestSummary    string    `json:"latest_summary" db:"latest_summary"`
}

type HistoryItem struct {
	ID          int       `json:"id"`
	Filename    string    `json:"filename"`
	Filesize    int64     `json:"filesize"`
	Status      string    `json:"status"`
	UploadedAt  time.Time `json:"uploaded_at"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
	Summary     string    `json:"summary,omitempty"`
}

type HistoryResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    []HistoryItem `json:"data,omitempty"`
	Total   int           `json:"total,omitempty"`
}
//File ini tidak berisi logika, melainkan hanya definisi bentuk data agar pengolahan dan pengiriman data menjadi konsisten.