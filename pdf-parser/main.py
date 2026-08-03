from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="IPO PDF Parser")

class ParseRequest(BaseModel):
    file_path: str
    ipo_id: int
    doc_type: str

@app.get("/docs")
def docs():
    return {"message": "Swagger UI is at /docs by default in FastAPI"}

@app.get("/health")
def health():
    return {"status": "ok", "service": "pdf-parser"}

@app.post("/parse")
def parse_pdf(req: ParseRequest):
    # TODO: Implement camelot and pdfplumber extraction
    return {
        "status": "success",
        "ipo_id": req.ipo_id,
        "message": "Parsing endpoint placeholder"
    }
