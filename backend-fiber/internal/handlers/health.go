package handlers

import (
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct { // menyimpan dependecy db
	DB *sql.DB //biar db bs dipakai banyak fungsi
}

func NewHealthHandler(db *sql.DB) *HealthHandler { //untuk inisialisasi
	return &HealthHandler{
		DB: db,
	}
} //Dependency Injection

func (h *HealthHandler) Health(c *fiber.Ctx) error { //method healthhandler, fiber ctx request dn response
	return c.JSON(fiber.Map{
		"status":   "healthy",
		"service":  "fiber-backend",
		"version":  "1.0.0",
		"database": "connected",
	})
}

func (h *HealthHandler) TestDB(c *fiber.Ctx) error { //endpoint tesdb apkh db bnr bnr bs diakses GET/HEALTH/DB
	var count int                                                       //jumlah file pdfnya
	err := h.DB.QueryRow("SELECT COUNT(*) FROM pdf_files").Scan(&count) // ngambil query trus hitung jumlahnya, hasil dimasukkan ke var count
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("DB Error: %v", err), //jika error, kembalikan error, ngirim pesan error json
		})
	}

	return c.JSON(fiber.Map{
		"status":    "ok",  //jika berhasil, kembalikan ok
		"pdf_count": count, //jumlah file pdf
	})
}

//for learn alat pengecek health
