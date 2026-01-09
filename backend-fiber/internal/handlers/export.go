package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"regexp" //hapus markdown bold italic
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ExportHandler struct{}

func NewExportHandler() *ExportHandler {
	return &ExportHandler{}
}

type ExportRequest struct {
	Summary  string `json:"summary"`
	Filename string `json:"filename,omitempty"`
	Title    string `json:"title,omitempty"`
}

// helper
func cleanMarkdown(text string) string {
	// Remove bold **text** -> text
	boldRegex := regexp.MustCompile(`\*\*(.*?)\*\*`)
	text = boldRegex.ReplaceAllString(text, "$1")

	// Remove italic *text* -> text
	italicRegex := regexp.MustCompile(`\*(.*?)\*`)
	text = italicRegex.ReplaceAllString(text, "$1")

	return text
}


// ExportCSV exports summary as CSV file
func (h *ExportHandler) ExportCSV(c *fiber.Ctx) error {
	var req ExportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Summary == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Summary is required"})
	}

	// Clean markdown formatting
	cleanSummary := cleanMarkdown(req.Summary)

	// Create CSV buffer
	var buf bytes.Buffer 
	writer := csv.NewWriter(&buf)

	// Write header with more structured columns
	writer.Write([]string{"No", "Type", "Content"}) //kolom csv

	// mecah teks perbaris
	lines := strings.Split(cleanSummary, "\n")
	rowNum := 1
	
	for _, line := range lines {
		line = strings.TrimSpace(line)  //bersihin spasi
		if line == "" {
			continue
		}

		
		lineType := "Paragraph"
		content := line
		
		if strings.HasPrefix(line, "•") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			lineType = "Bullet Point"
			content = strings.TrimPrefix(line, "•")
			content = strings.TrimPrefix(content, "-")
			content = strings.TrimPrefix(content, "*")
			content = strings.TrimSpace(content)
		} else if strings.HasSuffix(line, ":") {
			lineType = "Heading"
		}

		writer.Write([]string{
			strings.TrimSpace(strings.Split(content, ":")[0]),
			lineType,
			content,
		})
		rowNum++
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate CSV"})
	}

	// Set filename
	filename := req.Filename
	if filename == "" {
		filename = "summary"
	}
	if !strings.HasSuffix(filename, ".csv") {
		filename += ".csv"
	}

	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename="+filename)

	return c.Send(buf.Bytes())
}

// ExportJSON exports summary as JSON file
func (h *ExportHandler) ExportJSON(c *fiber.Ctx) error {
	var req ExportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Summary == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Summary is required"})
	}

	// Clean markdown formatting
	cleanSummary := cleanMarkdown(req.Summary)

	// Parse summary into structured format
	lines := strings.Split(cleanSummary, "\n")
	var points []string
	var paragraphs []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line is a bullet point
		if strings.HasPrefix(line, "•") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			point := strings.TrimPrefix(line, "•")
			point = strings.TrimPrefix(point, "-")
			point = strings.TrimPrefix(point, "*")
			points = append(points, strings.TrimSpace(point))
		} else {
			paragraphs = append(paragraphs, line)
		}
	}

	// Build JSON response
	title := req.Title
	if title == "" {
		title = "PDF Summary"
	}

	exportData := map[string]interface{}{
		"title":       title,
		"exported_at": time.Now().Format(time.RFC3339),
		"content": map[string]interface{}{
			"full_text":  cleanSummary,
			"paragraphs": paragraphs,
			"points":     points,
		},
		"metadata": map[string]interface{}{
			"total_paragraphs": len(paragraphs),
			"total_points":     len(points),
			"character_count":  len(cleanSummary),
		},
	}

	jsonBytes, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate JSON"})
	}

	// Set filename
	filename := req.Filename
	if filename == "" {
		filename = "summary"
	}
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	c.Set("Content-Type", "application/json; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename="+filename)

	return c.Send(jsonBytes)
}
