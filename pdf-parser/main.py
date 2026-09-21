from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="IPO PDF Parser")

class ParseRequest(BaseModel):
    s3_key: str
    ipo_id: int
    doc_type: str

from typing import List, Optional, Dict, Any
from extractors.financials import parse_financials
from extractors.text_extractor import extract_risk_factors, extract_objects_of_issue, extract_promoter_background
import pdfplumber
import boto3
import os
import tempfile

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

def download_from_r2(s3_key: str) -> str:
    """Downloads a file from Cloudflare R2 to a temporary file and returns the path."""
    s3 = boto3.client('s3',
        endpoint_url=f"https://{os.environ['R2_ACCOUNT_ID']}.r2.cloudflarestorage.com",
        aws_access_key_id=os.environ['R2_ACCESS_KEY_ID'],
        aws_secret_access_key=os.environ['R2_SECRET_ACCESS_KEY'],
        region_name='auto'
    )
    
    fd, temp_path = tempfile.mkstemp(suffix=".pdf")
    os.close(fd) # Close the file descriptor, boto3 will open it
    
    s3.download_file(os.environ['R2_BUCKET_NAME'], s3_key, temp_path)
    return temp_path

@app.post("/parse", response_model=ParseResponse)
def parse_pdf(req: ParseRequest):
    temp_path = None
    try:
        temp_path = download_from_r2(req.s3_key)
        
        financials_data = parse_financials(temp_path)
        objects_text = extract_objects_of_issue(temp_path)
        risks_text = extract_risk_factors(temp_path)
        promoters_text = extract_promoter_background(temp_path)
        
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
    except Exception as e:
        return ParseResponse(
            status="error",
            ipo_id=req.ipo_id,
            financials=[],
            objects_of_issue="",
            risk_factors="",
            promoters="",
            confidence_scores={},
            message=str(e)
        )
    finally:
        if temp_path and os.path.exists(temp_path):
            os.remove(temp_path)


class ValidateRequest(BaseModel):
    s3_key: str

class ValidateResponse(BaseModel):
    valid: bool
    page_count: int
    error: Optional[str] = None

@app.post("/validate", response_model=ValidateResponse)
def validate_pdf(req: ValidateRequest):
    temp_path = None
    try:
        temp_path = download_from_r2(req.s3_key)
        with pdfplumber.open(temp_path) as pdf:
            page_count = len(pdf.pages)
            if page_count < 50:
                return ValidateResponse(valid=False, page_count=page_count, error="Page count is less than 50")
            return ValidateResponse(valid=True, page_count=page_count)
    except Exception as e:
        return ValidateResponse(valid=False, page_count=0, error=str(e))
    finally:
        if temp_path and os.path.exists(temp_path):
            os.remove(temp_path)

class ExtractTextRequest(BaseModel):
    s3_key: str

class ExtractTextResponse(BaseModel):
    status: str
    text: str
    error: Optional[str] = None

@app.post("/extract-text", response_model=ExtractTextResponse)
def extract_full_text(req: ExtractTextRequest):
    temp_path = None
    try:
        temp_path = download_from_r2(req.s3_key)
        full_text = []
        with pdfplumber.open(temp_path) as pdf:
            for page in pdf.pages:
                text = page.extract_text()
                if text:
                    full_text.append(text)
        return ExtractTextResponse(status="success", text="\n".join(full_text))
    except Exception as e:
        return ExtractTextResponse(status="error", text="", error=str(e))
    finally:
        if temp_path and os.path.exists(temp_path):
            os.remove(temp_path)
