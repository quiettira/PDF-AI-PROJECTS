package handlers

import (
	"database/sql"
	"encoding/json" //persing response pyhton
	"fmt" //format nama file d pesan
	"log"
	"os"
	"path/filepath"
	"strings"
	"time" //bikin namafile unik katanya

	"pdf-backend-fiber/internal/config" //disini dia utk max size d folder upload
	"pdf-backend-fiber/internal/services"

	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct { 
	DB     *sql.DB
	Config config.Config
	Python *services.PythonClient
}

func NewUploadHandler(db *sql.DB, cfg config.Config) *UploadHandler {
	return &UploadHandler{
		DB:     db,
		Config: cfg,
		Python: services.NewPythonClient(cfg.PythonAPI),
	}
}

func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	style := c.Query("style", "standard")
	style = strings.ToLower(strings.TrimSpace(style))
	
	switch style {
	case "standard", "executive", "bullets", "detailed":
	default:
		style = "standard"
	}

	file, err := c.FormFile("file") //ngambil file dari form-data, field hrs namanya file
	if err != nil {
		return c.Status(400).JSON(fiber.Map{ //jika error, kembalikan error
			"error": "File tidak ditemukan",
		})
	}

	if filepath.Ext(file.Filename) != ".pdf" { //jika file bukan pdf
		return c.Status(400).JSON(fiber.Map{
			"error": "Hanya file PDF yang diizinkan",
		})
	}

	if file.Size > h.Config.MaxFileSize { //harus sesuai config yaitu 10 mb 
		return c.Status(400).JSON(fiber.Map{
			"error": "File terlalu besar (maks 10MB)",
		})
	}

	// Ensure upload directory exists
	_ = os.MkdirAll(h.Config.UploadDir, os.ModePerm) //tentang upload dir ada atau tidak

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename) //bikin namafile unik katanya
	savePath := filepath.Join(h.Config.UploadDir, filename) //simpan file di upload dir

	if err := c.SaveFile(file, savePath); err != nil { //simpan file ke disk
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal menyimpan file",
		})
	}

	// Save to database
	var pdfID int
	err = h.DB.QueryRow(
		`INSERT INTO pdf_files (filename, original_filename, filepath, filesize, upload_time)
		 VALUES ($1, $2, $3, $4, NOW()) RETURNING id`, //returning itu ambil id pdf yg baru dibuat
		filename,
		file.Filename,
		savePath,
		file.Size,
	).Scan(&pdfID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal simpan metadata PDF",
		})
	}

	// Summarize with Python service
	summaryResult, duration, err := h.Python.Summarize(savePath, style)
	summaryText := ""
	language := "unknown"

	if err != nil { //jika error
		log.Printf("Python service failed: %v, using fallback", err)
		summaryText = "Ringkasan tidak tersedia - Python service sedang maintenance" //graceful fallback
		duration = 0
	} else {
		var response struct { //untuk parsing response pyhton
			Summary  string `json:"summary"`
			Language string `json:"detected_language"`
			Style    string `json:"style"`
		}
		if err := json.Unmarshal([]byte(summaryResult), &response); err == nil { //jika berhasil parsing
			summaryText = response.Summary
			language = response.Language
		} else {
			summaryText = summaryResult
		}
	}

	// Save summary
	_, err = h.DB.Exec(
		`INSERT INTO summaries (pdf_id, summary_text, summary_style, process_time_ms, language_detected)
		 VALUES ($1, $2, $3, $4, $5)`,
		pdfID,
		summaryText,
		style,
		duration,
		language,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal simpan summary",
		})
	}

	return c.JSON(fiber.Map{
		"pdf_id":            pdfID,
		"filename":          filename,
		"original_filename": file.Filename,
		"style":             style,
		"summary":           summaryText,
		"language":          language,
		"process_time_ms":   duration,
		"success":           true,
	})
}
//for learn this is for upload pdf and auto ai summarizer.
//handlers itu pokok penghubung anatara user d be, mengtur request d response