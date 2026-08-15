import os
import re
import time
import logging
from typing import List, Dict, Any, Optional
from google import genai
from google.genai import errors as genai_errors
from pydantic import BaseModel, Field
from .pdf_utils import find_pages_with_keywords, extract_text_from_pages

logger = logging.getLogger(__name__)

# Initialize Gemini Client (Will use GEMINI_API_KEY from environment)
try:
    client = genai.Client()
except Exception as e:
    logger.error(f"Failed to initialize Gemini Client: {e}")
    client = None

class FinancialData(BaseModel):
    year: int
    revenue: float
    pat: float
    ebitda: float
    total_assets: float
    total_debt: float
    equity: float

class FinancialResponse(BaseModel):
    unit: str = Field(description="The unit of the financial numbers. One of: 'Lakhs', 'Millions', 'Crores', 'Thousands', 'Actual'", default="Lakhs")
    data: list[FinancialData] = Field(description="List of financial data for the last 3 years")

def _ai_fallback(raw_text: str, force_model: str = None) -> List[Dict[str, Any]]:
    """
    Sends the raw text to Gemini and asks for a structured JSON extraction of the financial metrics.
    Includes rate-limit-aware retry with exponential backoff.
    """
    if not client:
        raise ValueError("Gemini Client not initialized (Missing GEMINI_API_KEY)")
        
    prompt = f"""
    You are an expert financial analyst. I am providing you the raw text from the financial statements section of an Indian IPO Red Herring Prospectus (DRHP).
    
    Extract the following financial metrics for the last 3 fiscal years (e.g. 2024, 2023, 2022). 
    If a value is not present, use 0.0. 
    Ensure all values are extracted as simple floating point numbers (e.g. 1500.50). 
    IMPORTANT: Make sure you use the actual financial amounts, NOT the calendar years (do not extract '2024' as the revenue).
    
    CRITICAL - UNIT EXTRACTION:
    DO NOT normalize the units yourself. Extract the exact numbers as they appear in the tables. 
    Instead, you MUST determine the unit used in the financial statements and return it in the `unit` field.
    - Check the header/footnote of the financial table for the unit declaration.
    - It will typically say "in Lakhs", "in ₹ Lakhs", "in Millions", "in Crores", or "in Thousands".
    - If no unit is specified and numbers are huge, it might be "Actual" (single rupees).
    - Return one of: 'Lakhs', 'Millions', 'Crores', 'Thousands', 'Actual'.
    
    Look out for synonyms:
    - pat could be "Profit for the year", "Profit/(Loss) for the period", "Restated Profit After Tax"
    - total_debt could be "Borrowings", "Financial Liabilities", "Long-term Borrowings" + "Short-term Borrowings"
    - equity could be "Net Worth", "Share Capital" + "Reserves", "Total Equity", "Shareholders Funds"
    - total_assets could be "Total Assets", "Total Non-Current Assets" + "Total Current Assets", "Total Assets (A+B)", "Assets"
    - revenue could be "Revenue from Operations", "Total Income", "Net Revenue"
    
    NOTE: The text might be scrambled or space-separated because it's extracted from a PDF. Reconstruct the columns mentally.
    
    The JSON fields must be: year, revenue, pat, ebitda, total_assets, total_debt, equity.
    
    RAW TEXT TO ANALYZE:
    {raw_text[:100000]}
    """
    
    if force_model:
        model_name = force_model
    else:
        model_name = os.environ.get('GEMINI_MODEL', 'gemini-3.1-flash-lite')
    
    # Rate-limit-aware retry with exponential backoff
    max_retries = 4
    base_delay = 45  # seconds — slightly above the 39s the API suggests

    for attempt in range(max_retries):
        try:
            response = client.models.generate_content(
                model=model_name,
                contents=prompt,
                config={
                    'response_mime_type': 'application/json',
                    'response_schema': FinancialResponse,
                    'temperature': 0.0,
                },
            )
            
            try:
                scale_factor = 1.0
                unit_str = response.parsed.unit.lower()
                if "lakh" in unit_str:
                    scale_factor = 100.0
                elif "million" in unit_str:
                    scale_factor = 10.0
                elif "thousand" in unit_str:
                    scale_factor = 10000.0
                elif "actual" in unit_str or "rupee" in unit_str:
                    scale_factor = 10000000.0
                
                result_data = []
                for item in response.parsed.data:
                    d = item.model_dump()
                    if scale_factor != 1.0:
                        d['revenue'] = round(d['revenue'] / scale_factor, 2)
                        d['pat'] = round(d['pat'] / scale_factor, 2)
                        d['ebitda'] = round(d['ebitda'] / scale_factor, 2)
                        d['total_assets'] = round(d['total_assets'] / scale_factor, 2)
                        d['total_debt'] = round(d['total_debt'] / scale_factor, 2)
                        d['equity'] = round(d['equity'] / scale_factor, 2)
                    result_data.append(d)
                return result_data
            except Exception as e:
                logger.error(f"AI extraction failed to parse response: {e}")
                return []
                
        except genai_errors.ClientError as e:
            if "429" in str(e):
                # Extract retry delay from error if available, otherwise use exponential backoff
                delay = base_delay * (2 ** attempt)
                logger.warning(
                    f"Rate limited by Gemini API (attempt {attempt + 1}/{max_retries}). "
                    f"Waiting {delay}s before retry..."
                )
                if attempt < max_retries - 1:
                    time.sleep(delay)
                    continue
                else:
                    logger.error("Max retries exhausted due to rate limiting. Returning empty.")
                    return []
            else:
                raise

    return []

