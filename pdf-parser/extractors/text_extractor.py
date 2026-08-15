import pdfplumber
from .pdf_utils import find_pages_with_keywords, extract_text_from_pages


def extract_risk_factors(file_path: str) -> str:
    """
    Finds the 'Risk Factors' section and extracts the first few pages of it.
    """
    keywords = ["risk factors", "internal risks", "external risks"]
    with pdfplumber.open(file_path) as pdf:
        pages = find_pages_with_keywords(pdf, keywords)
        if not pages:
            return ""
        start_page = pages[0]
        return extract_text_from_pages(pdf, start_page, num_pages=3)


def extract_objects_of_issue(file_path: str) -> str:
    """
    Finds the 'Objects of the Issue' or 'Use of Proceeds' section and extracts it.
    """
    keywords = ["objects of the issue", "objects of the offer", "use of proceeds"]
    with pdfplumber.open(file_path) as pdf:
        pages = find_pages_with_keywords(pdf, keywords)
        if not pages:
            return ""
        start_page = pages[0]
        return extract_text_from_pages(pdf, start_page, num_pages=2)


def extract_promoter_background(file_path: str) -> str:
    """
    Finds the 'Our Promoters' section and extracts the text.
    """
    keywords = ["our promoters and promoter group", "our promoters", "history and certain corporate matters"]
    with pdfplumber.open(file_path) as pdf:
        pages = find_pages_with_keywords(pdf, keywords)
        if not pages:
            return ""
        start_page = pages[0]
        return extract_text_from_pages(pdf, start_page, num_pages=2)
