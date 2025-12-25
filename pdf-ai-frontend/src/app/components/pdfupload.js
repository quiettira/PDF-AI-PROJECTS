"use client";
import { useState } from "react";
import Navbar from "./Navbar";
import styles from "./pdfupload.module.css";

const API_BASE_URL = "http://127.0.0.1:8000";

const SUMMARY_STYLES = [
  { value: "standard", label: "Standard", icon: "📄", iconClass: "emoji emoji-document", description: "Ringkasan paragraf normal" },
  { value: "executive", label: "Executive", icon: "👔", iconClass: "emoji", description: "Ringkasan untuk eksekutif" },
  { value: "bullets", label: "Bullet Points", icon: "•", iconClass: "emoji", description: "Format poin-poin" },
  { value: "detailed", label: "Detailed", icon: "📖", iconClass: "emoji", description: "Ringkasan detail" }
];

export default function PdfUploader() {
  const [file, setFile] = useState(null); //menyimpan file PDF yang diunggah
  const [textInput, setTextInput] = useState(""); // menyimpan teks pratinjau dari file PDF
  const [summary, setSummary] = useState(""); // menyimpan hasil ringkasan
  const [summaryStyle, setSummaryStyle] = useState("standard"); //menyimpan gaya ringkasan yang dipilih
  const [loading, setLoading] = useState(false); //menyimpan status loading
  const [error, setError] = useState(""); //menyimpan pesan error

  // Error handling yang ramah ke user
  const parseError = (err) => {
    if (err.message?.includes("Failed to fetch")) {
      return `Gagal terhubung ke backend (${API_BASE_URL})`;
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

  const handleSubmit = async () => { //dipanggil kalau klik tombol ringkas lalu akan di ringkas
    if (!file) {
      setError("Unggah PDF terlebih dahulu."); //nyegah error sebelum request
      return;
    }

    setLoading(true); 
    setSummary(""); 
    setError("");

    //mengirim request ke backend untuk ringkasan
    try {
      const data = await fetchAPI("/summarize", file, { style: summaryStyle });
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

  return (
    <div className={styles.container}>
      <div className={styles.wrapper}>
        {/* Header */}
        <div className={styles.header}>
          <h1 className={styles.title}>Asisten Ringkasan AI</h1>
          <p className={styles.subtitle}>
            Unggah dokumen PDF untuk mendapatkan ringkasan yang kalian butuhkan.
          </p>
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
                  <span className={styles.fileName}>{file.name}</span>
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
              <textarea
                className={styles.resultTextArea}
                value={summary}
                readOnly
                placeholder="Hasil ringkasan akan muncul di sini..."
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}