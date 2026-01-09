package routes

import (
	"database/sql"

	"pdf-backend-fiber/internal/config"
	"pdf-backend-fiber/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, db *sql.DB, cfg config.Config) {
	// Initialize handlers
	uploadHandler := handlers.NewUploadHandler(db, cfg)
	pdfHandler := handlers.NewPdfHandler(db, cfg)
	healthHandler := handlers.NewHealthHandler(db)
	exportHandler := handlers.NewExportHandler()

	// Routes
	app.Post("/upload/init", uploadHandler.InitChunkUpload)
	app.Post("/upload/chunk", uploadHandler.UploadChunk)
	app.Get("/upload/status", uploadHandler.UploadStatus)
	app.Post("/upload/complete", uploadHandler.CompleteChunkUpload)
	app.Get("/pdf/:id", pdfHandler.GetPDF)
	app.Delete("/pdf/:id", pdfHandler.DeletePDF)
	app.Get("/history", pdfHandler.GetHistory)
	app.Get("/health", healthHandler.Health)
	app.Get("/test-db", healthHandler.TestDB)
	app.Get("/simple-pdfs", pdfHandler.SimplePDFs)
	app.Get("/simple-pdf/:id", pdfHandler.SimplePDFByID)
	app.Put("/update-pdf/:id", pdfHandler.UpdatePDF)
	app.Post("/resummarize/:id", pdfHandler.Resummarize)
	app.Get("/summaries/:id", pdfHandler.GetSummaries)

	// Export routes (CSV & JSON)
	app.Post("/export/csv", exportHandler.ExportCSV)
	app.Post("/export/json", exportHandler.ExportJSON)
}

//Kita membuat handler sekali
//DB & config disuntikkan ke handler
//Supaya:
//rapi
//gampang dites
//tidak pakai global variable
// Ini disebut Dependency Injection
