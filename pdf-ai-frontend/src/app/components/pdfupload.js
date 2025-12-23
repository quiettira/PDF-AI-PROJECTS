"use client";
import { useState } from "react";
import styles from "./pdfupload.module.css";

export default function PdfUploader() {
  const [file, setFile] = useState(null);
  const [summary, setSummary] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async () => {
    if (!file) return;

    setLoading(true);
    setSummary("");
    setError("");

    const formData = new FormData();
    formData.append("file", file);

    try {
      const res = await fetch("http://127.0.0.1:8000/summarize", {
        method: "POST",
        body: formData,
      });

      if (!res.ok) {
        let errText = "Gagal meringkas PDF";
        try {
          const errJson = await res.json();
          errText = errJson.detail || JSON.stringify(errJson);
        } catch {
          errText = await res.text();
        }
        throw new Error(errText);
      }

      const data = await res.json();
      setSummary(data.summary || "");
    } catch (err) {
      setError(
        err.message?.includes("Failed to fetch")
          ? "Gagal terhubung ke backend (127.0.0.1:8000)"
          : err.message
      );
    }

    setLoading(false);
  };

  return (
    <div className={styles.container}>
      <div className={styles.wrapper}>
        {/* Header */}
        <div className={styles.header}>
          <div className={styles.iconBox}>📄</div>
          <h1 className={styles.title}>PDF Summarizer</h1>
          <p className={styles.subtitle}>
            Upload file PDF untuk mendapatkan ringkasan otomatis
          </p>
        </div>

        {/* Form Card */}
        <div className={styles.card}>
          <div className={styles.uploadBox}>
            <label className={styles.label}>Upload File PDF</label>

            <input
              type="file"
              accept=".pdf"
              onChange={(e) => setFile(e.target.files[0])}
              className="hidden"
              id="file-upload"
            />
            <label htmlFor="file-upload" className={styles.dropZone}>
              <div className={styles.dropZoneContent}>
                <div className={styles.uploadIcon}>📤</div>
                <p className={styles.dropZoneText}>Klik untuk memilih file</p>
                <p className={styles.dropZoneHint}>Format: PDF (Maks. 10MB)</p>
              </div>
            </label>

            {file && (
              <div className={styles.fileInfo}>
                <span className={styles.fileIcon}>📄</span>
                <span className={styles.fileName}>{file.name}</span>
                <button
                  type="button"
                  onClick={() => setFile(null)}
                  className={styles.clearButton}
                >
                  ✕ Hapus
                </button>
              </div>
            )}
          </div>

          <button
            onClick={handleSubmit}
            disabled={loading || !file}
            className={styles.submitButton}
          >
            {loading ? (
              <>
                <div className={styles.loadingSpinner} />
                <span>Memproses...</span>
              </>
            ) : (
              <>
                <span>✨</span>
                <span>Summarize PDF</span>
              </>
            )}
          </button>
        </div>

        {/* Error Message */}
        {error && (
          <div className={styles.error}>
            <div>
              <h3 className={styles.errorTitle}>Error</h3>
              <p>{error}</p>
            </div>
          </div>
        )}

        {/* Summary Result */}
        {summary && (
          <div className={styles.summaryCard}>
            <div className={styles.summaryHeader}>
              <span className={styles.summaryIcon}>✅</span>
              <h2 className={styles.summaryTitle}>Hasil Ringkasan</h2>
            </div>
            <div className={styles.summaryContent}>{summary}</div>
          </div>
        )}
      </div>
    </div>
  );
}