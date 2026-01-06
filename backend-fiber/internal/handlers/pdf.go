package handlers

import (
	"database/sql"
	"encoding/json" //untuk parsing response pyhton
	"fmt" 
	"os"
	"strconv"
	"time" 

	"pdf-backend-fiber/internal/config"
	"pdf-backend-fiber/internal/models"
	"pdf-backend-fiber/internal/services"

	"github.com/gofiber/fiber/v2"
)

type PdfHandler struct { //menyimpan semua kebutuhan handlerpdf
	DB     *sql.DB
	Config config.Config
	Python *services.PythonClient
}

func NewPdfHandler(db *sql.DB, cfg config.Config) *PdfHandler {
	return &PdfHandler{
		DB:     db,
		Config: cfg,
		Python: services.NewPythonClient(cfg.PythonAPI),
	}
} //inisialisasi piton client, dipanggilnya di routes

func (h *PdfHandler) GetPDF(c *fiber.Ctx) error {
	pdfID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid PDF ID"})
	}

	var id int
	var filename, originalFilename, fp string
	var filesize int64
	var uploadTime, createdAt time.Time

	err = h.DB.QueryRow(`
		SELECT id, filename, original_filename, filepath, filesize, upload_time, created_at
		FROM pdf_files WHERE id = $1
	`, pdfID).Scan(&id, &filename, &originalFilename, &fp, &filesize, &uploadTime, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "PDF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// Get summaries
	summaryRows, err := h.DB.Query(`
		SELECT summary_text, summary_style, process_time_ms, language_detected, created_at
		FROM summaries WHERE pdf_id = $1 ORDER BY created_at DESC
	`, pdfID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get summaries"})
	}
	defer summaryRows.Close()

	var summaries []map[string]interface{}
	for summaryRows.Next() {
		var summaryText, summaryStyle, languageDetected string
		var processTimeMs int64
		var summaryCreatedAt time.Time

		if err := summaryRows.Scan(&summaryText, &summaryStyle, &processTimeMs, &languageDetected, &summaryCreatedAt); err != nil {
			continue
		}

		summaries = append(summaries, map[string]interface{}{
			"text":              summaryText,
			"style":             summaryStyle,
			"process_time_ms":   processTimeMs,
			"language_detected": languageDetected,
			"created_at":        summaryCreatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"id":                id,
		"filename":          filename,
		"original_filename": originalFilename,
		"filesize":          filesize,
		"upload_time":       uploadTime,
		"created_at":        createdAt,
		"summaries":         summaries,
	})
}

func (h *PdfHandler) DeletePDF(c *fiber.Ctx) error {
	pdfID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid PDF ID"})
	}

	var fp string
	err = h.DB.QueryRow("SELECT filepath FROM pdf_files WHERE id = $1", pdfID).Scan(&fp)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "PDF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	_, err = h.DB.Exec("DELETE FROM pdf_files WHERE id = $1", pdfID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete from database"})
	}

	if err := os.Remove(fp); err != nil {
		fmt.Printf("Warning: Failed to delete file %s: %v\n", fp, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "PDF deleted successfully",
	})
}

func (h *PdfHandler) DownloadPDF(c *fiber.Ctx) error {
	pdfID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid PDF ID"})
	}

	var fp, originalFilename string
	err = h.DB.QueryRow("SELECT filepath, original_filename FROM pdf_files WHERE id = $1", pdfID).Scan(&fp, &originalFilename)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "PDF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{"error": "File not found on disk"})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", originalFilename))
	return c.SendFile(fp)
}

