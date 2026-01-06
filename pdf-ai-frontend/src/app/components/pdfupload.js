"use client";
import { useState, useEffect } from "react";
import styles from "./pdfupload.module.css";

const API_BASE_URL = process.env.NEXT_PUBLIC_PY_API_BASE_URL || "http://localhost:8000"; // Python AI service
const GO_API_BASE_URL = process.env.NEXT_PUBLIC_GO_API_BASE_URL || "http://localhost:8080"; // Go backend

const SUMMARY_STYLES = [
  { value: "standard", label: "Standard", icon: "📄", iconClass: "emoji emoji-document", description: "Ringkasan paragraf normal" },
  { value: "executive", label: "Executive", icon: "👔", iconClass: "emoji", description: "Ringkasan untuk eksekutif" },
  { value: "bullets", label: "Bullet Points", icon: "•", iconClass: "emoji", description: "Format poin-poin" },
  { value: "detailed", label: "Detailed", icon: "📖", iconClass: "emoji", description: "Ringkasan detail" }
];

export default function PdfUploader() {
  const [file, setFile] = useState(null);
  const [textInput, setTextInput] = useState("");
  const [summary, setSummary] = useState("");
  const [summaryStyle, setSummaryStyle] = useState("standard");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [backendStatus, setBackendStatus] = useState({ go: false, python: false });

  // Check backend health on component mount
  useEffect(() => {
    const checkHealth = async () => {
      try {
        const goResponse = await fetch(`${GO_API_BASE_URL}/health`);
        const pythonResponse = await fetch(`${API_BASE_URL}/health`);
        
        setBackendStatus({
          go: goResponse.ok,
          python: pythonResponse.ok
        });
      } catch (err) {
        console.log("Backend health check failed:", err);
        setBackendStatus({
          go: false,
          python: false
        });
      }
    };
    
    // Check immediately
    checkHealth();
    
    // Check every 10 seconds
    const interval = setInterval(checkHealth, 10000);
    
    // Cleanup interval on unmount
    return () => clearInterval(interval);
  }, []);

  // Error handling yang ramah ke user
  const parseError = (err) => {
    if (err.message?.includes("Failed to fetch")) {
      return `❌ Gagal terhubung ke backend. Pastikan:
      
🔧 Go Backend (${GO_API_BASE_URL}) sedang berjalan
🐍 Python Service (${API_BASE_URL}) sedang berjalan
🗄️ PostgreSQL database tersedia

Jalankan: start-all.bat untuk memulai semua service`;
    }
    if (err.message?.includes("NetworkError")) {
      return "❌ Masalah jaringan. Periksa koneksi internet atau service backend.";
    }
    return err.message;
  };

  // Helper function for API calls (manggil backend)
  // Satu pintu untuk semua request ke backend
  // Dipakai oleh:
  // - /preview
  // - /summarize
  const fetchAPI = async (endpoint, file, params = {}) => {
    const formData = new FormData(); // FormData adalah objek yang digunakan untuk mengirim data form secara aman melalui HTTP POST.
    formData.append("file", file); // Menambahkan file ke formData

    const queryString = new URLSearchParams(params).toString(); // objek params jadi string
    const url = queryString //jika ada params, maka tambahkan params ke url, jika tidak, maka gunakan url tanpa params
      ? `${API_BASE_URL}${endpoint}?${queryString}` //url dengan params
      : `${API_BASE_URL}${endpoint}`; //url tanpa params

    const res = await fetch(url, { //mengirim request ke backend
      method: "POST", 
      body: formData, //mengirim data formData ke backend
    });

    if (!res.ok) { // intinya mengatasi error 
      let errText = "Gagal memproses PDF";
      try {
        const errJson = await res.json();
        errText = errJson.detail || JSON.stringify(errJson);
      } catch {
        errText = await res.text();
      }
      throw new Error(errText);
    }

    return res.json();
  };

  const handleFileSelect = async (selectedFile) => { //mengirim request ke backend untuk preview file PDF
    setFile(selectedFile); 
    setError(""); 

    if (!selectedFile) { //file dihapus atau tidak ada file yang diunggah maka clear textInput dan summary
      setTextInput("");
      setSummary("");
      return;
    }

    // Validate file size (10MB limit)
    const maxSize = 10 * 1024 * 1024; // 10MB in bytes
    if (selectedFile.size > maxSize) {
      setError(`❌ File terlalu besar! Maksimal 10MB. File Anda: ${(selectedFile.size / 1024 / 1024).toFixed(2)}MB`);
      setFile(null);
      return;
    }

    // Validate file type
    if (!selectedFile.name.toLowerCase().endsWith('.pdf')) {
      setError("❌ Hanya file PDF yang diizinkan!");
      setFile(null);
      return;
    }

    setLoading(true);
    try {
      const data = await fetchAPI("/preview", selectedFile);
      setTextInput(data.preview_text || ""); //ngambil data awal pdf yang diunggah
      setSummary(""); // Clear previous summary when new file is selected
    } catch (err) {
      setError(parseError(err));
      setTextInput("");
    } finally {
      setLoading(false);
    }
  };

//mengubah textInput ketika user mengubah teks di textarea
  const handleTextChange = (e) => {
    if (!file) return; //jika tidak ada file yang diunggah maka tidak ada yang diubah
    setTextInput(e.target.value);
  };

  const handleSubmit = async () => {
    if (!file) {
      setError("Unggah PDF terlebih dahulu.");
      return;
    }

    setLoading(true);
    setSummary("");
    setError("");

    try {
      // Always use Go backend (upload + summarize in one call)
      const formData = new FormData();
      formData.append("file", file);
      
      const response = await fetch(
        `${GO_API_BASE_URL}/upload?style=${encodeURIComponent(summaryStyle)}`,
        {
        method: "POST",
        body: formData,
        }
      );
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || "Upload failed");
      }
      
      const data = await response.json();
      setSummary(data.summary || "");
    } catch (err) {
      setError(parseError(err));
    } finally {
      setLoading(false);
    }
  };

  //nyalin ke clipboard
  const handleCopy = () => {
    if (summary) {
      navigator.clipboard.writeText(summary);
    }
  };

 const handleExportTxt = async () => {
  if (!summary) return;

  try {
    const res = await fetch(`${API_BASE_URL}/export/txt`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        summary: summary, // ✅ HARUS OBJEK
      }),
    });

    const blob = await res.blob();
    const url = window.URL.createObjectURL(blob);

    const a = document.createElement("a");
    a.href = url;
    a.download = "summary.txt";
    a.click();

    window.URL.revokeObjectURL(url);
  } catch (err) {
    setError("Gagal export TXT");
  }
};

  const handleExportPdf = async () => {
    if (!summary) return;

    try {
      const res = await fetch(`${API_BASE_URL}/export/pdf`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          summary: summary,
        }),
      });

      if (!res.ok) {
        throw new Error("Failed to export PDF");
      }

      const blob = await res.blob();
      const url = window.URL.createObjectURL(blob);

      const a = document.createElement("a");
      a.href = url;
      a.download = "summary.pdf";
      a.click();

      window.URL.revokeObjectURL(url);
    } catch (err) {
      setError("Gagal export PDF: " + err.message);
    }
  };

  // Helper function to render highlights
  const renderSummary = (text) => {
    if (!text) return null;
    
    // Split text by **highlight** pattern
    const parts = text.split(/(\*\*.*?\*\*)/g);
    
    return parts.map((part, index) => {
      if (part.startsWith('**') && part.endsWith('**')) {
        // Remove asterisks and wrap in highlight span
        return (
          <span key={index} className={styles.highlight}>
            {part.slice(2, -2)}
          </span>
        );
      }
      return part;
    });
  };

  return (
    <div className={styles.container}>
      <div className={styles.wrapper}>
        {/* Header */}
        <div className={styles.header}>
          <h1 className={styles.title}>Asisten Ringkasan AI</h1>
          <p className={styles.subtitle}>
            Unggah dokumen PDF untuk mendapatkan ringkasan yang kalian butuhkan.
          </p>
          
          {/* Backend Status */}
          <div className={styles.statusContainer}>
            <span className={styles.statusItem} style={{ color: backendStatus.go ? "#10b981" : "#ef4444" }}>
              {backendStatus.go ? "✅" : "❌"} Go Backend
            </span>
            <span className={styles.statusItem} style={{ color: backendStatus.python ? "#10b981" : "#ef4444" }}>
              {backendStatus.python ? "✅" : "❌"} Python AI
            </span>
          </div>
        </div>

        {/* Error Message */}
        {error && ( 
          <div className={styles.error}>
            <h3 className={styles.errorTitle}>Error</h3>
            <p>{error}</p>
          </div>
        )}

        {/* Main Content - Two Column Layout */}
        <div className={styles.contentWrapper}>
          {/* Left Panel - Teks Asli */}
          <div className={styles.leftPanel}>
            <h2 className={styles.panelTitle}>
              <span className={`${styles.panelTitleIcon} emoji emoji-document`}>📄</span>
              File PDF Asli
            </h2>

            {/* Upload Area */}
            <div className={styles.uploadArea}>
              <input
                type="file"
                accept=".pdf"
                onChange={(e) => handleFileSelect(e.target.files[0])}
                className={styles.fileInput}
                id="file-upload"
              />
              <div className={styles.uploadZone}> 
                <label htmlFor="file-upload" className={styles.uploadButton}>
                  <span className={`${styles.uploadButtonIcon} emoji emoji-upload`}>⬆️</span>
                  Pilih File PDF 
                </label>
                <p className={styles.uploadHint}>Upload PDF untuk melihat pratinjau</p>
              </div>

              {file && ( 
                <div className={styles.fileInfo}>
                  <span className={`${styles.fileIcon} emoji emoji-document`}>📄</span>
                  <div style={{ flex: 1 }}>
                    <div className={styles.fileName}>{file.name}</div>
                    <div style={{ fontSize: "0.8rem", color: "#666" }}>
                      Size: {(file.size / 1024 / 1024).toFixed(2)}MB / 10MB
                    </div>
                  </div>
                  <button
                    type="button"
                    onClick={() => handleFileSelect(null)} //menghapus file yang diunggah
                    className={styles.clearButton}
                    title="Hapus file"
                  >
                    <span className="emoji emoji-error">✕</span>
                  </button>
                </div>
              )}
            </div>

            {/* Text Input Area */}
            <textarea
              className={styles.textArea}
              placeholder="Pratinjau isi PDF akan muncul di sini setelah Anda mengunggah file."
              value={textInput}
              onChange={handleTextChange}
              disabled={!file} //jika tidak ada file yang diunggah maka textarea tidak bisa diubah
            />

            {/* Style Selector */}
            {(file || textInput.trim()) && (
              <div className={styles.styleSelector}>
                <label className={styles.styleLabel}>Pilih Gaya Ringkasan</label>
                <div className={styles.styleOptions}>
                  {SUMMARY_STYLES.map((style) => ( //looping untuk setiap style ringkasan
                    <button
                      key={style.value}
                      onClick={() => setSummaryStyle(style.value)}
                      className={`${styles.styleOption} ${
                        summaryStyle === style.value ? styles.styleOptionActive : ""
                      }`}
                      title={style.description}
                      type="button"
                    >
                      <span className={`${styles.styleIcon} ${style.iconClass || "emoji"}`}>{style.icon}</span>
                      <span className={styles.styleLabelText}>{style.label}</span>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Submit Button */}
            <button
              onClick={handleSubmit}
              disabled={loading || (!file && !textInput.trim())}
              className={styles.submitButton}
            >
              {loading ? (
                <>
                  <div className={styles.loadingSpinner} />
                  <span>Memproses...</span>
                </>
              ) : (
                <>
                  <span className={`${styles.submitButtonIcon} emoji emoji-star`}>⭐</span>
                  <span>Buatkan Ringkasan</span>
                </>
              )}
            </button>
          </div>

          {/* Right Panel - Hasil Ringkasan */}
          <div className={styles.rightPanel}>
            <h2 className={styles.panelTitle}>
              <span className={`${styles.panelTitleIcon} emoji emoji-star`}>⭐</span>
              Hasil Ringkasan
            </h2>

            <div className={styles.resultContainer}>
              {summary && (
                <button
                  onClick={handleCopy}
                  className={styles.copyButton}
                  title="Salin ringkasan"
                >
                  <span className="emoji emoji-copy">📋</span>
                </button>
                
              )}
              
              <div className={styles.resultTextArea} tabIndex={0}>
                {summary ? (
                  renderSummary(summary)
                ) : (
                  <span className={styles.placeholderText}>
                    Hasil ringkasan akan muncul di sini.
                  </span>
                )}
              </div>
              
              <div>
                <button 
                  onClick={handleExportTxt}
                  disabled={!summary}
                  className={styles.exportButton}
                >
                  📄 Export TXT
                </button>

                <button 
                  onClick={handleExportPdf}
                  disabled={!summary}
                  className={styles.exportButton}
                >
                  📑 Export PDF
                </button>
              </div>

            </div>
          </div>
        </div>
      </div>
    </div>
  );
}