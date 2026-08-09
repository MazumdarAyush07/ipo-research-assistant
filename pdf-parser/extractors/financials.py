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
    print(f"Failed to initialize Gemini Client: {e}")
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
    data: list[FinancialData] = Field(description="List of financial data for the last 3 years")

def _ai_fallback(raw_text: str) -> List[Dict[str, Any]]:
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
    Look out for synonyms:
    - total_debt could be "Borrowings", "Financial Liabilities"
    - equity could be "Net Worth", "Share Capital" + "Reserves", "Total Equity"
    - total_assets could be "Total Assets", "Net Block"
    
    The JSON fields must be: year, revenue, pat, ebitda, total_assets, total_debt, equity.
    
    RAW TEXT TO ANALYZE:
    {raw_text[:8000]}  # Limit token size for cost efficiency
    """
    
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
                },
            )
            
            try:
                return [item.model_dump() for item in response.parsed.data]
            except Exception as e:
                logger.error(f"AI extraction failed to parse response: {e}")
                return []
                
        except genai_errors.ClientError as e:
            if e.status_code == 429:
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
    keywords = [
        "restated consolidated statement of profit and loss",
        "restated consolidated statement of assets and liabilities",
        "restated statement of profit and loss",
        "restated statement of assets and liabilities",
        "restated cash flow statement"
    ]
    
    pages = find_pages_with_keywords(file_path, keywords)
    if not pages:
        return []
    
    logger.info(f"Found financial keywords on pages: {pages[:5]}")
    
    # Extract raw text from those specific pages
    raw_text = ""
    for p in pages[:5]:
        raw_text += extract_text_from_pages(file_path, p, 1) + "\n"
        
    if not raw_text.strip():
        return []
        
    # Add a small inter-request delay to avoid blasting the free tier quota
    time.sleep(2)
    
    logger.info("Sending text to Gemini AI for extraction...")
    return _ai_fallback(raw_text)
