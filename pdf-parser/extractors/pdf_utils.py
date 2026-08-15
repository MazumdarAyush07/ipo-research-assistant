import pdfplumber
from typing import List, Dict

import logging

logger = logging.getLogger(__name__)


def find_pages_with_keywords(pdf, keyword_weights: Dict[str, int]) -> List[int]:
    """
    Scans the open pdf object and ranks pages based on keyword weight density.
    Returns the page numbers sorted by highest density first.
    """
    page_scores = []
    LARGE_PDF_THRESHOLD = 800  # Pages above this get adaptive sampling
    MAX_PAGES_TO_SCAN = 800    # Max pages sampled for huge docs
    try:
        total_pages = len(pdf.pages)
        if total_pages > LARGE_PDF_THRESHOLD:
            # Huge PDF: sample every Nth page so we never read more than
            # MAX_PAGES_TO_SCAN pages total.
            step = max(1, total_pages // MAX_PAGES_TO_SCAN)
            logger.info(f"Large PDF ({total_pages} pages). Sampling every {step} pages.")
        else:
            # Normal PDF: full scan
            step = 1

        for i in range(0, total_pages, step):
            page = pdf.pages[i]
            text = page.extract_text()
            page.flush_cache()
            if not text:
                continue
            text_flat = text.lower().replace('\n', ' ')
            score = 0
            if isinstance(keyword_weights, dict):
                for kw, weight in keyword_weights.items():
                    if kw.lower() in text_flat:
                        score += weight
            else:
                for kw in keyword_weights:
                    if kw.lower() in text_flat:
                        score += 1
            if score > 0:
                if step > 1:
                    # Include surrounding pages so we don't miss context
                    for offset in range(step):
                        neighbour = i + offset
                        if neighbour < total_pages:
                            page_scores.append((neighbour, score))
                else:
                    page_scores.append((i, score))
    except Exception as e:
        logger.error(f"Error scanning PDF: {e}")

    # Deduplicate, then sort pages by score descending, page number ascending
    seen = set()
    unique_scores = []
    for page_num, score in page_scores:
        if page_num not in seen:
            seen.add(page_num)
            unique_scores.append((page_num, score))
    unique_scores.sort(key=lambda x: (-x[1], x[0]))
    return [p[0] for p in unique_scores]


def extract_text_from_pages(pdf, start_page: int, num_pages: int) -> str:
    """
    Extracts text from a sequence of pages using an already-open pdf object.
    """
    extracted_text = []
    try:
        total_pages = len(pdf.pages)
        end_page = min(start_page + num_pages, total_pages)
        for i in range(start_page, end_page):
            page = pdf.pages[i]
            text = page.extract_text(layout=True)
            page.flush_cache()
            if text:
                extracted_text.append(text)
    except Exception as e:
        logger.error(f"Error extracting text: {e}")

    return "\n".join(extracted_text)
