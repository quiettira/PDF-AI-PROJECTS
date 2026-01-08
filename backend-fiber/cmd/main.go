package main

import (
	"log" // utk mnmpilkan pesan error ke terminal
	"os" 
	"time" 

	"pdf-backend-fiber/internal/config"
	"pdf-backend-fiber/internal/database"
	"pdf-backend-fiber/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors" 
	"github.com/gofiber/fiber/v2/middleware/logger" //untuk log request
)

func main() {
	// biar gapake utc, jadi sinkron sama waktu indo
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil { //kalo gagal ttp jalan tp warning dksh
		log.Printf("Warning: Could not load Asia/Jakarta timezone: %v", err)
	} else {
		time.Local = loc
	}

	// Load config
	cfg := config.Load()

	// Ensure upload directory exists
	_ = os.MkdirAll(cfg.UploadDir, os.ModePerm) //kalo misal gada folder upload, bakal dibuat, kalo ud ad ya gapapa krn ad _ =

	// Initialize database
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: int(cfg.MaxFileSize),
	})

	// Middleware
	app.Use(logger.New()) //setiap request bs tmpil ke terminal
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type",
	}))

	// Setup routes
	routes.Setup(app, db, cfg)

	// Start server
	log.Println("🚀 Fiber server running on :8080")
	log.Fatal(app.Listen(":8080"))
}

//for learn, jadi ini itu pengatur awal sebelum backend siap menerima request,
//jadi semua yang ada di internal taruh sini bslalu dijalan go cmd/main.go
