import pdfplumber
from typing import List, Optional, Tuple, Dict

import logging

logger = logging.getLogger(__name__)

def find_pages_with_keywords(file_path: str, keyword_weights: Dict[str, int]) -> List[int]:
    """
    Scans the document starting from 50% and ranks pages based on keyword weight density.
    Returns the page numbers sorted by highest density first.
    """
    page_scores = []
    try:
        with pdfplumber.open(file_path) as pdf:
            total_pages = len(pdf.pages)
            # Scan the entire document to ensure we don't miss tables placed early in the PDF
            start_from = 0
            
            pages_since_high_score = 0
            found_financials = False
            
            for i in range(start_from, total_pages):
                page = pdf.pages[i]
                text = page.extract_text()
                if text:
                    text_flat = text.lower().replace('\n', ' ')
                    score = 0
                    
                    if isinstance(keyword_weights, dict):
                        for kw, weight in keyword_weights.items():
                            if kw.lower() in text_flat:
                                score += weight
                    else:
                        # Fallback for backwards compatibility with lists
                        for kw in keyword_weights:
                            if kw.lower() in text_flat:
                                score += 1
                            
                    if score > 0:
                        page_scores.append((i, score))
                        
    except Exception as e:
        logger.error(f"Error scanning PDF {file_path}: {e}")
        
    # Sort pages by score descending, then by page number ascending
    page_scores.sort(key=lambda x: (-x[1], x[0]))
    
    return [p[0] for p in page_scores]

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
                text = page.extract_text(layout=True)
                if text:
                    extracted_text.append(text)
    except Exception as e:
        logger.error(f"Error extracting text from {file_path}: {e}")
        
    return "\n".join(extracted_text)
