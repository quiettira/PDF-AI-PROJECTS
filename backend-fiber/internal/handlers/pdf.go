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

func getJakartaLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Local
	}
	return loc
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

	jakartaLoc := getJakartaLocation()

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

	uploadTime = uploadTime.In(jakartaLoc)
	createdAt = createdAt.In(jakartaLoc)

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
			"created_at":        summaryCreatedAt.In(jakartaLoc),
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

func (h *PdfHandler) GetHistory(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10) //brp data perhalaman utk pagination nih
	offset := c.QueryInt("offset", 0) //mulai dr data ke brp

	jakartaLoc := getJakartaLocation()

	// Get total count buat paginasi
	var totalCount int
	err := h.DB.QueryRow(`SELECT COUNT(*) FROM pdf_files`).Scan(&totalCount)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghitung total data"})
	} //frontend perlu ta totalnya buat pagination

	// Single query with latest_summary - NO MORE JOIN!
	rows, err := h.DB.Query(`
		SELECT id, filename, filesize, created_at, COALESCE(latest_summary, '') as latest_summary
		FROM pdf_files
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal ambil data"})
	}
	defer rows.Close()

	var history []models.HistoryItem
	for rows.Next() { //1 baris 1 pdf
		var item models.HistoryItem
		var latestSummary string

		if err := rows.Scan(&item.ID, &item.Filename, &item.Filesize, &item.UploadedAt, &latestSummary); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal parsing data"})
		}

		item.Status = "completed"
		item.Summary = latestSummary
		item.UploadedAt = item.UploadedAt.In(jakartaLoc)

		// Set ProcessedAt if summary exists
		if latestSummary != "" {
			item.ProcessedAt = time.Now().In(jakartaLoc)
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
	jakartaLoc := getJakartaLocation()

	rows, err := h.DB.Query(`
		SELECT id, filename, COALESCE(original_filename, filename) as original_filename, 
		       filesize, upload_time, COALESCE(latest_summary, '') as latest_summary
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
		var filename, originalFilename, latestSummary string
		var filesize int64
		var uploadTime time.Time

		if err := rows.Scan(&id, &filename, &originalFilename, &filesize, &uploadTime, &latestSummary); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Scan Error: %v", err)})
		}

		uploadTime = uploadTime.In(jakartaLoc)

		pdfs = append(pdfs, map[string]interface{}{
			"id":                id,
			"filename":          filename,
			"original_filename": originalFilename,
			"filesize":          filesize,
			"upload_time":       uploadTime,
			"latest_summary":    latestSummary,
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

	jakartaLoc := getJakartaLocation()

	var id int
	var filename, originalFilename, fp, latestSummary string
	var filesize int64
	var uploadTime time.Time

	err = h.DB.QueryRow(`
		SELECT id, filename, COALESCE(original_filename, filename) as original_filename, 
		       filepath, filesize, upload_time, COALESCE(latest_summary, '') as latest_summary
		FROM pdf_files WHERE id = $1
	`, pdfID).Scan(&id, &filename, &originalFilename, &fp, &filesize, &uploadTime, &latestSummary)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "PDF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Database error: %v", err)})
	}

	uploadTime = uploadTime.In(jakartaLoc)

	return c.JSON(fiber.Map{
		"id":                id,
		"filename":          filename,
		"original_filename": originalFilename,
		"filepath":          fp,
		"filesize":          filesize,
		"upload_time":       uploadTime,
		"latest_summary":    latestSummary,
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

func (h *PdfHandler) GetSummaries(c *fiber.Ctx) error {
	pdfID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid PDF ID"})
	}

	jakartaLoc := getJakartaLocation()

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
			"created_at":        createdAt.In(jakartaLoc),
		})
	}

	return c.JSON(fiber.Map{
		"pdf_id":    pdfID,
		"summaries": summaries,
		"count":     len(summaries),
	})
}

//logika utama pdf sumarizernya daari crud,dll
