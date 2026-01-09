"use client";

import { useState } from "react";
import PdfUpload from "./components/pdfupload";
import PdfManager from "./components/PdfManager";
import Sidebar from "./components/Sidebar";

export default function Home() {
  const [activeTab, setActiveTab] = useState("uploader");

  return (
    <div style={{ display: "flex", minHeight: "100vh" }}>
      <Sidebar activeTab={activeTab} setActiveTab={setActiveTab} />
      <main style={{ flex: 1, overflow: "auto" }}>
        {activeTab === "uploader" ? <PdfUpload /> : <PdfManager />}
      </main>
    </div>
  );
}
