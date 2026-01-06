"use client";

import styles from "./sidebar.module.css";

export default function Sidebar({ activeTab, setActiveTab }) {
  return (
    <aside className={styles.sidebar}>
      {/* Header */}
      <div className={styles.header}>
        <h1 className={styles.title}>PDF ARAI</h1>
        <p className={styles.subtitle}>Asisten Ringkasan AI</p>
      </div>

      {/* Navigation Menu */}
      <nav className={styles.nav}>
        <button
          onClick={() => setActiveTab("uploader")}
          className={`${styles.navItem} ${activeTab === "uploader" ? styles.active : ""}`}
        >
          <span className={styles.icon} aria-hidden="true">
            📄
          </span>
          <span className={styles.label}>Summarizer</span>
        </button>
        
        <button
          onClick={() => setActiveTab("manager")}
          className={`${styles.navItem} ${activeTab === "manager" ? styles.active : ""}`}
        >
          <span className={styles.icon} aria-hidden="true">
            📁
          </span>
          <span className={styles.label}>Manager</span>
        </button>
        <div className={styles.header}>
          <h1 className={styles.title}>Info ARAI</h1>
          <p className={styles.subtitle}>PDF only</p>
          <p className={styles.subtitle}>10 MB max size</p>
          <p className={styles.subtitle}>Powered by Gemini 2.5 Flash</p>
        </div>
      </nav>
    </aside>
  );
}