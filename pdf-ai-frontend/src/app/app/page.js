"use client";

import { useState } from "react";
import Link from "next/link";
import PdfUpload from "../components/pdfupload";
import PdfManager from "../components/PdfManager";
import styles from "./app.module.css";

export default function AppPage() {
  const [activeTab, setActiveTab] = useState("uploader");
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className={styles.container}>
      {/* Hamburger Menu */}
      <button 
        className={styles.hamburger}
        onClick={() => setSidebarOpen(!sidebarOpen)}
        aria-label="Toggle menu"
      >
        <span className={`${styles.hamburgerLine} ${sidebarOpen ? styles.open : ""}`}></span>
        <span className={`${styles.hamburgerLine} ${sidebarOpen ? styles.open : ""}`}></span>
        <span className={`${styles.hamburgerLine} ${sidebarOpen ? styles.open : ""}`}></span>
      </button>

      {/* Overlay */}
      {sidebarOpen && (
        <div className={styles.overlay} onClick={() => setSidebarOpen(false)} />
      )}

      {/* Sidebar */}
      <aside className={`${styles.sidebar} ${sidebarOpen ? styles.sidebarOpen : ""}`}>
        <div className={styles.header}>
          <h1 className={styles.title}>📄 PDF AI Summarizer</h1>
          <p className={styles.subtitle}>Bootcamp Project</p>
        </div>

        <Link href="/" className={styles.backButton}>
          ← Kembali ke Beranda
        </Link>

        <nav className={styles.nav}>
          <button
            onClick={() => { setActiveTab("uploader"); setSidebarOpen(false); }}
            className={`${styles.navItem} ${activeTab === "uploader" ? styles.active : ""}`}
          >
            <span className={styles.icon}>📄</span>
            <span className={styles.label}>PDF Summarizer</span>
          </button>
          <button
            onClick={() => { setActiveTab("manager"); setSidebarOpen(false); }}
            className={`${styles.navItem} ${activeTab === "manager" ? styles.active : ""}`}
          >
            <span className={styles.icon}>📁</span>
            <span className={styles.label}>PDF Manager</span>
          </button>
        </nav>

        <div className={styles.infoSection}>
          <h4 className={styles.infoTitle}>ℹ️ Info</h4>
          <div className={styles.infoContent}>
            <p>📏 Max file: 10MB</p>
            <p>📋 Format: PDF only</p>
            <p>🤖 AI: Gemini 2.5 Flash</p>
          </div>
        </div>
      </aside>

      <main className={styles.main}>
        {activeTab === "uploader" ? <PdfUpload /> : <PdfManager />}
      </main>
    </div>
  );
}
