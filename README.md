# 📄 PDF AI Summarizer

Aplikasi web untuk meringkas file PDF secara otomatis menggunakan AI Gemini. Dibangun dengan Next.js (frontend) dan FastAPI (backend).

## 🏗️ Arsitektur

- **Frontend** → Next.js 16 dengan React 19 (App Router)
- **Backend** → FastAPI dengan Python 3.10+
- **AI Engine** → Google Gemini 2.5 Flash
- **Komunikasi** → HTTP REST API (POST /summarize)

## 📋 Prasyarat

- Node.js 18+ dan npm
- Python 3.10+ dengan pip
- Google AI Studio Account (daftar di [aistudio.google.com](https://aistudio.google.com))
- Git (opsional)

> ⚠️ **Penting:** API key tidak boleh disimpan di repository

## 🚀 Quick Start

### 1️⃣ Setup Backend (FastAPI)

```bash
# Masuk ke folder backend
cd pdf-ai-summarizer

# Buat virtual environment (opsional tapi recommended)
python -m venv venv
venv\Scripts\activate  # Windows
# atau
source venv/bin/activate  # macOS/Linux

# Install dependencies
pip install -r requirements.txt

# Buat file .env
# Isi: GEMINI_API_KEY=your_api_key_here
# Jangan commit file ini!

# Jalankan backend
python -m uvicorn main:app --reload
```

✅ Backend running di `http://127.0.0.1:8000`

### 2️⃣ Setup Frontend (Next.js)

```bash
# Masuk ke folder frontend
cd pdf-ai-frontend

# Install dependencies
npm install

# Jalankan development server
npm run dev
```

✅ Frontend running di `http://localhost:3000`

## 💡 Cara Menggunakan

1. Buka browser → `http://localhost:3000`
2. Klik area upload atau tombol "Upload File"
3. Pilih file PDF dari komputer
4. Klik tombol "Summarize PDF"
5. Tunggu proses selesai (2-10 detik)
6. Lihat hasil ringkasan di layar

## 📁 Struktur Project

```
PDFPROJECTS/
├── pdf-ai-frontend/              # Next.js App
│   ├── src/app/
│   │   ├── components/
│   │   │   ├── pdfupload.js
│   │   │   └── pdfupload.module.css
│   │   ├── globals.css
│   │   ├── layout.js
│   │   └── page.js
│   ├── package.json
│   └── next.config.mjs
│
├── pdf-ai-summarizer/            # FastAPI Backend
│   ├── main.py
│   ├── .env                       # ⚠️ Git ignored (local only)
│   ├── .env.example               # Template
│   ├── requirements.txt
│   └── venv/
│
└── README.md                      # Dokumentasi
```

## 🔌 API Endpoint

### POST `/summarize`

Mengirim file PDF untuk diringkas.

**Request:**
```bash
curl -X POST "http://127.0.0.1:8000/summarize" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@document.pdf"
```

**Response (200 OK):**
```json
{
  "provider": "gemini",
  "summary": "Ringkasan PDF Anda di sini..."
}
```

**Response (400 Bad Request):**
```json
{
  "detail": "PDF tidak mengandung teks"
}
```

## 🛠️ Tech Stack

| Komponen | Technology | Version |
|----------|-----------|---------|
| Frontend | Next.js | 16.1.1 |
| UI Library | React | 19.0.0 |
| Styling | CSS Modules + Poppins Font | - |
| Backend | FastAPI | 0.104+ |
| PDF Parser | PyPDF2 | 3.0+ |
| AI Model | Gemini | 2.5 Flash |
| Python | Python | 3.10+ |

## 🔐 Keamanan

### Environment Variables

**File `.env` (JANGAN commit):**
```env
GEMINI_API_KEY=your_google_ai_key_here
```

**File `.env.example` (untuk dokumentasi):**
```env
GEMINI_API_KEY=your_key_here
```

### Best Practices

✅ Simpan API key di `.env` lokal  
✅ Gunakan `.env.example` untuk template  
✅ Tambahkan `.env` ke `.gitignore`  
✅ Jangan pernah share API key  
✅ Rotasi key secara berkala  

## ❌ Troubleshooting

### Backend tidak bisa diakses
```bash
# Pastikan backend sedang running
python -m uvicorn main:app --reload

# Cek di http://127.0.0.1:8000/docs
```

### CORS Error
Backend sudah dikonfigurasi untuk allow localhost:3000. Jika error masih terjadi:
- Restart backend
- Clear browser cache
- Cek console browser untuk detail error

### "API Key Invalid"
- Pastikan sudah register di [Google AI Studio](https://aistudio.google.com)
- Copy key dengan benar (tanpa spasi)
- Restart backend setelah update `.env`

### "ModuleNotFoundError"
```bash
# Aktivasi virtual environment
venv\Scripts\activate

# Install lagi
pip install -r requirements.txt
```

### Port sudah digunakan
```bash
# Frontend (change port)
npm run dev -- -p 3001

# Backend (change port)
uvicorn main:app --reload --port 8001
```

## 📝 Development

### Struktur Komponen Frontend

**pdfupload.js** - Main component yang handle:
- Upload file PDF
- Display loading state
- Show error message
- Display summary hasil

**pdfupload.module.css** - Styling dengan:
- CSS Modules untuk isolation
- Responsive design (mobile-first)
- Font Poppins dari Google Fonts

### Backend Flow

1. **extract_text_from_pdf()** → Parse PDF dan ekstrak text
2. **summarize_with_gemini()** → Kirim ke Gemini API
3. **@app.post("/summarize")** → Handle request/response

### Testing API

**Gunakan Swagger UI:**
```
http://127.0.0.1:8000/docs
```

Upload file PDF langsung dari browser untuk test.

## 📦 Dependencies

### Backend (`requirements.txt`)
```
fastapi==0.104.1
uvicorn==0.24.0
PyPDF2==3.0.1
google-generativeai==0.3.0
python-dotenv==1.0.0
python-multipart==0.0.6
```

### Frontend (`package.json`)
```json
{
  "dependencies": {
    "next": "^16.1.1",
    "react": "^19.0.0"
  }
}
```

## 🚀 Production Deployment

> 📌 Untuk production, perlu beberapa konfigurasi tambahan

**Backend:**
- Gunakan production server (Gunicorn + Uvicorn)
- Setup CORS dengan domain spesifik
- Add request validation & rate limiting

**Frontend:**
- Run `npm run build`
- Deploy ke Vercel / Netlify
- Setup environment variables di hosting provider

## 📚 Learning Resources

- [Next.js Docs](https://nextjs.org/docs)
- [FastAPI Tutorial](https://fastapi.tiangolo.com/)
- [PyPDF2 Guide](https://pypdf.readthedocs.io/)
- [Google Gemini API](https://ai.google.dev/)
- [React Hooks](https://react.dev/reference/react)

## 🤝 Kontribusi

Punya ide? Ingin fix bug?

1. Fork repository
2. Buat branch feature (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add feature'`)
4. Push ke branch (`git push origin feature/amazing`)
5. Buat Pull Request

## 📄 License

MIT License - Bebas untuk digunakan dan dimodifikasi

## 📞 Support

Jika ada pertanyaan atau menemukan bug:
- Buat issue di GitHub
- Check error di console browser/terminal
- Review dokumentasi di `/docs` endpoint backend

---

**Made with ❤️ using Next.js & FastAPI**
