FROM python:3.11-slim

WORKDIR /app

# Install system dependencies required by camelot and pdfplumber
RUN apt-get update && apt-get install -y \
    ghostscript \
    python3-tk \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
