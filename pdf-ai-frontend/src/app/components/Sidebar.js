"use client";

import { useState } from "react";
import styles from "./sidebar.module.css";

export default function Sidebar({ activeTab, setActiveTab }) {
  const [isOpen, setIsOpen] = useState(false);

  const handleNavClick = (tab) => {
    setActiveTab(tab);
    setIsOpen(false); // Close menu on mobile after selection
  };

  return (
    <>
      {/* Hamburger Button */}
      <button 
        className={styles.hamburger}
        onClick={() => setIsOpen(!isOpen)}
        aria-label="Toggle menu"
      >
        <span className={`${styles.hamburgerLine} ${isOpen ? styles.open : ""}`}></span>
        <span className={`${styles.hamburgerLine} ${isOpen ? styles.open : ""}`}></span>
        <span className={`${styles.hamburgerLine} ${isOpen ? styles.open : ""}`}></span>
      </button>

      {/* Overlay untuk mobile */}
      {isOpen && (
        <div 
          className={styles.overlay}
          onClick={() => setIsOpen(false)}
        />
      )}

      <aside className={`${styles.sidebar} ${isOpen ? styles.open : ""}`}>
        {/* Header */}
        <div className={styles.header}>
          <h1 className={styles.title}>PDF ARAI</h1>
          <p className={styles.subtitle}>Asisten Ringkasan AI</p>
        </div>

        {/* Navigation Menu */}
        <nav className={styles.nav}>
          <button
            onClick={() => handleNavClick("uploader")}
            className={`${styles.navItem} ${activeTab === "uploader" ? styles.active : ""}`}
          >
            <span className={styles.icon} aria-hidden="true">
              📄
            </span>
            <span className={styles.label}>Summarizer</span>
          </button>
          
          <button
            onClick={() => handleNavClick("manager")}
            className={`${styles.navItem} ${activeTab === "manager" ? styles.active : ""}`}
          >
            <span className={styles.icon} aria-hidden="true">
              📁
            </span>
            <span className={styles.label}>Manager</span>
          </button>
          <div className={styles.infoSection}>
            <h3 className={styles.infoTitle}>Info ARAI</h3>
            <div className={styles.infoContent}>
              <p>PDF only</p>
              <p>10 MB max size</p>
              <p>Powered by Gemini 2.5 Flash</p>
            </div>
          </div>
        </nav>
      </aside>
    </>
  );
}