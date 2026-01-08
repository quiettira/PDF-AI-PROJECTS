package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json" //persing response pyhton
	"fmt"           //format nama file d pesan
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time" //bikin namafile unik katanya

	"pdf-backend-fiber/internal/config" //disini dia utk max size d folder upload
	"pdf-backend-fiber/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UploadHandler struct {
	DB     *sql.DB
	Config config.Config
	Python *services.PythonClient
}

func isValidPDFFileHeader(fileHeaderBytes []byte) bool { //cek apakah file pdf valid, nanti dipake nanti di completechunkupload
	trimmed := bytes.TrimLeft(fileHeaderBytes, "\x00\t\n\r\f ")
	return bytes.HasPrefix(trimmed, []byte("%PDF-")) //header file harus diawali "%PDF-"
} // Dia TrimLeft dulu (buang byte kosong/spasi/enter) lalu cek prefix %PDF-.

func NewUploadHandler(db *sql.DB, cfg config.Config) *UploadHandler { //nah ini buat handler sekali alu diupload di routes
	return &UploadHandler{
		DB:     db,
		Config: cfg,
		Python: services.NewPythonClient(cfg.PythonAPI),
	}
}

type uploadMeta struct { //tanpa ini pengecekan biasanya akan salah
	OriginalFilename string `json:"original_filename"`
	FileSize         int64  `json:"file_size"`
	ChunkSize        int64  `json:"chunk_size"`
	TotalChunks      int    `json:"total_chunks"`
	Style            string `json:"style"`
	CreatedAtUnix    int64  `json:"created_at_unix"`
}

// expectedChunkSize menghitung ukuran chunk yang seharusnya.
// Semua chunk biasanya sama dengan meta.ChunkSize, kecuali chunk terakhir bisa lebih kecil.
// Ini dipakai untuk:
// - validasi agar chunk yang tersimpan tidak korup/setengah
// - membuat retry idempotent (chunk yg sudah valid tidak perlu diupload ulang)
func (h *UploadHandler) expectedChunkSize(meta uploadMeta, chunkIndex int) (int64, error) {
	if chunkIndex < 0 || chunkIndex >= meta.TotalChunks { 
		return 0, fmt.Errorf("chunk_index out of range") 
	}
	start := int64(chunkIndex) * meta.ChunkSize //hitung byte ke brp chunk dimulai
	if start < 0 || start >= meta.FileSize {
		return 0, fmt.Errorf("chunk range invalid")
	}
	remain := meta.FileSize - start //Hitung sisa byte file dari posisi start sampai akhir
	if remain <= 0 {
		return 0, fmt.Errorf("chunk range invalid")
	}
	if remain >= meta.ChunkSize { //jika sisa byte lebih besar dari chunk size, maka return chunk size
		return meta.ChunkSize, nil
	}
	return remain, nil //jika sisa byte lebih kecil dari chunk size, maka return sisa byte
}

// atomicWriteMultipartFile menyimpan file upload secara atomic:
// tulis dulu ke file .tmp lalu rename ke nama final.
// Tujuan utamanya agar kalau koneksi putus di tengah upload, file .part tidak pernah tersimpan setengah.
func atomicWriteMultipartFile(fileHeader *multipart.FileHeader, finalPath string) (int64, error) {
	f, err := fileHeader.Open() //buka file
	if err != nil {
		return 0, err
	}
	defer f.Close()

	dir := filepath.Dir(finalPath)
	tmp, err := os.CreateTemp(dir, ".chunk-*.tmp")
	if err != nil {
		return 0, err
	}
	tmpName := tmp.Name() 

	closeErr := func() error { //helper 
		if err := tmp.Close(); err != nil { 
			return err
		}
		return nil
	}

	n, copyErr := io.Copy(tmp, f) //copy file ke tmp
	if copyErr != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return 0, copyErr
	}
	if err := tmp.Sync(); err != nil { //pakai sync untuk memastikan data tersimpan ke disk
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return 0, err
	}
	if err := closeErr(); err != nil {
		_ = os.Remove(tmpName)
		return 0, err
	}
	if err := os.Rename(tmpName, finalPath); err != nil {
		_ = os.Remove(tmpName)
		return 0, err
	}
	return n, nil
}

func (h *UploadHandler) chunkRootDir() string { //folder tempat menyimpan chunk
	return filepath.Join(h.Config.UploadDir, ".chunks")
}

func (h *UploadHandler) uploadDir(uploadID string) string { //folder tempat menyimpan chunk per upload
	return filepath.Join(h.chunkRootDir(), uploadID)
}

func (h *UploadHandler) uploadMetaPath(uploadID string) string { //path untuk menyimpan metadata
	return filepath.Join(h.uploadDir(uploadID), "meta.json")
}

func (h *UploadHandler) chunkPath(uploadID string, chunkIndex int) string {
	return filepath.Join(h.uploadDir(uploadID), fmt.Sprintf("%08d.part", chunkIndex))
} //biar file urut rapi dan gampang merge 0..N tanpa sorting aneh.

