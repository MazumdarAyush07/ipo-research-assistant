import camelot
import pandas as pd
from typing import List, Dict, Any, Optional
from .pdf_utils import find_pages_with_keywords, extract_text_from_pages
import re

def parse_financials(file_path: str) -> List[Dict[str, Any]]:
    """
    Attempts to extract financials using Camelot. If it fails, falls back to raw text parsing.
    Returns a list of dicts: [{'year': 2023, 'revenue': 100, ...}]
    """
    keywords = [
        "restated consolidated statement of profit and loss",
        "restated consolidated statement of assets and liabilities",
        "restated statement of profit and loss",
        "restated statement of assets and liabilities"
    ]
    
    pages = find_pages_with_keywords(file_path, keywords)
    if not pages:
        return []
    
    financials_data = []
    
    try:
        # Camelot expects 1-indexed string pages like "10,11,12"
        page_str = ",".join([str(p + 1) for p in pages[:3]]) # Try first 3 matched pages
        tables = camelot.read_pdf(file_path, pages=page_str, flavor='stream')
        
        if tables.n > 0:
            df = tables[0].df
            financials_data = _parse_dataframe(df)
            
        if not financials_data:
            raise ValueError("Camelot returned empty data or failed to parse")
            
    except Exception as e:
        print(f"Camelot extraction failed: {e}. Falling back to text extraction.")
        # Fallback to pdfplumber raw text extraction
        raw_text = extract_text_from_pages(file_path, pages[0], 3)
        financials_data = _parse_raw_text(raw_text)

    return financials_data

def _parse_dataframe(df: pd.DataFrame) -> List[Dict[str, Any]]:
    """
    Very basic heuristic to extract data from a dataframe.
    """
    # This is highly DRHP-dependent.
    # We look for rows that contain "Revenue from operations", "Total Income", "Profit for the year"
    
    # Initialize a dummy response for now
    # We will need to map columns to years (typically the columns to the right of the label)
    results = [
        {"year": 2023, "revenue": 0.0, "pat": 0.0, "ebitda": 0.0, "total_assets": 0.0, "total_debt": 0.0, "equity": 0.0},
        {"year": 2022, "revenue": 0.0, "pat": 0.0, "ebitda": 0.0, "total_assets": 0.0, "total_debt": 0.0, "equity": 0.0},
        {"year": 2021, "revenue": 0.0, "pat": 0.0, "ebitda": 0.0, "total_assets": 0.0, "total_debt": 0.0, "equity": 0.0}
    ]
    
    return results

def _parse_raw_text(text: str) -> List[Dict[str, Any]]:
    """
    Fallback: parse raw text with regex to find financial numbers.
    """
    results = [
        {"year": 2023, "revenue": 0.0, "pat": 0.0, "ebitda": 0.0, "total_assets": 0.0, "total_debt": 0.0, "equity": 0.0},
        {"year": 2022, "revenue": 0.0, "pat": 0.0, "ebitda": 0.0, "total_assets": 0.0, "total_debt": 0.0, "equity": 0.0},
        {"year": 2021, "revenue": 0.0, "pat": 0.0, "ebitda": 0.0, "total_assets": 0.0, "total_debt": 0.0, "equity": 0.0}
    ]
    
    # Basic regex example: look for "Revenue from operations" followed by numbers
    # This is a very rough heuristic
    lines = text.split('\n')
    for line in lines:
        line_lower = line.lower()
        if "revenue from operations" in line_lower or "total income" in line_lower:
            numbers = re.findall(r'[\d,]+\.?\d*', line)
            # Assuming the numbers are the last 3 columns (most recent year first or last)
            # Just a placeholder heuristic
            if len(numbers) >= 3:
                try:
                    results[0]['revenue'] = float(numbers[-3].replace(',', ''))
                    results[1]['revenue'] = float(numbers[-2].replace(',', ''))
                    results[2]['revenue'] = float(numbers[-1].replace(',', ''))
                except:
                    pass
                    
        if "profit for the year" in line_lower or "profit after tax" in line_lower:
            numbers = re.findall(r'[\d,]+\.?\d*', line)
            if len(numbers) >= 3:
                try:
                    results[0]['pat'] = float(numbers[-3].replace(',', ''))
                    results[1]['pat'] = float(numbers[-2].replace(',', ''))
                    results[2]['pat'] = float(numbers[-1].replace(',', ''))
                except:
                    pass
                    
    return results
