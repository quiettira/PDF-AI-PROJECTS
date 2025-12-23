import os
import io
from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.middleware.cors import CORSMiddleware   # ✅ TAMBAH INI
from PyPDF2 import PdfReader
from dotenv import load_dotenv

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
# Summarize Logic
# =========================
def summarize_with_gemini(text: str):
    prompt = f"""
    Please summarize the following document in clear, concise paragraphs.
    Highlight key ideas and structure the summary for readability.

    Document:
    {text[:5000]}
    """

    response = client.models.generate_content(
        model="gemini-2.5-flash",   
        contents=prompt
    )

    return response.text

def summarize_mock(text: str):
    return text[:300] + " ... (mock summary)"

# =========================
# API Endpoint
# =========================
@app.post("/summarize")
async def summarize_pdf(file: UploadFile = File(...)):
    text = extract_text_from_pdf(file)

    try:
        if AI_PROVIDER == "gemini":
            summary = summarize_with_gemini(text)
        else:
            summary = summarize_mock(text)

        return {
            "provider": AI_PROVIDER,
            "summary": summary
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
