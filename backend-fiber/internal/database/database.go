package database

import (
	"context" //kontroltimeout, cancel query 
	"database/sql"
	"fmt" //mnyusun stringkoneksidb

	"pdf-backend-fiber/internal/config"

	_ "github.com/lib/pq" //Wajib agar sql.Open("postgres", ...) bisa jalan tamda _ hanya jalankan init 
)

func Init(cfg config.Config) (*sql.DB, error) { //pintu masuk db
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

	// koneksi pool
	db.SetMaxOpenConns(25) //max 25 koneksi aktif
	db.SetMaxIdleConns(5)  //max 5 koneksi nganggur

	// Retry connection with timeout
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		} else if i == 9 {
			return nil, fmt.Errorf("failed to connect to database after 10 attempts: %w", err)
		}
		fmt.Printf("Database connection attempt %d failed, retrying...\n", i+1)
		// Simple sleep without importing time package
		for j := 0; j < 3000000000; j++ {
		} // ~3 second delay
	}

	if err := autoMigrate(db); err != nil { //mengatur tabel
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	fmt.Println("✅ Database connected successfully")
	return db, nil //jika berhasil, kembalikan db dan nil
}

func autoMigrate(db *sql.DB) error { //membuat tabel, nambah kolom lek gada, update schemalama
	ctx := context.Background()

	// Create pdf_files table
	if _, err := db.ExecContext(ctx, ` 
		CREATE TABLE IF NOT EXISTS pdf_files (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL,
			original_filename VARCHAR(255) NOT NULL,
			filepath TEXT NOT NULL,
			filesize BIGINT NOT NULL,
			upload_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	// Add missing columns if they don't exist
	_, _ = db.ExecContext(ctx, `ALTER TABLE pdf_files ADD COLUMN IF NOT EXISTS original_filename VARCHAR(255)`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE pdf_files ADD COLUMN IF NOT EXISTS upload_time TIMESTAMPTZ DEFAULT NOW()`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE summaries ADD COLUMN IF NOT EXISTS summary_style VARCHAR(50) DEFAULT 'standard'`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE summaries ADD COLUMN IF NOT EXISTS language_detected VARCHAR(10)`)

	// Migrate legacy TIMESTAMP columns (no timezone) -> TIMESTAMPTZ
	// Assumption: existing values represent Asia/Jakarta local time.
	_, _ = db.ExecContext(ctx, `ALTER TABLE pdf_files ALTER COLUMN upload_time TYPE TIMESTAMPTZ USING upload_time AT TIME ZONE 'Asia/Jakarta'`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE pdf_files ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'Asia/Jakarta'`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE summaries ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'Asia/Jakarta'`)

	// Add latest_summary column to pdf_files table
	_, _ = db.ExecContext(ctx, `ALTER TABLE pdf_files ADD COLUMN IF NOT EXISTS latest_summary TEXT`)

	// Keep trigger logic minimal: only needed on INSERT (upload/resummarize)
	_, _ = db.ExecContext(ctx, `DROP TRIGGER IF EXISTS trigger_sync_latest_summary_insert ON summaries`)
	_, _ = db.ExecContext(ctx, `DROP TRIGGER IF EXISTS trigger_sync_latest_summary_update ON summaries`)
	_, _ = db.ExecContext(ctx, `DROP TRIGGER IF EXISTS trigger_sync_latest_summary_delete ON summaries`)
	_, _ = db.ExecContext(ctx, `DROP FUNCTION IF EXISTS sync_latest_summary()`)

	// Create trigger function to update latest_summary on new summary insert
	if _, err := db.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION update_latest_summary()
		RETURNS TRIGGER AS $$
		BEGIN
			UPDATE pdf_files
			SET latest_summary = NEW.summary_text
			WHERE id = NEW.pdf_id;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`); err != nil {
		return err
	}

	// Create trigger that fires after INSERT on summaries table
	_, _ = db.ExecContext(ctx, `DROP TRIGGER IF EXISTS trigger_update_latest_summary ON summaries`)
	if _, err := db.ExecContext(ctx, `
		CREATE TRIGGER trigger_update_latest_summary
		AFTER INSERT ON summaries
		FOR EACH ROW
		EXECUTE FUNCTION update_latest_summary();
	`); err != nil {
		return err
	}

	return nil
}

//for learn, penghubung db backend intinya semua isi db ada disini dari tabel dll
