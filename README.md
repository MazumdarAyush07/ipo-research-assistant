# IPO Research Assistant

An AI-powered institutional-grade system designed to automate IPO research. The assistant scrapes DRHPs, uses Gemini 3.5 Flash Lite to extract deep financial and risk data, tracks live subscriptions and GMP, benchmarks peers via Yahoo Finance, and evaluates the final thesis through an explainable scoring engine.

All automation pipelines and manual overrides are managed through a comprehensive Next.js **Admin Dashboard**.

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
Clone the repository and copy the environment variables template:

```bash
git clone https://github.com/MazumdarAyush07/ipo-research-assistant.git
cd ipo-research-assistant
cp .env.example .env
```

Populate the `.env` file with your credentials:
- `DATABASE_URL` (Your Neon connection string)
- `GEMINI_API_KEY` (Used for DRHP Financial Extraction and AI Analyst evaluation)

### 3. Run the Stack
Spin up the entire application locally using Docker Compose:

```bash
docker-compose up --build
```

- **Frontend UI / Admin Dashboard:** `http://localhost:3000`
- **Go Backend API:** `http://localhost:8080`
- **Python PDF Parser:** `http://localhost:8000`

---

## Operating the System (Admin Dashboard)

Instead of running terminal scripts, the entire IPO execution pipeline is managed via the **Admin Dashboard** at `http://localhost:3000/admin`. 

From the Admin UI, you can trigger:
1. **Automated Ingestion:** Scrape new IPO listings.
2. **Parsing & Analytics Audit:** Trigger background jobs to download DRHPs and extract financial/risk tables using Gemini.
3. **Trackers & Sync:** Sync live Peer Valuation data (Yahoo Finance), Subscriptions, and Grey Market Premium (GMP).
4. **Scoring Engine:** Recalculate component scores (Financials, Valuation, Promoters, Industry, Risk) and output a definitive recommendation (Apply, Caution, Avoid).
5. **Report Generation:** Compile the final data into a shareable HTML report saved in `/storage`.

---

## Production Deployment

The project is configured for a robust CI/CD cloud deployment setup:

- **Frontend (Edge):** Deployed on Vercel. Connects to the backend via the `NEXT_PUBLIC_API_URL` environment variable.
- **Backend (Core):** Deployed on an Oracle Cloud "Always Free" ARM VM (4 OCPUs, 24GB RAM).
- **Docker Production:** A dedicated `docker/docker-compose.prod.yml` isolates the backend services (Go, Python sidecar, Redis) and provisions persistent volume mounts for the `storage/` directory.
- **CI/CD:** Pushes to the `main` branch trigger a GitHub Action (`.github/workflows/deploy-backend.yml`) that securely SSHes into the Oracle VM to rebuild and restart the Docker containers automatically.
