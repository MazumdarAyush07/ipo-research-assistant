# IPO Research Assistant

🔗 **Live Demo:** [ipo-research-assistant.vercel.app](https://ipo-research-assistant.vercel.app)

An AI-powered institutional-grade system designed to automate IPO research. 
The assistant scrapes DRHPs, uses Gemini Flash to extract deep financial and 
risk data, tracks live subscriptions and GMP, benchmarks peers via Yahoo 
Finance, and evaluates the final thesis through an explainable scoring engine.

All automation pipelines and manual overrides are managed through a 
comprehensive Next.js **Admin Dashboard**.

## Tech Stack

- **Backend:** Go 1.22 + Fiber + Asynq (Task Queue)
- **Database:** PostgreSQL (Neon Serverless)
- **Cache/Queue Broker:** Redis
- **Frontend Dashboard:** Next.js 14 (App Router) + Tailwind CSS
- **AI Extraction Sidecar:** Python 3.11 + FastAPI + pdfplumber + Gemini API

---

## Local Setup

### 1. Prerequisites
- Docker and Docker Compose
- A free [Neon PostgreSQL](https://neon.tech/) database URL
- A [Google Gemini API Key](https://aistudio.google.com/app/apikey)

### 2. Getting Started

```bash
git clone https://github.com/MazumdarAyush07/ipo-research-assistant.git
cd ipo-research-assistant
cp .env.example .env
```

Populate the `.env` file with your credentials:
- `DATABASE_URL` (Your Neon connection string)
- `GEMINI_API_KEY` (Used for DRHP extraction and AI scoring)

### 3. Run the Stack

```bash
docker-compose up --build
```

- **Frontend UI / Admin Dashboard:** `http://localhost:3000`
- **Go Backend API:** `http://localhost:8080`
- **Python PDF Parser:** `http://localhost:8000`

---

## Operating the System (Admin Dashboard)

The entire IPO execution pipeline is managed via the **Admin Dashboard** 
at `http://localhost:3000/admin`.

From the Admin UI, you can trigger:
1. **Automated Ingestion:** Scrape new IPO listings.
2. **Parsing & Analytics Audit:** Trigger background jobs to download DRHPs 
   and extract financial/risk tables using Gemini.
3. **Trackers & Sync:** Sync live Peer Valuation data (Yahoo Finance), 
   Subscriptions, and Grey Market Premium (GMP).
4. **Scoring Engine:** Recalculate component scores (Financials, Valuation, 
   Promoters, Industry, Risk) and output a definitive recommendation 
   (Apply, Caution, Avoid).
5. **Report Generation:** Compile the final data into a shareable HTML 
   report saved in `/storage`.

---

## Production Deployment & Architecture

### Cost-Aware Infrastructure Design

PDF processing is intentionally kept off the hosted infrastructure. DRHPs 
are large (often hundreds of pages), processing is infrequent, and running 
a memory-intensive Python/pdfplumber sidecar in the cloud 24/7 for a batch 
job that runs a few times a week is wasteful.

Local processing writing directly to the cloud database is the right 
tradeoff for this workload. Because the database is Neon (serverless), both 
the local sidecar and the production backend naturally hit the same data 
layer with zero synchronization issues.

**What is deployed to production:**
- Go backend API + Asynq (Task Queue)
- Next.js Admin Dashboard — deployed on Vercel
- PostgreSQL (Neon Serverless)
- Redis

**What runs locally (on demand):**
- Python PDF Parser sidecar (pdfplumber + Gemini)
- Connects securely to the exact same production `DATABASE_URL`

### CI/CD
Pushes to `main` trigger automated pipelines that rebuild and redeploy 
the Go backend and Redis containers.