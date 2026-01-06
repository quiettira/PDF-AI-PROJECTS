#!/usr/bin/env python3
"""
Script untuk melihat daftar model yang tersedia di Ollama
"""

import requests
import json

def list_ollama_models():
    """List all available models in Ollama"""
    try:
        response = requests.get("http://localhost:11434/api/tags")
        if response.status_code == 200:
            models = response.json()
            print("Available Ollama Models:")
            print("-" * 40)
            
            if 'models' in models and models['models']:
                for model in models['models']:
                    name = model.get('name', 'Unknown')
                    size = model.get('size', 0)
                    modified = model.get('modified_at', 'Unknown')
                    
                    # Convert size to human readable format
                    size_gb = size / (1024**3) if size > 0 else 0
                    
                    print(f"Name: {name}")
                    print(f"Size: {size_gb:.2f} GB")
                    print(f"Modified: {modified}")
                    print("-" * 40)
            else:
                print("No models found. Make sure Ollama is running and has models installed.")
                
        else:
            print(f"Error connecting to Ollama: {response.status_code}")
            print("Make sure Ollama is running on localhost:11434")
            
    except requests.exceptions.ConnectionError:
        print("Error: Cannot connect to Ollama server.")
        print("Make sure Ollama is running on localhost:11434")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    list_ollama_models()