def parse_financials(file_path: str) -> List[Dict[str, Any]]:
    """
    Finds the financial pages, extracts the text, and calls Gemini for structured extraction.
    Returns a list of dicts: [{'year': 2023, 'revenue': 100, ...}]
    """
    keyword_weights = {
        # Table titles (high recall, high weight)
        "statement of profit and loss": 3,
        "statement of assets and liabilities": 3,
        "balance sheet": 3,
        "restated financial statements": 3,
        "restated statement": 3,
        "cash flow statement": 3,
        "annexure i": 3,
        "annexure ii": 3,
        "annexure iii": 3,
        "annexure iv": 3,
        "particulars": 3,
        
        # Row headers (high weight now — balance sheet rows are critical)
        "total assets": 3,
        "total equity and liabilities": 3,
        "non-current assets": 2,
        "current assets": 2,
        "total equity": 2,
        "shareholders funds": 2,
        "total liabilities": 2,
        "total income": 2,
        "revenue from operations": 2,
        "profit for the year": 2,
        "profit for the period": 2,
        "profit/(loss) for the year": 2,
        "profit before tax": 2,
        "cash flows from operating activities": 1,
        "net cash from operating activities": 1
    }
    
    import pdfplumber
    try:
        with pdfplumber.open(file_path) as pdf:
            pages = find_pages_with_keywords(pdf, keyword_weights)
            if not pages:
                return []
            
            # Extract raw text from those specific pages
            raw_text = ""
            for p in pages[:20]:
                raw_text += extract_text_from_pages(pdf, p, 1) + "\n"
    except Exception as e:
        logger.error(f"Failed to open PDF {file_path}: {e}")
        return []
        
    if not raw_text.strip():
        return []
        
    # Add a small inter-request delay to avoid blasting the free tier quota
    time.sleep(2)
    

    results = _ai_fallback(raw_text, force_model="gemini-3.1-flash-lite")
    
    needs_fallback = False
    if not results:
        needs_fallback = True
    else:
        for row in results:
            if row.get('revenue', 0) == 0.0 or row.get('pat', 0) == 0.0 or row.get('ebitda', 0) == 0.0 or \
               row.get('total_assets', 0) == 0.0 or row.get('equity', 0) == 0.0:
                needs_fallback = True
                break
                
    if needs_fallback:
        logger.warning("Missing critical metrics detected (0.0). Escalating to gemini-3.5-flash...")
        time.sleep(2)
        results = _ai_fallback(raw_text, force_model="gemini-3.5-flash")
        
    return results
