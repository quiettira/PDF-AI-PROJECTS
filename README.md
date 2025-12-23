# PDF AI Summarizer

Aplikasi web untuk meringkas file PDF secara otomatis menggunakan AI Gemini. Dibangun dengan Next.js (frontend) dan FastAPI (backend).

## 🏗️ Arsitektur

- **Frontend** → Next.js 16 dengan React 19 (App Router)
- **Backend** → FastAPI dengan Python 3.10+
- **AI Engine** → Google Gemini 2.5 Flash
- **Komunikasi** → HTTP REST API (POST /summarize)

## Prasyarat

- Node.js 18+ dan npm
- Python 3.10+
- Google AI Studio API Key ([daftar di sini](https://aistudio.google.com))

> ⚠️ **Penting:** API key tidak boleh disimpan di repository

### Backend (FastAPI)

```bash
cd pdf-ai-summarizer

# Buat virtual environment
python -m venv venv
venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Buat file .env
# GEMINI_API_KEY=your_api_key_here

# Jalankan backend
python -m uvicorn main:app --reload
```

Backend berjalan di: `http://127.0.0.1:8000`

### Frontend (Next.js)

```bash
cd pdf-ai-frontend

# Install & jalankan
npm install
npm run dev
```

Frontend berjalan di: `http://localhost:3000`

## Cara Pakai

1. Buka `http://localhost:3000`
2. Upload file PDF
3. Klik "Summarize PDF"
4. Tunggu hasil ringkasan

## Troubleshooting

### Folder frontend tidak muncul di GitHub
**Problem:** Frontend folder hanya muncul sebagai folder kosong di GitHub (tidak bisa diklik).

**Solusi:** Folder dianggap sebagai git submodule. Jalankan perintah ini di folder project utama:

```bash
git rm --cached pdf-ai-frontend
git add pdf-ai-frontend/
git commit -m "Fix: Add frontend files properly"
git push origin main
```

### Backend tidak bisa diakses
```bash
# Pastikan backend sedang running
python -m uvicorn main:app --reload
```

### CORS Error
- Restart backend
- Clear browser cache

### "API Key Invalid"
- Pastikan API key benar di file `.env`
- Restart backend setelah update `.env`

### AI Provider tidak mendukung
**Problem:** Beberapa AI provider (Claude, OpenAI, dll) memerlukan payment card atau tidak support API gratis.

**Solusi:** Gunakan **Google Gemini 2.5 Flash** yang gratis namun dengan batasan:
- ✅ Gratis tanpa perlu payment card
- ⚠️ **Limit: 5 request per hari** (free tier)
- ✅ Hasil ringkasan cukup bagus untuk PDF

Pastikan menggunakan:
```bash
# Di file .env
GEMINI_API_KEY=your_google_ai_key_here

# Di file main.py backend
model = genai.GenerativeModel('gemini-2.5-flash')
```

Jika sudah exceed limit harian, tunggu hingga 24 jam untuk reset atau upgrade ke plan berbayar di [Google AI Studio](https://aistudio.google.com).

---

**Made with love, thanks**
