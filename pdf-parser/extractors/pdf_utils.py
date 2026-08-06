import pdfplumber
from typing import List, Optional, Tuple

def find_pages_with_keywords(file_path: str, keywords: List[str], max_pages_to_scan: int = 500) -> List[int]:
    """
    Scans a PDF and returns a list of page numbers (0-indexed) that contain at least one of the keywords.
    """
    found_pages = []
    try:
        with pdfplumber.open(file_path) as pdf:
            total_pages = min(len(pdf.pages), max_pages_to_scan)
            for i in range(total_pages):
                page = pdf.pages[i]
                text = page.extract_text()
                if text:
                    text_lower = text.lower()
                    for keyword in keywords:
                        if keyword.lower() in text_lower:
                            found_pages.append(i)
                            break # Move to next page if any keyword is found
    except Exception as e:
        print(f"Error scanning PDF {file_path}: {e}")
    
    return found_pages

def extract_text_from_pages(file_path: str, start_page: int, num_pages: int) -> str:
    """
    Extracts text from a sequence of pages starting from start_page.
    """
    extracted_text = []
    try:
        with pdfplumber.open(file_path) as pdf:
            total_pages = len(pdf.pages)
            end_page = min(start_page + num_pages, total_pages)
            for i in range(start_page, end_page):
                page = pdf.pages[i]
                text = page.extract_text()
                if text:
                    extracted_text.append(text)
    except Exception as e:
        print(f"Error extracting text from {file_path}: {e}")
        
    return "\n".join(extracted_text)
