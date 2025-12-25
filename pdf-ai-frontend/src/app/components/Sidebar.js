import styles from "./sidebar.module.css";

export default function Sidebar() {
  return (
    <aside className={styles.sidebar}>
      <div className={`${styles.item} ${styles.active}`}>💬 Chat baru</div>
      <div className={styles.item}>📄 Dokumen</div>
      <div className={styles.item}>⭐ Favorit</div>
      <div className={styles.item}>⚙️ Pengaturan</div>
    </aside>
  );
}
