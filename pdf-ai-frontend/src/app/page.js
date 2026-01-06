"use client";

import { useState } from "react";
import Sidebar from "./components/Sidebar";
import PdfUploader from "./components/pdfupload";
import PdfManager from "./components/PdfManager";
import styles from "./page.module.css";

export default function Home() {
  const [activeTab, setActiveTab] = useState("uploader");

  return (
    <div className={styles.container}>
      <Sidebar activeTab={activeTab} setActiveTab={setActiveTab} />
      
      <main className={styles.main}>
        {activeTab === "uploader" && <PdfUploader />}
        {activeTab === "manager" && <PdfManager />}
      </main>
    </div>
  );
}
