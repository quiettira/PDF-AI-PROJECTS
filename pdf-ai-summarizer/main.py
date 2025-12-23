import os
import io
from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.middleware.cors import CORSMiddleware   # ✅ TAMBAH INI
from PyPDF2 import PdfReader
from dotenv import load_dotenv
from langdetect import detect, DetectorFactory
from langdetect.lang_detect_exception import LangDetectException

# Set seed untuk consistency dalam language detection
DetectorFactory.seed = 0

load_dotenv()

app = FastAPI(title="PDF AI Summarizer")

# =========================
# ✅ TAMBAHKAN CORS DI SINI (WAJIB)
# =========================
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000"],  # Next.js
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# =========================
# Detect API Keys
# =========================
gemini_api_key = os.getenv("GEMINI_API_KEY")
openai_api_key = os.getenv("OPENAI_API_KEY")

# =========================
# Determine Provider
# =========================
AI_PROVIDER = (
    "gemini" if gemini_api_key else
    "mock"
)

print(f"Active AI Provider: {AI_PROVIDER}")

# =========================
# Init Gemini (if available)
# =========================
if AI_PROVIDER == "gemini":
    from google import genai
    client = genai.Client(api_key=gemini_api_key)

# =========================
# PDF Text Extract
# =========================
def extract_text_from_pdf(upload_file: UploadFile):
    pdf_bytes = upload_file.file.read()
    reader = PdfReader(io.BytesIO(pdf_bytes))

    text = ""
    for page in reader.pages:
        if page.extract_text():
            text += page.extract_text() + "\n"

    if not text.strip():
        raise HTTPException(
            status_code=400,
            detail="PDF tidak mengandung teks"
        )

    return text

# =========================
# Detect Language
# =========================
def detect_language(text: str):
    """
    Deteksi bahasa dari teks.
    Returns: 'id' untuk Indonesian, 'en' untuk English, dll
    """
    try:
        # Ambil sample dari teks (hindari teks yang terlalu panjang)
        sample = text[:1000]
        detected_lang = detect(sample)
        return detected_lang
    except LangDetectException:
        # Default ke English jika deteksi gagal
        return "en"

# =========================
# Get Prompt by Language
# =========================
def get_summarize_prompt(text: str, language: str):
    """
    Buat prompt yang sesuai dengan bahasa yang terdeteksi
    """
    if language == "id":  # Indonesian
        prompt = f"""
Buatkan ringkasan dari dokumen berikut dalam bahasa Indonesia.
Ringkasan harus jelas, singkat, dan terstruktur dalam paragraf-paragraf.
Sorot ide-ide kunci dan jelaskan poin-poin penting.

Dokumen:
{text[:5000]}
"""
    else:  # English (default)
        prompt = f"""
Please summarize the following document in English.
The summary should be clear, concise, and well-structured in paragraphs.
Highlight key ideas and explain important points.

Document:
{text[:5000]}
"""
    
    return prompt

# =========================
# Summarize Logic
# =========================
def summarize_with_gemini(text: str, language: str):
    prompt = get_summarize_prompt(text, language)

    response = client.models.generate_content(
        model="gemini-2.5-flash",   
        contents=prompt
    )

    return response.text

def summarize_mock(text: str, language: str):
    lang_label = "Bahasa Indonesia" if language == "id" else "English"
    return f"[{lang_label}] {text[:300]} ... (mock summary)"

# =========================
# API Endpoint
# =========================
# =========================
# API Endpoint
# =========================
@app.post("/summarize")
async def summarize_pdf(file: UploadFile = File(...)):
    text = extract_text_from_pdf(file)
    
    # Deteksi bahasa dari teks PDF
    detected_language = detect_language(text)

    try:
        if AI_PROVIDER == "gemini":
            summary = summarize_with_gemini(text, detected_language)
        else:
            summary = summarize_mock(text, detected_language)

        return {
            "provider": AI_PROVIDER,
            "detected_language": detected_language,
            "summary": summary
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