func normalizeStyle(style string) string {
	style = strings.ToLower(strings.TrimSpace(style))
	switch style {
	case "standard", "executive", "bullets", "detailed":
		return style
	default:
		return "standard" //defaultnya standard
	}
}

// InitChunkUpload membuat "upload session" untuk chunk upload.
// Di sini server:
// - validasi file (hanya .pdf, size <= MaxFileSize)
// - normalisasi style
// - membuat folder .chunks/<upload_id>
// - menulis meta.json sebagai sumber kebenaran (berapa total chunk, ukuran file, dll)
func (h *UploadHandler) InitChunkUpload(c *fiber.Ctx) error {
	var req struct {
		OriginalFilename string `json:"original_filename"`
		FileSize         int64  `json:"file_size"`
		ChunkSize        int64  `json:"chunk_size"`
		TotalChunks      int    `json:"total_chunks"`
		Style            string `json:"style"`
		UploadID         string `json:"upload_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if req.OriginalFilename == "" || req.FileSize <= 0 || req.ChunkSize <= 0 || req.TotalChunks <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid upload metadata"})
	}

	if strings.ToLower(filepath.Ext(req.OriginalFilename)) != ".pdf" {
		return c.Status(400).JSON(fiber.Map{"error": "Hanya file PDF yang diizinkan"})
	}

	if req.FileSize > h.Config.MaxFileSize {
		return c.Status(400).JSON(fiber.Map{"error": "File terlalu besar (maks 10MB)"})
	}

	style := normalizeStyle(req.Style)

	uploadID := strings.TrimSpace(req.UploadID) //kirim upload id
	if uploadID == "" {
		uploadID = uuid.NewString() //lek gaada buat id baru
	}

	if err := os.MkdirAll(h.chunkRootDir(), os.ModePerm); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to init upload"})
	}
	if err := os.MkdirAll(h.uploadDir(uploadID), os.ModePerm); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to init upload"})
	}

	meta := uploadMeta{ //single source of truth
		OriginalFilename: req.OriginalFilename,
		FileSize:         req.FileSize,
		ChunkSize:        req.ChunkSize,
		TotalChunks:      req.TotalChunks,
		Style:            style,
		CreatedAtUnix:    time.Now().Unix(),
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to init upload"})
	}
	if err := os.WriteFile(h.uploadMetaPath(uploadID), b, 0644); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to init upload"})
	}

	return c.JSON(fiber.Map{
		"upload_id":    uploadID,
		"chunk_size":   meta.ChunkSize,
		"total_chunks": meta.TotalChunks,
		"style":        meta.Style,
		"success":      true,
	})
}

// UploadChunk menerima 1 potongan (chunk) dari file.
// Kunci penting di sini:
// - upload_id = identitas satu file
// - chunk_index = urutan chunk
// - idempotent: kalau chunk sudah ada & ukurannya benar => sudah sukses (retry aman)
// - atomic write: chunk tidak akan tersimpan setengah
func (h *UploadHandler) UploadChunk(c *fiber.Ctx) error {
	uploadID := strings.TrimSpace(c.FormValue("upload_id"))
	chunkIndexStr := strings.TrimSpace(c.FormValue("chunk_index"))
	if uploadID == "" || chunkIndexStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing upload_id or chunk_index"})
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil || chunkIndex < 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid chunk_index"})
	}

	// meta.json wajib ada; kalau tidak ada berarti upload session belum di-init
	metaBytes, err := os.ReadFile(h.uploadMetaPath(uploadID))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Upload session not found"})
	}

	var meta uploadMeta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid upload session"})
	}

	if chunkIndex >= meta.TotalChunks {
		return c.Status(400).JSON(fiber.Map{"error": "chunk_index out of range"})
	}

	// hitung ukuran chunk yang seharusnya, supaya bisa deteksi chunk korup/setengah
	expectedSize, err := h.expectedChunkSize(meta, chunkIndex)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "chunk_index out of range"})
	}

	file, err := c.FormFile("chunk")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Chunk file not found"})
	}

	if err := os.MkdirAll(h.uploadDir(uploadID), os.ModePerm); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save chunk"})
	}

	// path final chunk: .chunks/<upload_id>/<chunk_index>.part
	chunkPath := h.chunkPath(uploadID, chunkIndex)
	if st, err := os.Stat(chunkPath); err == nil {
		if st.Size() == expectedSize {
			return c.JSON(fiber.Map{"success": true, "upload_id": uploadID, "chunk_index": chunkIndex, "already_uploaded": true})
		}
		// kalau file sudah ada tapi ukurannya tidak sesuai, berarti korup/setengah -> hapus lalu upload ulang
		_ = os.Remove(chunkPath)
	}

	written, err := atomicWriteMultipartFile(file, chunkPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save chunk"})
	}
	if written != expectedSize {
		_ = os.Remove(chunkPath)
		return c.Status(400).JSON(fiber.Map{"error": "Chunk size mismatch"})
	}

	return c.JSON(fiber.Map{"success": true, "upload_id": uploadID, "chunk_index": chunkIndex})
}

// UploadStatus dipakai untuk resume.
// Server mengembalikan daftar chunk_index yang sudah diterima (file *.part yang ada di folder upload).
func (h *UploadHandler) UploadStatus(c *fiber.Ctx) error {
	uploadID := strings.TrimSpace(c.Query("upload_id"))
	if uploadID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing upload_id"})
	}

	metaBytes, err := os.ReadFile(h.uploadMetaPath(uploadID))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Upload session not found"})
	}

	var meta uploadMeta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid upload session"})
	}

	entries, err := os.ReadDir(h.uploadDir(uploadID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read upload status"})
	}

	received := make([]int, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".part") {
			continue
		}
		idxStr := strings.TrimSuffix(name, ".part")
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			continue
		}
		received = append(received, idx)
	}
	sort.Ints(received)

	return c.JSON(fiber.Map{
		"success":       true,
		"upload_id":     uploadID,
		"received":      received,
		"total_chunks":  meta.TotalChunks,
		"chunk_size":    meta.ChunkSize,
		"file_size":     meta.FileSize,
		"original_name": meta.OriginalFilename,
		"style":         meta.Style,
	})
}

// CompleteChunkUpload menyelesaikan upload chunk:
// - validasi semua chunk 0..N-1 sudah ada
// - gabungkan semua chunk berurutan ke file final di UploadDir
// - validasi hasil merge (magic bytes %PDF- dan size harus sama)
// - simpan metadata ke DB + panggil Python summarizer
// - bersihkan folder chunk supaya hemat storage
func (h *UploadHandler) CompleteChunkUpload(c *fiber.Ctx) error {
	var req struct {
		UploadID string `json:"upload_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	uploadID := strings.TrimSpace(req.UploadID)
	if uploadID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing upload_id"})
	}

	metaBytes, err := os.ReadFile(h.uploadMetaPath(uploadID))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Upload session not found"})
	}

	var meta uploadMeta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid upload session"})
	}

	for i := 0; i < meta.TotalChunks; i++ {
		if _, err := os.Stat(h.chunkPath(uploadID, i)); err != nil {
			if os.IsNotExist(err) {
				return c.Status(409).JSON(fiber.Map{"error": "Chunks incomplete", "missing_chunk": i})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Failed to validate chunks"})
		}
	}

	if err := os.MkdirAll(h.Config.UploadDir, os.ModePerm); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
	}
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), meta.OriginalFilename)
	savePath := filepath.Join(h.Config.UploadDir, filename)

	dst, err := os.OpenFile(savePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
	}
	defer dst.Close()

	for i := 0; i < meta.TotalChunks; i++ {
		p := h.chunkPath(uploadID, i)
		src, err := os.Open(p)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read chunk"})
		}
		_, copyErr := io.Copy(dst, src)
		_ = src.Close()
		if copyErr != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to assemble file"})
		}
	}

	if err := dst.Sync(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to finalize file"})
	}

	assembled, err := os.Open(savePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membaca file upload"})
	}
	defer assembled.Close()

	header := make([]byte, 1024)
	n, _ := io.ReadFull(assembled, header)
	if n <= 0 || !isValidPDFFileHeader(header[:n]) {
		_ = os.Remove(savePath)
		return c.Status(400).JSON(fiber.Map{"error": "File bukan PDF valid (signature tidak sesuai)"})
	}

	fi, err := assembled.Stat()
	if err != nil {
		_ = os.Remove(savePath)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to stat file"})
	}
	if fi.Size() != meta.FileSize {
		_ = os.Remove(savePath)
		return c.Status(400).JSON(fiber.Map{"error": "File size mismatch"})
	}
	if fi.Size() > h.Config.MaxFileSize {
		_ = os.Remove(savePath)
		return c.Status(400).JSON(fiber.Map{"error": "File terlalu besar (maks 10MB)"})
	}

	var pdfID int
	err = h.DB.QueryRow(
		`INSERT INTO pdf_files (filename, original_filename, filepath, filesize, upload_time)
		 VALUES ($1, $2, $3, $4, NOW()) RETURNING id`,
		filename,
		meta.OriginalFilename,
		savePath,
		fi.Size(),
	).Scan(&pdfID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal simpan metadata PDF"})
	}

	summaryResult, duration, err := h.Python.Summarize(savePath, meta.Style)
	summaryText := ""
	language := "unknown"

	if err != nil {
		log.Printf("Python service failed: %v, using fallback", err)
		summaryText = "Ringkasan tidak tersedia - Python service sedang maintenance"
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
		meta.Style,
		duration,
		language,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal simpan summary"})
	}

	_ = os.RemoveAll(h.uploadDir(uploadID))

	return c.JSON(fiber.Map{
		"pdf_id":            pdfID,
		"filename":          filename,
		"original_filename": meta.OriginalFilename,
		"style":             meta.Style,
		"summary":           summaryText,
		"language":          language,
		"process_time_ms":   duration,
		"success":           true,
	})
}

//for learn this is for upload pdf and auto ai summarizer.
//handlers itu pokok penghubung anatara user d be, mengtur request d response
