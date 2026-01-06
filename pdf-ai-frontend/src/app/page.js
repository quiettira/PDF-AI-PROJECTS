"use client";

import { useState } from "react";
import Sidebar from "./components/Sidebar";
import PdfUploader from "./components/pdfupload";
import PdfManager from "./components/PdfManager";

export default function Home() {
  const [activeTab, setActiveTab] = useState("uploader");

  return (
    <div style={{ display: "flex", minHeight: "100vh", backgroundColor: "#f8f9fa" }}>
      <Sidebar activeTab={activeTab} setActiveTab={setActiveTab} />
      
      <main style={{ 
        flex: 1, 
        padding: "30px", 
        backgroundColor: "#f8f9fa",
        overflowY: "auto"
      }}>
        {activeTab === "uploader" && <PdfUploader />}
        {activeTab === "manager" && <PdfManager />}
      </main>
    </div>
  );
}
