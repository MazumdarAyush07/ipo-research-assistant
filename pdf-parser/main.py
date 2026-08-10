from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="IPO PDF Parser")

class ParseRequest(BaseModel):
    file_path: str
    ipo_id: int
    doc_type: str

from typing import List, Optional, Dict, Any
from extractors.financials import parse_financials
from extractors.text_extractor import extract_risk_factors, extract_objects_of_issue, extract_promoter_background

class FinancialYear(BaseModel):
    year: int
    revenue: float
    pat: float
    ebitda: float
    total_assets: float
    total_debt: float
    equity: float

class ParseResponse(BaseModel):
    status: str
    ipo_id: int
    financials: List[FinancialYear]
    objects_of_issue: str
    risk_factors: str
    promoters: str
    confidence_scores: Dict[str, float]
    message: Optional[str] = None

@app.get("/docs")
def docs():
    return {"message": "Swagger UI is at /docs by default in FastAPI"}

@app.get("/health")
def health():
    return {"status": "ok", "service": "pdf-parser"}

@app.post("/parse", response_model=ParseResponse)
def parse_pdf(req: ParseRequest):
    print(f"Parsing PDF for IPO ID: {req.ipo_id} at {req.file_path}")
    
    financials_data = parse_financials(req.file_path)
    objects_text = extract_objects_of_issue(req.file_path)
    risks_text = extract_risk_factors(req.file_path)
    promoters_text = extract_promoter_background(req.file_path)
    
    financials = [FinancialYear(**year_data) for year_data in financials_data]
    
    return ParseResponse(
        status="success",
        ipo_id=req.ipo_id,
        financials=financials,
        objects_of_issue=objects_text,
        risk_factors=risks_text,
        promoters=promoters_text,
        confidence_scores={
            "financials": 0.85 if len(financials) > 0 else 0.0,
            "text_extraction": 0.90
        }
    )

import pdfplumber

class ValidateRequest(BaseModel):
    file_path: str

class ValidateResponse(BaseModel):
    valid: bool
    page_count: int
    error: Optional[str] = None

@app.post("/validate", response_model=ValidateResponse)
def validate_pdf(req: ValidateRequest):
    try:
        with pdfplumber.open(req.file_path) as pdf:
            page_count = len(pdf.pages)
            if page_count < 50:
                return ValidateResponse(valid=False, page_count=page_count, error="Page count is less than 50")
            return ValidateResponse(valid=True, page_count=page_count)
    except Exception as e:
        return ValidateResponse(valid=False, page_count=0, error=str(e))
