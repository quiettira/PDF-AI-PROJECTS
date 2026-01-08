package config

import (
	"os"
	"strconv" //string ke angka, cz always env itu string, jd ngubah ke int untuk max size
)

type Config struct { //wadah configurasi
	MaxFileSize int64  `json:"max_file_size"`
	UploadDir   string `json:"upload_dir"`
	PythonAPI   string `json:"python_api"`
	DBHost      string `json:"db_host"`
	DBPort      string `json:"db_port"`
	DBUser      string `json:"db_user"`
	DBPassword  string `json:"-"` //pake - biar ga di convert ke json
	DBName      string `json:"db_name"`
}

func Load() Config {
	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "10485760"), 10, 64) // 10MB default

	return Config{
		MaxFileSize: maxFileSize,
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		PythonAPI:   getEnv("PYTHON_API_URL", "http://localhost:8000/summarize"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres123"),
		DBName:      getEnv("DB_NAME", "pdf_summarizer"),
	}
} //

func getEnv(key, defaultValue string) string { //fungsi buat ngecek environtment variable, kalo gak ada pake default
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

//for learn, ini pengatur semua seetting penting be.