func (h *PdfHandler) GetHistory(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	// Get total count
	var totalCount int
	err := h.DB.QueryRow(`SELECT COUNT(*) FROM pdf_files`).Scan(&totalCount)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghitung total data"})
	}

	rows, err := h.DB.Query(`
		SELECT id, filename, filesize, created_at
		FROM pdf_files
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal ambil data"})
	}
	defer rows.Close()

	var history []models.HistoryItem
	for rows.Next() {
		var item models.HistoryItem
		if err := rows.Scan(&item.ID, &item.Filename, &item.Filesize, &item.UploadedAt); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal parsing data"})
		}

		item.Status = "completed"

		// Get summary if exists
		var sum sql.NullString
		var ptms sql.NullInt64
		if err := h.DB.QueryRow(`
			SELECT summary_text, process_time_ms
			FROM summaries
			WHERE pdf_id = $1
			ORDER BY id DESC
			LIMIT 1
		`, item.ID).Scan(&sum, &ptms); err == nil {
			if sum.Valid {
				item.Summary = sum.String
			}
			if ptms.Valid {
				item.ProcessedAt = time.Now()
			}
		}

		history = append(history, item)
	}

	response := models.HistoryResponse{
		Success: true,
		Message: "Data berhasil diambil",
		Data:    history,
		Total:   totalCount,
	}

	return c.JSON(response)
}

func (h *PdfHandler) SimplePDFs(c *fiber.Ctx) error {
	rows, err := h.DB.Query(`
		SELECT id, filename, COALESCE(original_filename, filename) as original_filename, filesize, upload_time
		FROM pdf_files
		ORDER BY id DESC
		LIMIT 20
	`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Query Error: %v", err)})
	}
	defer rows.Close()

	var pdfs []map[string]interface{}
	for rows.Next() {
		var id int
		var filename, originalFilename string
		var filesize int64
		var uploadTime time.Time

		if err := rows.Scan(&id, &filename, &originalFilename, &filesize, &uploadTime); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Scan Error: %v", err)})
		}

		pdfs = append(pdfs, map[string]interface{}{
			"id":                id,
			"filename":          filename,
			"original_filename": originalFilename,
			"filesize":          filesize,
			"upload_time":       uploadTime,
		})
	}

	return c.JSON(fiber.Map{
		"pdfs":  pdfs,
		"count": len(pdfs),
	})
}

func (h *PdfHandler) SimplePDFByID(c *fiber.Ctx) error {
	pdfID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid PDF ID"})
	}

	var id int
	var filename, originalFilename, fp string
	var filesize int64
	var uploadTime time.Time

	err = h.DB.QueryRow(`
		SELECT id, filename, COALESCE(original_filename, filename) as original_filename, filepath, filesize, upload_time
		FROM pdf_files WHERE id = $1
	`, pdfID).Scan(&id, &filename, &originalFilename, &fp, &filesize, &uploadTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "PDF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Database error: %v", err)})
	}

	return c.JSON(fiber.Map{
		"id":                id,
		"filename":          filename,
		"original_filename": originalFilename,
		"filepath":          fp,
		"filesize":          filesize,
		"upload_time":       uploadTime,
	})
}

func (h *PdfHandler) UpdatePDF(c *fiber.Ctx) error {
	pdfID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid PDF ID"})
	}

	var updateData struct {
		OriginalFilename string `json:"original_filename,omitempty"`
		Description      string `json:"description,omitempty"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	var exists bool
	err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM pdf_files WHERE id = $1)", pdfID).Scan(&exists)
	if err != nil || !exists {
		return c.Status(404).JSON(fiber.Map{"error": "PDF not found"})
	}

	if updateData.OriginalFilename != "" {
		_, err = h.DB.Exec("UPDATE pdf_files SET original_filename = $1 WHERE id = $2",
			updateData.OriginalFilename, pdfID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Update failed: %v", err)})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "PDF metadata updated successfully",
		"pdf_id":  pdfID,
	})
}

func (h *PdfHandler) Resummarize(c *fiber.Ctx) error {
	pdfID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid PDF ID"})
	}

	var requestData struct {
		Style string `json:"style"`
	}
	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	if requestData.Style == "" {
		requestData.Style = "standard"
	}

	var fp string
	err = h.DB.QueryRow("SELECT filepath FROM pdf_files WHERE id = $1", pdfID).Scan(&fp)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "PDF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	summaryResult, duration, err := h.Python.Summarize(fp, requestData.Style)
	summaryText := ""
	language := "unknown"

	if err != nil {
		summaryText = "Re-summarization failed - Python service error"
		duration = 0
	} else {
		var response struct {
			Summary  string `json:"summary"`
			Language string `json:"detected_language"`
			Style    string `json:"style"`
		}
		if err := json.Unmarshal([]byte(summaryResult), &response); err == nil {
			summaryText = response.Summary
			language = response.Language
		} else {
			summaryText = summaryResult
		}
	}

	_, err = h.DB.Exec(
		`INSERT INTO summaries (pdf_id, summary_text, summary_style, process_time_ms, language_detected)
		 VALUES ($1, $2, $3, $4, $5)`,
		pdfID,
		summaryText,
		requestData.Style,
		duration,
		language,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save new summary"})
	}

	return c.JSON(fiber.Map{
		"success":         true,
		"message":         "PDF re-summarized successfully",
		"pdf_id":          pdfID,
		"new_summary":     summaryText,
		"style":           requestData.Style,
		"language":        language,
		"process_time_ms": duration,
	})
}

func (h *PdfHandler) UpdateSummary(c *fiber.Ctx) error {
	summaryID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Summary ID"})
	}

	var updateData struct {
		SummaryText string `json:"summary_text"`
		Style       string `json:"style,omitempty"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if updateData.SummaryText == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Summary text is required"})
	}

	result, err := h.DB.Exec(
		`UPDATE summaries SET summary_text = $1, summary_style = COALESCE($2, summary_style)
		 WHERE id = $3`,
		updateData.SummaryText,
		updateData.Style,
		summaryID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Update failed: %v", err)})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Summary not found"})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"message":    "Summary updated successfully",
		"summary_id": summaryID,
	})
}

func (h *PdfHandler) GetSummaries(c *fiber.Ctx) error {
	pdfID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid PDF ID"})
	}

	rows, err := h.DB.Query(`
		SELECT id, summary_text, summary_style, process_time_ms, language_detected, created_at
		FROM summaries
		WHERE pdf_id = $1
		ORDER BY created_at DESC
	`, pdfID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Query error: %v", err)})
	}
	defer rows.Close()

	var summaries []map[string]interface{}
	for rows.Next() {
		var id int
		var summaryText, summaryStyle, languageDetected string
		var processTimeMs int64
		var createdAt time.Time

		if err := rows.Scan(&id, &summaryText, &summaryStyle, &processTimeMs, &languageDetected, &createdAt); err != nil {
			continue
		}

		summaries = append(summaries, map[string]interface{}{
			"id":                id,
			"summary_text":      summaryText,
			"summary_style":     summaryStyle,
			"process_time_ms":   processTimeMs,
			"language_detected": languageDetected,
			"created_at":        createdAt,
		})
	}

	return c.JSON(fiber.Map{
		"pdf_id":    pdfID,
		"summaries": summaries,
		"count":     len(summaries),
	})
}
//logika utama pdf sumarizernya daari crud,dll