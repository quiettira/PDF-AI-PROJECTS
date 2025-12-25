"use client";

import { useState } from "react";
import Navbar from "./components/Navbar";
import Sidebar from "./components/Sidebar";
import PdfUploader from "./components/pdfupload";

export default function Home() {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <>
      <Navbar
        sidebarOpen={sidebarOpen}
        setSidebarOpen={setSidebarOpen}
      />

      <div style={{ display: "flex" }}>
        {sidebarOpen && <Sidebar />}

        <main style={{ flex: 1 }}>
          <PdfUploader />
        </main>
      </div>
    </>
  );
}
