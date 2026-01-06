import os
import io
import re
from html import escape
from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.middleware.cors import CORSMiddleware   
from PyPDF2 import PdfReader
from dotenv import load_dotenv
from langdetect import detect, DetectorFactory
from langdetect.lang_detect_exception import LangDetectException
from fastapi.responses import PlainTextResponse, StreamingResponse
from reportlab.platypus import SimpleDocTemplate, Paragraph
from reportlab.lib.styles import getSampleStyleSheet
from reportlab.lib.pagesizes import A4
from pydantic import BaseModel
from fastapi import Body


# Set seed untuk consistency dalam language detection
DetectorFactory.seed = 0

load_dotenv()

app = FastAPI(title="PDF AI Summarizer")

# =========================
# ✅ TAMBAHKAN CORS DI SINI (WAJIB)
# =========================

cors_origins_env = os.getenv("CORS_ORIGINS", "http://localhost:3000")
cors_origins = [o.strip() for o in cors_origins_env.split(",") if o.strip()]

app.add_middleware(
    CORSMiddleware,
    allow_origins=cors_origins if cors_origins != ["*"] else ["*"],
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
    import google.generativeai as genai
    genai.configure(api_key=gemini_api_key)

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
            "id": "Format ringkasan sebagai poin-poin (bullet points) yang mudah dicerna:\n- Gunakan format bullet (•) atau dash (-) untuk setiap poin utama\n- Setiap poin maksimal 1-2 baris\n- Kelompokkan poin-poin yang related dengan subheading jika perlu\n- WAJIB gunakan format bullet points, JANGAN paragraf\n- Contoh format:\n• Poin pertama\n• Poin kedua\n• Poin ketiga",
            "en": "Format the summary as bullet points that are easy to digest:\n- Use bullet (•) or dash (-) format for each main point\n- Each point should be 1-2 lines maximum\n- Group related points with subheadings if needed\n- MUST use bullet point format, NOT paragraphs\n- Example format:\n• First point\n• Second point\n• Third point"
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

PENTING: Gunakan format Markdown bold (**kata**) untuk menyorot (highlight) kata kunci, nama penting, atau poin utama agar pembaca lebih mudah menangkap inti sari.

Dokumen:
{text[:5000]}
"""
    else:
        prompt = f"""
Please summarize the following document in English.

Format instructions:
{instruction}

IMPORTANT: Use Markdown bold (**word**) to highlight key terms, important names, or main points so the reader can easily grasp the essence.

Document:
{text[:5000]}
"""
    
    return prompt

# =========================
# Summarize Logic
# =========================
def summarize_with_gemini(text: str, language: str, style: str = "standard"):
    prompt = get_summarize_prompt(text, language, style)

    model = genai.GenerativeModel("gemini-2.5-flash")
    
    # Konfigurasi generation dengan temperature 0.3 untuk konsistensi yang baik
    generation_config = genai.types.GenerationConfig(
        temperature=0.3,
        top_p=0.8,
        top_k=40,
        max_output_tokens=2048,
    )
    
    response = model.generate_content(
        prompt,
        generation_config=generation_config
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

class ExportRequest(BaseModel):
    content: str

@app.post("/export/txt")
async def export_txt(summary: str = Body(..., embed=True)):
    # Bersihkan markdown formatting untuk file TXT agar rapi
    clean_summary = re.sub(r'\*\*(.*?)\*\*', r'\1', summary)  # Hapus bold **kata** -> kata
    clean_summary = re.sub(r'\*(.*?)\*', r'\1', clean_summary)     # Hapus italic *kata* -> kata
    
    return PlainTextResponse(
        content=clean_summary,
        headers={
            "Content-Disposition": "attachment; filename=summary.txt"
        }
    )

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "provider": AI_PROVIDER,
        "version": "1.0.0"
    }

@app.post("/extract-text")
async def extract_text_endpoint(file: UploadFile = File(...)):
    """Extract text from PDF without summarization"""
    text = extract_text_from_pdf(file)
    detected_language = detect_language(text)
    
    return {
        "text": text,
        "detected_language": detected_language,
        "length": len(text)
    }

@app.post("/export/pdf")
async def export_pdf(summary: str = Body(..., embed=True)):
    try:
        buffer = io.BytesIO()
        doc = SimpleDocTemplate(buffer, pagesize=(595.27, 841.89))  # A4 size
        styles = getSampleStyleSheet()
        story = []

        def clean_markdown_for_pdf(text):
            """Clean markdown formatting for PDF generation"""
            if not text:
                return "No content available"
            
            # Remove markdown bold formatting **text** -> text
            text = re.sub(r'\*\*(.*?)\*\*', r'<b>\1</b>', text)
            
            # Remove markdown italic formatting *text* -> text  
            text = re.sub(r'\*(.*?)\*', r'<i>\1</i>', text)
            
            # Handle bullet points properly
            lines = text.split('\n')
            formatted_lines = []
            
            for line in lines:
                line = line.strip()
                if not line:
                    formatted_lines.append('<br/>')
                    continue
                    
                # Convert bullet points
                if line.startswith('•') or line.startswith('-') or line.startswith('*'):
                    # Remove the bullet and add proper formatting
                    clean_line = re.sub(r'^[•\-\*]\s*', '', line)
                    formatted_lines.append(f'• {clean_line}')
                else:
                    formatted_lines.append(line)
            
            return '<br/>'.join(formatted_lines)

        # Clean and format the summary
        formatted_summary = clean_markdown_for_pdf(summary)
        
        # Add title
        title_style = styles['Title']
        title_style.fontSize = 16
        title_style.spaceAfter = 20
        story.append(Paragraph("PDF Summary", title_style))
        
        # Add content with proper formatting
        normal_style = styles['Normal']
        normal_style.fontSize = 11
        normal_style.leading = 14
        normal_style.spaceAfter = 6
        
        # Split into paragraphs and add each one
        paragraphs = formatted_summary.split('<br/><br/>')
        for para in paragraphs:
            if para.strip():
                story.append(Paragraph(para, normal_style))

        # Build PDF
        doc.build(story)
        buffer.seek(0)

        return StreamingResponse(
            io.BytesIO(buffer.getvalue()),
            media_type="application/pdf",
            headers={
                "Content-Disposition": "attachment; filename=summary.pdf",
                "Content-Length": str(len(buffer.getvalue()))
            }
        )
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"PDF generation failed: {str(e)}")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)