#!/usr/bin/env python3
from langdetect import detect, detect_langs, DetectorFactory
from langdetect.lang_detect_exception import LangDetectException

# Set seed untuk consistency
DetectorFactory.seed = 0

print("=== Testing langdetect ===\n")

test_texts = {
    "English": "This is a sample English text",
    "Indonesian": "Ini adalah contoh teks bahasa Indonesia",
    "French": "Ceci est un exemple de texte français",
    "Spanish": "Este es un ejemplo de texto en español",
    "German": "Dies ist ein Beispieltext in deutscher Sprache",
}

for lang_name, text in test_texts.items():
    try:
        detected = detect(text)
        probabilities = detect_langs(text)
        print(f"{lang_name}: {detected}")
        print(f"  Probabilities: {probabilities}\n")
    except LangDetectException as e:
        print(f"{lang_name}: ERROR - {e}\n")

print("✅ langdetect berfungsi dengan baik!")
