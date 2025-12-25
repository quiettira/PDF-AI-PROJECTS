import os
import io
from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.middleware.cors import CORSMiddleware   
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
        # Ambil sample dari teks 
        sample = text[:1000]
        detected_lang = detect(sample)
        return detected_lang
    except LangDetectException:
        # Default ke English jika deteksi gagal
        return "en"

# =========================
# Get Prompt by Language and Style
# =========================
def get_summarize_prompt(text: str, language: str, style: str = "standard"):
    """
    Buat prompt yang sesuai dengan bahasa dan gaya ringkasan yang dipilih
    
    Styles:
    - standard: Ringkasan paragraf normal
    - executive: Ringkasan untuk eksekutif (fokus pada hasil, impact)
    - bullets: Format poin-poin (bullet points)
    - detailed: Ringkasan detail dengan penjelasan mendalam
    """
    
    style_instructions = {
        "standard": {
            "id": "Ringkasan harus jelas, singkat, dan terstruktur dalam paragraf-paragraf.\nSorot ide-ide kunci dan jelaskan poin-poin penting.",
            "en": "The summary should be clear, concise, and well-structured in paragraphs.\nHighlight key ideas and explain important points."
        },
        "executive": {
            "id": "Buatkan ringkasan eksekutif yang fokus pada:\n- Apa masalahnya?\n- Solusi/rekomendasi utama\n- Impact atau hasil yang diharapkan\n\nGunakan bahasa yang ringkas dan actionable, cocok untuk decision makers.",
            "en": "Create an executive summary focusing on:\n- What is the main issue?\n- Key solutions/recommendations\n- Expected impact or results\n\nUse concise, actionable language suitable for decision makers."
        },
        "bullets": {
            "id": "Format ringkasan sebagai poin-poin (bullet points) yang mudah dicerna:\n- Setiap poin maksimal 1-2 baris\n- Gunakan bullet (•) atau nomor untuk setiap poin\n- Kelompokkan poin-poin yang related dengan subheading jika perlu",
            "en": "Format the summary as bullet points that are easy to digest:\n- Each point should be 1-2 lines maximum\n- Use bullets (•) or numbers for each point\n- Group related points with subheadings if needed"
        },
        "detailed": {
            "id": "Buatkan ringkasan detail yang mencakup:\n- Latar belakang/konteks\n- Poin-poin utama dengan penjelasan mendalam\n- Nuansa dan detail penting\n- Kesimpulan dan implikasi\n\nBisa lebih panjang untuk menangkap informasi yang lebih komprehensif.",
            "en": "Create a detailed summary that includes:\n- Background/context\n- Main points with deep explanation\n- Important nuances and details\n- Conclusions and implications\n\nCan be longer to capture more comprehensive information."
        }
    }
    
    lang = "id" if language == "id" else "en"
    instruction = style_instructions.get(style, style_instructions["standard"])[lang]
    
    if lang == "id":
        prompt = f"""
Buatkan ringkasan dari dokumen berikut dalam bahasa Indonesia.

Instruksi format:
{instruction}

Dokumen:
{text[:5000]}
"""
    else:
        prompt = f"""
Please summarize the following document in English.

Format instructions:
{instruction}

Document:
{text[:5000]}
"""
    
    return prompt

# =========================
# Summarize Logic
# =========================
def summarize_with_gemini(text: str, language: str, style: str = "standard"):
    prompt = get_summarize_prompt(text, language, style)

    response = client.models.generate_content(
        model="gemini-2.5-flash",   
        contents=prompt
    )

    return response.text

def summarize_mock(text: str, language: str, style: str = "standard"):
    lang_label = "Bahasa Indonesia" if language == "id" else "English"
    style_label = f" ({style})" if style != "standard" else ""
    return f"[{lang_label}{style_label}] {text[:300]} ... (mock summary)"

# =========================
# API Endpoints
# =========================

@app.post("/preview")
async def preview_pdf(file: UploadFile = File(...)):
    """Extract and return PDF text content for preview"""
    text = extract_text_from_pdf(file)
    
    # Deteksi bahasa dari teks PDF
    detected_language = detect_language(text)
    
    # Limit preview text to first 2000 characters
    preview_text = text[:2000]
    
    return {
        "detected_language": detected_language,
        "preview_text": preview_text,
        "total_length": len(text),
        "preview_length": len(preview_text),
        "is_truncated": len(text) > 2000
    }

@app.post("/summarize")
async def summarize_pdf(file: UploadFile = File(...), style: str = "standard"):
    """
    Summarize PDF with selected style.
    
    Query parameter:
    - style: standard, executive, bullets, or detailed (default: standard)
    """
    text = extract_text_from_pdf(file)
    
    # Deteksi bahasa dari teks PDF
    detected_language = detect_language(text)
    
    # Validasi style
    valid_styles = ["standard", "executive", "bullets", "detailed"]
    if style not in valid_styles:
        style = "standard"

    try:
        if AI_PROVIDER == "gemini":
            summary = summarize_with_gemini(text, detected_language, style)
        else:
            summary = summarize_mock(text, detected_language, style)

        return {
            "provider": AI_PROVIDER,
            "detected_language": detected_language,
            "style": style,
            "summary": summary
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
