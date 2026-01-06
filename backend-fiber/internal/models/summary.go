package models

import "time"

type Summary struct { 
	ID               int       `json:"id" db:"id"`
	PdfID            int       `json:"pdf_id" db:"pdf_id"`
	SummaryText      string    `json:"summary_text" db:"summary_text"`
	SummaryStyle     string    `json:"summary_style" db:"summary_style"`
	ProcessTimeMs    int64     `json:"process_time_ms" db:"process_time_ms"`
	LanguageDetected string    `json:"language_detected" db:"language_detected"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

type SummaryResponse struct {
	PdfID     int       `json:"pdf_id"`
	Summaries []Summary `json:"summaries"`
	Count     int       `json:"count"`
}