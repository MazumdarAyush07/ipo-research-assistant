# Execution Pipeline (Production)

This document outlines the daily operating procedure for running the IPO Research Assistant in a production environment. 

Due to the unreliability of automated PDF links from third-party sites (e.g., Chittorgarh) and the high token cost of AI extraction, this pipeline utilizes a **Human-In-The-Loop (HITL)** architecture. We strictly audit downloaded files before feeding them to the AI parser to prevent garbage-in-garbage-out scenarios.

## Prerequisites
Ensure the Docker containers are running:
```bash
docker-compose up -d
```

---

## Step 1: Automated Ingestion & Download
Trigger the system to detect new IPOs and attempt automated downloads of their Red Herring Prospectus (DRHP/RHP) documents.

```bash
python3 scripts/batch_trigger.py
```
**What happens under the hood:** 
- Queries the backend for all active IPOs.
- Instructs the `worker` service to scrape Chittorgarh for the prospectus link.
- Downloads the PDF, checks it for minimum size (>500KB) and basic integrity, and stores it in `/storage/<ipo-slug>/drhp.pdf`.
- If the link is a "Notice", "Checklist", or missing, the worker will safely abort the download.

---

## Step 2: The Morning Audit (The Checkpoint)
Verify exactly what was successfully downloaded and what failed or was corrupted.

```bash
python3 scripts/audit_downloads.py
```
**What happens under the hood:**
- Scans the `/storage/` directory.
- Checks if the PDF exists for each IPO.
- Validates that the PDF can be opened and has **>50 pages** (guaranteeing it is a full prospectus, not a 2-page notice).
- Highlights any IPOs with missing or corrupted PDFs.

---

## Step 3: Manual Intervention (HITL)
For any IPO flagged as "Missing DRHP" or "Below 50 Pages" in Step 2:
1. Manually search SEBI's portal (sebi.gov.in) or BSE/NSE for the official **Red Herring Prospectus**.
2. Download the valid PDF.
3. Drop the file directly into the respective folder: `/storage/<ipo-slug>/drhp.pdf`.
4. (Optional) Re-run `audit_downloads.py` to ensure you've achieved 100% coverage.

---

## Step 4: Batch Parsing (AI Extraction)
Once the `/storage/` directory contains pristine, complete PDFs for all IPOs, trigger the AI analysis.

```bash
python3 scripts/batch_parse.py
```
**What happens under the hood:**
- Instructs the Go backend to queue parsing jobs for every valid document in storage.
- The `pdf-parser` Python sidecar ingests the documents, extracts raw text using `pdfplumber`, and feeds up to 60,000 characters to Gemini AI to extract structured financial data, risk factors, and promoter backgrounds.
- Saves the extracted structured JSON to the `financials` PostgreSQL table.

---

## Step 5: Final Audit & Metrics
Verify that the AI successfully extracted the financials without hallucination or truncation.

```bash
python3 scripts/audit_financials.py
```
**What happens under the hood:**
- Generates a CSV report (`reports/audit_financials.csv`) containing the extraction results.
- Highlights "0.0" values which indicate the AI failed to locate the balance sheet tables.
- Once verified, the data is ready for the Phase 7 Metrics Calculator to compute CAGR, EBITDA Margins, and ROE.
