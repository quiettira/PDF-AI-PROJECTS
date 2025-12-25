"use client";
import styles from "./navbar.module.css";

export default function Navbar({ sidebarOpen, setSidebarOpen }) {
  return (
    <nav className={styles.navbar}>
      <div className={styles.navbarInner}>
        {/* kiri */}
        <div className={styles.left}>
          <button
            className={styles.menuBtn}
            onClick={() => setSidebarOpen(!sidebarOpen)}
            aria-label="Toggle Sidebar"
          >
            {sidebarOpen ? "✕" : "☰"}
          </button>

          <span className={styles.brand}>ARAI</span>
        </div>

        {/* kanan */}
        <div className={styles.right}>
          <span className={styles.lang}>ID</span>
          <span className={styles.plus}>✨ Plus</span>
          <button className={styles.register}>Daftar</button>
        </div>
      </div>
    </nav>
  );
}
