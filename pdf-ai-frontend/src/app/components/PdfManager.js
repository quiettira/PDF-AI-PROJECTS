"use client";
import { useState, useEffect } from "react";
import styles from "./PdfManager.module.css";

const GO_API_BASE_URL = process.env.NEXT_PUBLIC_GO_API_BASE_URL || "http://localhost:8080";

export default function PdfManager() {
  const [pdfList, setPdfList] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [selectedFile, setSelectedFile] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [editingPdf, setEditingPdf] = useState(null);
  const [editForm, setEditForm] = useState({ original_filename: "" });
  const [viewingPdf, setViewingPdf] = useState(null);
  const [pdfDetails, setPdfDetails] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 6;

  // Fetch PDF list from Go backend
  const fetchPdfList = async () => {
    setLoading(true);
    try {
      const response = await fetch(`${GO_API_BASE_URL}/simple-pdfs`);
      if (!response.ok) throw new Error("Failed to fetch PDF list");
      const data = await response.json();
      setPdfList(data.pdfs || []);
      setCurrentPage(1);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Upload PDF to Go backend
  const handleUpload = async (file) => {
    if (!file) return;
    
    // Validate file size (10MB limit)
    const maxSize = 10 * 1024 * 1024; // 10MB in bytes
    if (file.size > maxSize) {
      setError(`❌ File terlalu besar! Maksimal 10MB. File Anda: ${(file.size / 1024 / 1024).toFixed(2)}MB`);
      setSelectedFile(null);
      return;
    }

    // Validate file type
    if (!file.name.toLowerCase().endsWith('.pdf')) {
      setError("❌ Hanya file PDF yang diizinkan!");
      setSelectedFile(null);
      return;
    }
    
    setUploading(true);
    setError("");
    
    try {
      const formData = new FormData();
      formData.append("file", file);
      
      const response = await fetch(`${GO_API_BASE_URL}/upload`, {
        method: "POST",
        body: formData,
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || "Upload failed");
      }
      
      const result = await response.json();
      console.log("Upload success:", result);
      
      // Refresh list after upload
      await fetchPdfList();
      setSelectedFile(null);
      
    } catch (err) {
      setError(err.message);
    } finally {
      setUploading(false);
    }
  };

  // View PDF details (READ operation)
  const handleViewDetails = async (pdfId) => {
    try {
      const response = await fetch(`${GO_API_BASE_URL}/simple-pdf/${pdfId}`);
      if (!response.ok) throw new Error("Failed to fetch PDF details");
      const data = await response.json();
      
      // Get summaries for this PDF
      const summariesResponse = await fetch(`${GO_API_BASE_URL}/summaries/${pdfId}`);
      if (summariesResponse.ok) {
        const summariesData = await summariesResponse.json();
        data.summaries = summariesData.summaries || [];
      } else {
        data.summaries = [];
      }
      
      setPdfDetails(data);
      setViewingPdf(pdfId);
    } catch (err) {
      setError(err.message);
    }
  };

  // Update PDF
  const handleUpdate = async (pdfId) => {
    try {
      const response = await fetch(`${GO_API_BASE_URL}/update-pdf/${pdfId}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(editForm),
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || "Update failed");
      }
      
      // Refresh list after update
      await fetchPdfList();
      setEditingPdf(null);
      setEditForm({ original_filename: "" });
      
    } catch (err) {
      setError(err.message);
    }
  };

  // Re-summarize PDF with different style
  const handleResummarize = async (pdfId, style) => {
    try {
      const response = await fetch(`${GO_API_BASE_URL}/resummarize/${pdfId}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ style }),
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || "Re-summarize failed");
      }
      
      // Refresh list after re-summarize
      await fetchPdfList();
      
    } catch (err) {
      setError(err.message);
    }
  };

  // Delete PDF
  const handleDelete = async (pdfId) => {
    if (!confirm("Are you sure you want to delete this PDF?")) return;
    
    try {
      const response = await fetch(`${GO_API_BASE_URL}/pdf/${pdfId}`, {
        method: "DELETE",
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || "Delete failed");
      }
      
      // Refresh list after deletion
      await fetchPdfList();
      
    } catch (err) {
      setError(err.message);
    }
  };

  // Format file size
  const formatFileSize = (bytes) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return "Unknown date";
    try {
      const date = new Date(dateString);
      if (isNaN(date.getTime())) return "Invalid date";
      
      // Display time in Indonesia timezone
      return date.toLocaleString("id-ID", {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        timeZone: 'Asia/Jakarta'
      });
    } catch (error) {
      return "Invalid date";
    }
  };

  useEffect(() => {
    fetchPdfList();
  }, []);

  useEffect(() => {
    const totalPages = Math.max(1, Math.ceil(pdfList.length / itemsPerPage));
    setCurrentPage((prev) => Math.min(prev, totalPages));
  }, [pdfList.length]);

  const totalPages = Math.max(1, Math.ceil(pdfList.length / itemsPerPage));
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedPdfList = pdfList.slice(startIndex, endIndex);

  const handlePageChange = (page) => {
    const next = Math.min(Math.max(1, page), totalPages);
    setCurrentPage(next);
  };

  const getPageNumbers = () => {
    if (totalPages <= 7) return Array.from({ length: totalPages }, (_, i) => i + 1);

    const pages = new Set([1, totalPages]);
    for (let p = currentPage - 1; p <= currentPage + 1; p++) {
      if (p > 1 && p < totalPages) pages.add(p);
    }
    return Array.from(pages).sort((a, b) => a - b);
  };

  return (
    <div className={styles.container}>
      <div className={styles.wrapper}>
        <h1 className={styles.title}>PDF AI MANAGER</h1>
        <p className={styles.subtitle}>
          ringkasan yang pernah kamu buat ada disini.
        </p>
        <div className={styles.uploadSection}>
          <div className={styles.uploadArea}>
            <h2 className={styles.sectionTitle}>➕ Upload PDF Baru</h2>

            <input
              type="file"
              accept=".pdf"
              onChange={(e) => setSelectedFile(e.target.files[0])}
              className={styles.fileInput}
              id="pdf-upload"
              disabled={uploading}
            />
            <label htmlFor="pdf-upload" className={styles.uploadButton}>
              📄 Pilih File PDF
            </label>
            
            {selectedFile && (
              <div className={styles.selectedFile}>
                <div style={{ flex: 1 }}>
                  <div>📄 {selectedFile.name}</div>
                  <div style={{ fontSize: "0.8rem", color: "#94a3b8", marginTop: "4px" }}>
                    Size: {(selectedFile.size / 1024 / 1024).toFixed(2)}MB / 10MB
                  </div>
                </div>
                <button
                  onClick={() => handleUpload(selectedFile)}
                  disabled={uploading}
                  className={styles.uploadBtn}
                >
                  {uploading ? "⏳ Uploading..." : "🚀 Upload & Summarize"}
                </button>
                <button
                  onClick={() => setSelectedFile(null)}
                  className={styles.cancelBtn}
                >
                  ❌
                </button>
              </div>
            )}
          </div>
        </div>

        {error && (
          <div className={styles.error}>
            <h3>❌ Error</h3>
            <p>{error}</p>
            <button onClick={() => setError("")} className={styles.closeError}>
              ✕
            </button>
          </div>
        )}

        <div className={styles.listSection}>
          <div className={styles.listHeader}>
            <h2 className={styles.sectionTitle}>📋 Daftar PDF</h2>
            <button onClick={fetchPdfList} className={styles.refreshBtn}>
              🔄 Refresh
            </button>
          </div>

          {loading ? (
            <div className={styles.loading}>⏳ Loading...</div>
          ) : pdfList.length === 0 ? (
            <div className={styles.empty}>
              📭 Belum ada PDF yang diupload
            </div>
          ) : (
            <>
              <div className={styles.pdfGrid}>
                {paginatedPdfList.map((pdf) => (
                  <div key={pdf.id} className={styles.pdfCard}>
                    <div className={styles.pdfHeader}>
                      <h3 className={styles.pdfTitle}>📄 {pdf.filename}</h3>
                      <div className={styles.pdfActions}>
                        <button
                          onClick={() => handleViewDetails(pdf.id)}
                          className={styles.viewBtn}
                          title="View Details"
                        >
                          👁️
                        </button>
                        <button
                          onClick={() => window.open(`${GO_API_BASE_URL}/pdf/${pdf.id}/download`, "_blank")}
                          className={styles.downloadBtn}
                          title="Download PDF"
                        >
                          📥
                        </button>
                        <button
                          onClick={() => {
                            setEditingPdf(pdf.id);
                            setEditForm({ original_filename: pdf.original_filename });
                          }}
                          className={styles.editBtn}
                          title="Edit PDF"
                        >
                          ✏️
                        </button>
                        <button
                          onClick={() => handleDelete(pdf.id)}
                          className={styles.deleteBtn}
                          title="Delete PDF"
                        >
                          🗑️
                        </button>
                      </div>
                    </div>
                
                {/* Edit Form */}
                {editingPdf === pdf.id && (
                  <div className={styles.editForm}>
                    <h4>✏️ Edit PDF</h4>
                    <input
                      type="text"
                      value={editForm.original_filename}
                      onChange={(e) => setEditForm({ ...editForm, original_filename: e.target.value })}
                      placeholder="Original filename"
                      className={styles.editInput}
                    />
                    <div className={styles.resummarizeSection}>
                      <h5>🔄 Re-summarize with different style:</h5>
                      <div className={styles.styleButtons}>
                        <button onClick={() => handleResummarize(pdf.id, 'standard')} className={styles.styleBtn}>Standard</button>
                        <button onClick={() => handleResummarize(pdf.id, 'executive')} className={styles.styleBtn}>Executive</button>
                        <button onClick={() => handleResummarize(pdf.id, 'bullets')} className={styles.styleBtn}>Bullets</button>
                        <button onClick={() => handleResummarize(pdf.id, 'detailed')} className={styles.styleBtn}>Detailed</button>
                      </div>
                    </div>

                    <div className={styles.editActions}>
                      <button
                        onClick={() => handleUpdate(pdf.id)}
                        className={styles.saveBtn}
                      >
                        💾 Save
                      </button>
                      <button
                        onClick={() => setEditingPdf(null)}
                        className={styles.editCancelBtn}
                      >
                        ❌ Cancel
                      </button>
                    </div>
                  </div>
                )}

                <div className={styles.pdfInfo}>
                  <p><strong>Original:</strong> {pdf.original_filename}</p>
                  <p><strong>Size:</strong> {formatFileSize(pdf.filesize)}</p>
                  <p><strong>Uploaded:</strong> {formatDate(pdf.upload_time)}</p>
                  {pdf.summary && pdf.summary.process_time_ms && (
                    <p><strong>Process Time:</strong> {pdf.summary.process_time_ms}ms</p>
                  )}
                  {pdf.summary && pdf.summary.language_detected && (
                    <p><strong>Language:</strong> {pdf.summary.language_detected}</p>
                  )}
                </div>

                {pdf.summary && (
                  <div className={styles.summaryPreview}>
                    <h4>📝 Summary Preview ({pdf.summary.style}):</h4>
                    <p className={styles.summaryText}>
                      {typeof pdf.summary.text === "string" && pdf.summary.text.trim().length > 0
                        ? (pdf.summary.text.trim().length > 200
                            ? `${pdf.summary.text.trim().substring(0, 200)}...`
                            : pdf.summary.text.trim())
                        : "No summary available"}
                    </p>
                  </div>
                )}
              </div>
            ))}
          </div>

          <div className={styles.pagination}>
            <div className={styles.paginationInfo}>
              Menampilkan {startIndex + 1}-{Math.min(endIndex, pdfList.length)} dari {pdfList.length}
            </div>
            <div className={styles.paginationControls}>
              <button
                type="button"
                className={styles.pageBtn}
                onClick={() => handlePageChange(currentPage - 1)}
                disabled={currentPage === 1}
              >
                Prev
              </button>

              {getPageNumbers().map((page, idx, arr) => (
                <span key={page} className={styles.pageNumberWrap}>
                  {idx > 0 && page - arr[idx - 1] > 1 && (
                    <span className={styles.pageEllipsis}>…</span>
                  )}
                  <button
                    type="button"
                    className={`${styles.pageBtn} ${page === currentPage ? styles.pageBtnActive : ""}`}
                    onClick={() => handlePageChange(page)}
                  >
                    {page}
                  </button>
                </span>
              ))}

              <button
                type="button"
                className={styles.pageBtn}
                onClick={() => handlePageChange(currentPage + 1)}
                disabled={currentPage === totalPages}
              >
                Next
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  </div>

  {/* PDF Details Modal */}
  {viewingPdf && pdfDetails && (
    <div className={styles.modal}>
      <div className={styles.modalContent}>
        <div className={styles.modalHeader}>
          <h2>📄 PDF Details</h2>
          <button 
            onClick={() => {
              setViewingPdf(null);
              setPdfDetails(null);
            }}
            className={styles.closeBtn}
          >
            ✕
          </button>
        </div>
        
        <div className={styles.modalBody}>
          <div className={styles.detailSection}>
            <h3>📋 File Information</h3>
            <div>
              <p><strong>ID:</strong> {pdfDetails.id}</p>
              <p><strong>Filename:</strong> {pdfDetails.filename}</p>
              <p><strong>Original Name:</strong> {pdfDetails.original_filename}</p>
              <p><strong>File Size:</strong> {formatFileSize(pdfDetails.filesize)}</p>
              <p><strong>Upload Time:</strong> {formatDate(pdfDetails.upload_time)}</p>
              <p><strong>File Path:</strong> {pdfDetails.filepath}</p>
            </div>
          </div>
          
          {pdfDetails.summaries && pdfDetails.summaries.length > 0 && (
            <div className={styles.detailSection}>
              <h3>📝 Summaries ({pdfDetails.summaries.length})</h3>
              {pdfDetails.summaries.map((summary, index) => (
                <div key={summary.id} className={styles.summaryDetail}>
                  <h4>Summary #{index + 1} - {summary.summary_style}</h4>
                  <div>
                    <p><strong>Language:</strong> {summary.language_detected}</p>
                    <p><strong>Process Time:</strong> {summary.process_time_ms}ms</p>
                    <p><strong>Created:</strong> {formatDate(summary.created_at)}</p>
                  </div>
                  <div className={styles.summaryText}>
                    <strong>Content:</strong>
                    <div className={styles.summaryContent}>
                      {summary.summary_text}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )}
</div>
);
}