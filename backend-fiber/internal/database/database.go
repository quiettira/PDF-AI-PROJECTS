package database

import (
	"context" //kontroltimeout
	"database/sql"
	"fmt" //mnyusun stringkoneksidb

	"pdf-backend-fiber/internal/config"

	_ "github.com/lib/pq" //Wajib agar sql.Open("postgres", ...) bisa jalan
)

func Init(cfg config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=30",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr) //membuat koneksi db
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Retry connection with timeout
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		} else if i == 9 {
			return nil, fmt.Errorf("failed to connect to database after 10 attempts: %w", err)
		}
		fmt.Printf("Database connection attempt %d failed, retrying...\n", i+1)
		// Simple sleep without importing time package
		for j := 0; j < 3000000000; j++ {} // ~3 second delay
	}

	if err := autoMigrate(db); err != nil { //mengatur tabel
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	fmt.Println("✅ Database connected successfully")
	return db, nil //jika berhasil, kembalikan db dan nil
}

func autoMigrate(db *sql.DB) error { //membuat tabel, nambah kolom lek gada
	ctx := context.Background() 

	// Create pdf_files table
	if _, err := db.ExecContext(ctx, ` 
		CREATE TABLE IF NOT EXISTS pdf_files (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL,
			original_filename VARCHAR(255) NOT NULL,
			filepath TEXT NOT NULL,
			filesize BIGINT NOT NULL,
			upload_time TIMESTAMP NOT NULL DEFAULT NOW(),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	// Create summaries table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS summaries (
			id SERIAL PRIMARY KEY,
			pdf_id INT NOT NULL REFERENCES pdf_files(id) ON DELETE CASCADE,
			summary_text TEXT NOT NULL,
			summary_style VARCHAR(50) NOT NULL DEFAULT 'standard',
			process_time_ms BIGINT NOT NULL,
			language_detected VARCHAR(10),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	// Add missing columns if they don't exist
	_, _ = db.ExecContext(ctx, `ALTER TABLE pdf_files ADD COLUMN IF NOT EXISTS original_filename VARCHAR(255)`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE pdf_files ADD COLUMN IF NOT EXISTS upload_time TIMESTAMP DEFAULT NOW()`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE summaries ADD COLUMN IF NOT EXISTS summary_style VARCHAR(50) DEFAULT 'standard'`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE summaries ADD COLUMN IF NOT EXISTS language_detected VARCHAR(10)`)

	return nil
}

//for learn, penghubung db backend intinya semua isi db ada disini dari tabel dll