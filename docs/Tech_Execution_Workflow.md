# IPO Research Assistant — Technology Execution Workflow

> **Stack:** Go + Fiber · PostgreSQL · Next.js 14 · Redis · Claude API

---

# Philosophy

Don't build an IPO Predictor. Build an **IPO Research Assistant**.

Predicting returns is nearly impossible. Reducing the time required to perform good research from 3 hours to 10 minutes is a far more achievable and valuable goal.

The AI supports your decision — it does not replace it.

---

# Rules of Execution

- Build in this sequence: **Foundation → Data Layer → Core Engines → AI → UI → Deployment**
- Every phase must have a clear "Done" outcome before moving to the next.
- Never build UI before the API for that module exists.
- Every financial calculation must have unit tests.
- Database migrations are immutable — never edit, always add.
- Never delete scraped data. Archive (soft delete) instead.
- Every AI prompt lives in `/prompts/` — versioned and testable.
- All secrets in environment variables — never hardcoded.
- Optimise for minimal manual entry: auto-detect, auto-score.
- Every IPO analyzed immediately becomes training data for backtesting.

---

# What's in V1 vs V2

| Feature | V1 | V2 |
|---|---|---|
| IPO calendar + DRHP download | ✅ | |
| PDF parsing + financial extraction | ✅ | |
| Financial metrics calculator | ✅ | |
| Peer comparison | ✅ | |
| Subscription + GMP tracker | ✅ | |
| AI document analyzer (analyst mode) | ✅ | |
| Explainable scoring engine | ✅ | |
| Historical IPO database | ✅ | |
| HTML report generation | ✅ | |
| Next.js dashboard | ✅ | |
| Backtesting engine | ✅ | |
| Authentication | ❌ | ✅ |
| Cloud deployment | ❌ | ✅ |
| PDF report export | ❌ | ✅ |
| Email / push notifications | ❌ | ✅ |
| Multi-user support | ❌ | ✅ |
| Portfolio + XIRR tracking | ❌ | ✅ |
| AI Debate (Bull vs Bear analyst) | ❌ | ✅ |

> **V1 runs locally. That's intentional.** Deploy only after you've analyzed ~50 IPOs and trust the system.

---

# Tech Stack

| Layer | Technology | Reason |
|---|---|---|
| API Server | Go + Fiber | Fast, typed, minimal overhead — suits your domain expertise |
| Database | PostgreSQL 16 | Relational integrity for financial time-series data |
| ORM / Queries | sqlc + pgx/v5 | Type-safe SQL generation — no magic, no reflection |
| Migrations | goose | Same tooling as your Payment Ledger System |
| Cache | Redis | Avoid re-downloading the same DRHP or re-calling AI |
| Task Queue | Asynq (Redis-backed) | Async PDF parsing and report generation jobs |
| Frontend | Next.js 14 (App Router) | React Server Components + Tailwind for the dashboard |
| PDF Parsing | Python sidecar | pdfplumber / camelot for table extraction (Go bindings are weak here) |
| AI | Claude API (Sonnet) | DRHP analysis, risk extraction, explainable recommendations |
| Storage | Local FS | DRHPs and generated reports — S3 in V2 |
| Deployment | Docker Compose (local) | One-command dev; cloud lift deferred to V2 |

---

# High-Level Architecture

```text
IPO Calendar Monitor (cron, daily)
        │
        ▼  New IPO detected
  ┌─────┴──────────────┬──────────────────┐
  ▼                    ▼                   ▼
Download DRHP    Fetch Market Data    Fetch Peer Data
  └─────────────────────┬───────────────────┘
                        ▼
              Data Processing Layer
                        ▼
           Financial Metrics Calculator
                        ▼
            AI Analyst (not just summarizer)
                        ▼
            Explainable Scoring Engine
                        ▼
         PostgreSQL + Redis (cache layer)
                        ▼
              Historical IPO Database
                        ▼
         Next.js Dashboard + HTML Report
```

---

# Three Milestones

The phases below map to three outcome-driven milestones. Use these to track meaningful progress.

**Milestone 1 — Data Engine** *(Phases 1–5)*
> "I can collect and parse IPO data automatically."

**Milestone 2 — Intelligence Engine** *(Phases 6–10)*
> "I can evaluate an IPO in under 10 minutes with explainable scores."

**Milestone 3 — Decision Engine** *(Phases 11–13)*
> "I know which IPO strategy actually works for me."

---

# Project Structure

```text
ipo-research/
│
├── backend/                    (Go)
│     ├── cmd/server/           main.go
│     ├── internal/
│     │     ├── api/            HTTP handlers (Fiber)
│     │     ├── services/       Business logic
│     │     ├── workers/        Asynq job handlers
│     │     ├── scoring/        Explainable scoring engine
│     │     ├── models/         DB structs (sqlc generated)
│     │     └── ai/             Claude API client
│     ├── migrations/           goose .sql files
│     └── db/query/             .sql files for sqlc
│
├── pdf-parser/                 (Python sidecar service)
│     ├── main.py
│     └── extractors/
│
├── frontend/                   (Next.js 14)
│     ├── app/
│     ├── components/
│     └── lib/
│
├── prompts/                    versioned AI prompt templates
├── storage/                    downloaded DRHPs and RHPs
├── reports/                    generated HTML reports
└── docker/
      └── docker-compose.yml
```

---

# — MILESTONE 1: DATA ENGINE —

---

# Phase 1 — Project Foundation

**Goal:** Reproducible dev environment — one command, everything running.

## Tasks
- [x] Initialise Go module (`go mod init github.com/MazumdarAyush07/ipo-research`)
- [x] Scaffold folder structure as defined above
- [x] Initialise Next.js 14 with App Router, Tailwind CSS, TypeScript
- [x] Setup Docker Compose (Go server, Python sidecar, PostgreSQL, Redis)
- [x] Configure goose for migrations and create the baseline empty migration
- [x] Setup sqlc config (`sqlc.yaml`) pointing to `/db/query/`
- [x] Environment variable management (`.env.example` + godotenv in Go)
- [x] ESLint + Prettier for frontend
- [x] golangci-lint config for backend
- [x] GitHub Actions — basic CI: lint + `go test` + `next build`
- [x] README with local setup instructions

## Deliverables
- `docker-compose up` brings up all services and they communicate
- `GET /health` returns 200 from the Go server
- CI pipeline is green on the main branch

## Done when
`docker-compose up` brings up the Go server, Python sidecar, Next.js app, PostgreSQL, and Redis — all running and talking to each other locally.

---

# Phase 2 — Database Schema

**Goal:** Freeze the data model — no fundamental schema changes after this.

## Tables to Create
- [ ] `ipos` — name, sector, price_band_low, price_band_high, open_date, close_date, listing_date, status, source_url
- [ ] `financials` — fk:ipo, year, revenue, pat, ebitda, total_assets, total_debt, equity
- [ ] `valuation` — fk:ipo, issue_price, market_cap, pe_ratio, pb_ratio, ev_ebitda
- [ ] `peer_companies` — fk:ipo, name, ticker, pe, pb, ev_ebitda, roe, market_cap
- [ ] `subscription_data` — fk:ipo, category (QIB/NII/Retail/Employee), times_subscribed, recorded_at
- [ ] `gmp_history` — fk:ipo, gmp_amount, premium_percent, recorded_at
- [ ] `documents` — fk:ipo, type (DRHP/RHP/PROSPECTUS), file_path, downloaded_at, parsed_at
- [ ] `ai_analysis` — fk:ipo, business_model, moat, promoter_risk, legal_cases, customer_concentration, debt_assessment, key_risks, red_flags, management_assumptions, industry_outlook, raw_json
- [ ] `scores` — fk:ipo, financials_score, financials_reason, valuation_score, valuation_reason, promoter_score, promoter_reason, industry_score, industry_reason, risk_score, risk_reason, subscription_score, gmp_score, final_score, recommendation, scored_at
- [ ] `ipo_outcomes` — fk:ipo, issue_price, listing_price, listing_return_pct, one_month_return_pct, six_month_return_pct, one_year_return_pct, our_score_at_listing
- [ ] `reports` — fk:ipo, file_path, format (HTML/MD), generated_at

> Note the `scores` table now carries a `*_reason` field for every component — this is how explainability is persisted. Note also `ipo_outcomes` — every IPO you analyze feeds the backtesting engine from day one.

## Deliverables
- goose migration scripts from empty DB → full schema
- Reverse migrations verified
- Seed script: 2 sample IPOs with financials
- `docs/schema.md` with ER diagram description

## Done when
Migrations run clean from empty to full schema. Seed data loads without errors. Every entity from the architecture maps to at least one table.

---

# Phase 3 — IPO Calendar Monitor

**Goal:** Auto-detect new IPOs daily — no manual entry required.

## Tasks
- [ ] Build NSE IPO calendar scraper (Go HTTP client + goquery)
- [ ] Build BSE IPO calendar scraper
- [ ] Deduplication logic — match by name + open_date before inserting
- [ ] Asynq cron job — runs daily at 7:00 AM IST
- [ ] On new IPO detected → enqueue document download job
- [ ] Store: name, sector, open_date, close_date, listing_date, price_band, status
- [ ] `GET /api/ipos` — paginated list with filters (status, sector, date range)
- [ ] `POST /api/ipos` — manual override for IPOs not yet on the calendar

## Deliverables
- Cron job runs daily and inserts new IPOs
- Deduplication prevents double entries
- API returns structured IPO list

## Done when
Run the cron manually and it correctly identifies and inserts at least one upcoming IPO from NSE or BSE without duplicates.

---

# Phase 4 — Document Downloader

**Goal:** Automatically fetch DRHP, RHP, and Prospectus for every new IPO.

## Tasks
- [ ] Go service: download PDFs from NSE/BSE/SEBI links
- [ ] Save to `/storage/{company-slug}/{drhp|rhp|prospectus}.pdf`
- [ ] Redis cache: skip re-download if file already exists
- [ ] Update `documents` table: file_path, downloaded_at, status
- [ ] Asynq worker: triggered automatically when new IPO is detected
- [ ] Retry logic with exponential backoff (3 retries)
- [ ] File size validation — alert if PDF < 100KB (likely a bad download)
- [ ] `POST /api/ipos/:id/documents/trigger` — manual re-download trigger

## Deliverables
- `/storage/` directory populated with real DRHPs
- `documents` table has correct file paths and timestamps

## Done when
Worker is triggered for a new IPO and downloads the DRHP to the correct path without re-downloading on a second trigger.

---

# Phase 5 — PDF Parser (Python Sidecar)

**Goal:** Extract structured financial data from DRHP PDFs.

## Tasks
- [ ] Python FastAPI sidecar service with `POST /parse` endpoint
- [ ] Accept: `{ file_path, ipo_id, doc_type }`
- [ ] Use pdfplumber for raw text extraction
- [ ] Use camelot for financial table extraction
- [ ] Extract: Revenue, PAT, EBITDA, Total Assets, Total Debt, Equity (last 3 fiscal years)
- [ ] Extract: Objects of Issue (use of funds)
- [ ] Extract: Risk Factors section (top 20 risks as text)
- [ ] Extract: Promoter background paragraph
- [ ] Return structured JSON; Go backend saves to `financials` table
- [ ] Fallback: if table extraction fails, return raw text for AI to process
- [ ] Log extraction confidence score per field

## Deliverables
- Python sidecar runs as a Docker service
- `POST /parse` returns structured JSON for a real DRHP
- `financials` table populated after parsing

## Done when
Feed a real DRHP PDF and receive correctly structured revenue, PAT, and debt figures for at least 2 fiscal years.

---

# — MILESTONE 2: INTELLIGENCE ENGINE —

---

# Phase 6 — Financial Metrics Calculator

**Goal:** Compute all key ratios automatically from extracted data.

## Tasks
- [ ] Go service: `CalculateMetrics(ipo_id)` — reads from `financials` table
- [ ] Revenue CAGR (3Y, 5Y)
- [ ] PAT CAGR (3Y, 5Y)
- [ ] EBITDA Margin (each year)
- [ ] PAT Margin (each year)
- [ ] Return on Equity (ROE)
- [ ] Return on Capital Employed (ROCE)
- [ ] Debt-to-Equity Ratio
- [ ] Asset Turnover Ratio
- [ ] PE Ratio (issue price / annualised EPS)
- [ ] EV/EBITDA (market cap + debt − cash / EBITDA)
- [ ] Unit tests for every formula (table-driven tests in Go)
- [ ] `GET /api/ipos/:id/financials` — returns computed metrics

## Deliverables
- All metrics calculated and returned via API
- >90% test coverage on calculator package

## Done when
Unit tests pass for all financial formulas. API returns correct metrics for seed data, verified against manual calculation.

---

# Phase 7 — Peer Comparison

**Goal:** Benchmark IPO valuation against listed sector peers.

## Tasks
- [ ] Sector → peers mapping (JSON config file to start)
- [ ] Fetch peer market multiples (NSE/BSE API or scraping): PE, PB, EV/EBITDA, ROE
- [ ] Redis cache: peer data with 24-hour TTL
- [ ] Comparison engine: IPO valuation vs peer median
- [ ] Premium/discount calculation (% above or below peer median PE)
- [ ] Flag: >40% premium → Expensive | 10–40% → Fair | <10% → Attractive
- [ ] Store in `peer_companies` table
- [ ] `GET /api/ipos/:id/peers` — returns peer table + comparison metrics

## Deliverables
- Peer data fetched and cached for at least 5 sectors
- API returns peer comparison table with premium/discount flag

## Done when
For a hospital-sector IPO, the API returns Apollo, Max, Fortis, Narayana with current PE ratios and a correct premium/discount calculation.

---

# Phase 8 — Subscription & GMP Tracker

**Goal:** Track live demand signals during the IPO window.

## Subscription Tracker
- [ ] Scraper: fetch NSE/Chittorgarh subscription data every 2 hours during IPO open days
- [ ] Store: category (QIB, NII, Retail, Employee, Shareholder), times_subscribed, recorded_at
- [ ] Cron: active only between open_date and close_date + 1 day
- [ ] `GET /api/ipos/:id/subscription` — returns full time-series array

## GMP Tracker
- [ ] Scraper: fetch GMP from IPO Watch / InvestorGain every 6 hours
- [ ] Store: gmp_amount, premium_percent, recorded_at — never overwrite, always append
- [ ] Trend calculation: RISING, FALLING, or STABLE in last 24h
- [ ] Volatility: stddev of last 5 readings
- [ ] `GET /api/ipos/:id/gmp` — returns time-series + trend + volatility

## Deliverables
- Subscription data updates automatically during IPO window
- GMP stored as a time-series, not just today's value

## Done when
During an active IPO, the subscription endpoint returns at least 3 data points across different categories, and the GMP endpoint returns a trend value of RISING, FALLING, or STABLE.

---

# Phase 9 — AI Document Analyzer

**Goal:** Make Claude think like an analyst, not a summarizer.

The AI layer should ask adversarial questions — the kind a fund manager would ask before committing capital. Anyone can summarize a DRHP. The value is in surfacing what management doesn't want you to notice.

## Tasks
- [ ] Prompt library in `/prompts/` — one file per prompt type, versioned
- [ ] Go AI client: POST to Claude API with chunked DRHP text
- [ ] Chunking strategy: split by sections (Business, Risk Factors, Financials, Promoters)
- [ ] Redis cache: AI response per (ipo_id, prompt_version) — avoid re-calling for same DRHP
- [ ] Store result in `ai_analysis` table (including red_flags and management_assumptions)
- [ ] Fallback: if JSON parse fails, store raw text and flag for manual review
- [ ] `GET /api/ipos/:id/ai-analysis` — returns structured AI output

## Prompt Design

The prompt should instruct Claude to act as an analyst, not a summarizer:

```
You are a senior equity research analyst reviewing this DRHP before an IPO.

Answer each question based only on what is in the document.
If something is not mentioned, say "Not disclosed."

Questions:
1. What is the core business model in 2 sentences?
2. What is the real moat, if any? Be skeptical of generic claims.
3. What are the 3 biggest red flags in this document?
4. What assumptions is management making that could turn out to be wrong?
5. Which risks in the Risk Factors section are boilerplate, and which are unique to this company?
6. Who are the top 3 customers, and what % of revenue do they represent?
7. Why could this IPO fail to deliver returns?
8. Rate customer concentration 1–10 (10 = one customer = 100% revenue).
9. Rate promoter credibility: "Strong" | "Adequate" | "Weak" | "Concerns"
10. Rate litigation risk: "None" | "Minor" | "Significant"

Return ONLY a valid JSON object. No preamble. No explanation.
```

## Deliverables
- AI analysis runs end-to-end for a real DRHP
- Result includes red_flags and management_assumptions fields
- Cached — second call returns within 100ms

## Done when
Feed a real DRHP and receive a valid JSON with all fields populated, including at least 2 specific red flags. Second call returns the cached result.

---

# Phase 10 — Explainable Scoring Engine

**Goal:** Every score must explain itself. A number you can't understand is a number you won't trust.

## Scoring Breakdown

| Module | Weight | Key Signals |
|---|---|---|
| Financials | 40 pts | Revenue CAGR, PAT CAGR, Margins, ROE, Debt ratio |
| Valuation | 20 pts | PE vs peers, EV/EBITDA premium/discount |
| Promoter & Governance | 10 pts | AI promoter score, legal cases flag |
| Industry | 10 pts | AI industry outlook, sector tailwinds |
| Risk Assessment | 10 pts | AI risk rating, red flags count, customer concentration |
| Subscription | 5 pts | QIB / NII / Retail subscription levels |
| GMP Signal | 5 pts | GMP trend and momentum |

## Recommendation Thresholds

```
≥ 75  →  Apply
50–74  →  Apply with caution
< 50  →  Avoid
```

## Explainability Format

Every component score must carry a human-readable reason. Example:

```
Financial Score:  28 / 40
Reason: Revenue CAGR 32% (strong), PAT CAGR 27% (good),
        ROE 18% (adequate), Debt/Equity 0.3 (low risk).
        Deducted 12 pts: margins declining YoY.

Valuation Score:  12 / 20
Reason: PE of 45x vs peer median 31x — 45% premium.
        Flagged: Expensive.
```

## Tasks
- [ ] Go scoring package: `ScoreIPO(ipo_id)` — reads from all tables
- [ ] Scoring function for each module with configurable weights
- [ ] Each module returns `{ score, reason }` — reason is a plain English string
- [ ] Recommendation threshold logic
- [ ] Store all component scores + reasons in `scores` table
- [ ] Unit tests: edge cases (missing data, partial data, all-max, all-min)
- [ ] `GET /api/ipos/:id/score` — returns full score breakdown with reasons
- [ ] Score history: re-score after new data arrives (subscription updates, GMP changes)

## Done when
`ScoreIPO()` returns a score AND a reason string for every component. Unit tests cover all edge cases. The reason strings read like a human wrote them.

---

# — MILESTONE 3: DECISION ENGINE —

---

# Phase 11 — HTML Report Generator

**Goal:** Auto-produce a shareable research report for every IPO. HTML only in V1 — fast to generate, readable in any browser.

## Tasks
- [ ] Go template engine: render HTML report from score + financials + AI analysis
- [ ] Report sections: Summary, Explainable Score Breakdown, Financial Highlights, Peer Comparison, AI Analyst Findings (red flags, management assumptions, boilerplate vs unique risks), Recommendation
- [ ] Store generated reports in `/reports/{company-slug}/report.html`
- [ ] Store path + generated_at in `reports` table
- [ ] `POST /api/ipos/:id/report/generate` — trigger generation (async Asynq job)
- [ ] `GET /api/ipos/:id/report` — return file path / serve HTML
- [ ] Regenerate automatically when score is updated

## Sample Report Structure

```
IPO Research Report: {Company Name}
─────────────────────────────────────
Final Score:       82 / 100
Recommendation:    ✅ APPLY

Score Breakdown (with reasons)
  Financials    28 / 40  — Revenue CAGR 32%, margins declining YoY
  Valuation     12 / 20  — 45% premium to peer median PE
  Promoter       8 / 10  — Strong track record, no litigation
  Industry       9 / 10  — Tailwind: sector growing 18% CAGR
  Risk           5 / 10  — 3 red flags identified by AI
  Subscription   8 /  5  — QIB 42x, Retail 8x
  GMP            3 /  5  — Rising trend, +₹28 premium

Red Flags (AI identified)
  ✗ Top 2 customers = 68% of revenue
  ✗ Promoter pledged 22% of shares
  ✗ Revenue concentrated in one geography

AI Analyst Summary: [Claude-generated paragraph]
```

## Done when
`POST /generate` produces a readable HTML report with explainable scores and AI red flags for a real IPO.

---

# Phase 12 — Next.js Dashboard

**Goal:** Visual interface — build this only after all APIs exist.

## Pages & Components
- [ ] `/dashboard` — IPO cards grid (score, recommendation, sector, dates)
- [ ] `/ipo/:id` — Full IPO detail page
  - [ ] Explainable score breakdown (ring chart per component + reason text)
  - [ ] Financial highlights table (3Y trend)
  - [ ] Peer comparison table with premium/discount flag
  - [ ] Subscription live tracker (updates every 2h during window)
  - [ ] GMP trend chart (Recharts line graph)
  - [ ] AI Analyst findings — red flags, management assumptions, unique risks (structured cards)
  - [ ] View HTML Report button
- [ ] `/history` — all past IPOs with outcome tracking (listing return, 1M, 6M, 1Y)
- [ ] Design system: Tailwind + shadcn/ui
- [ ] Server components for initial data fetch; client components for live data
- [ ] Loading skeletons for all async data
- [ ] Mobile responsive (you'll check scores on your phone)

## Done when
All API endpoints are consumed. A full IPO's data is visible end-to-end in the browser — explainable score, financials, peers, GMP, AI red flags, and HTML report link.

---

# Phase 13 — Backtesting Engine

**Goal:** Validate your scoring model against historical IPOs. Without this, your score is an opinion. With this, it's evidence.

## Tasks
- [ ] Seed `ipo_outcomes` table with historical data (CSV import — last 100 mainboard IPOs to start)
- [ ] For each historical IPO: issue_price, listing_price, 1M return, 6M return, 1Y return
- [ ] Re-run `ScoreIPO()` using only data available before listing date (no subscription, no GMP — those come later)
- [ ] Correlation analysis: final score vs listing-day return
- [ ] Correlation analysis: each component score vs outcome
- [ ] Answer the questions that matter:
  - Does high QIB subscription reliably predict listing gains?
  - Does GMP add value beyond fundamentals?
  - Are valuation penalties too harsh or too lenient?
  - Which factors predict long-term returns vs day-1 listing pop?
- [ ] `/backtest` dashboard page — scatter plot of score vs return
- [ ] Weight sandbox: adjust scoring weights and see backtest results update live

## Done when
Backtest runs on 50+ historical IPOs and produces a correlation table between each component score and listing-day return. You can answer at least one of the four questions above with data.

---

# Phase 14 — Complete API Layer

**Goal:** Wire everything into a clean, documented REST API before V2 work begins.

## Endpoints

```
GET    /api/ipos                     — paginated list (filter by status, sector, date)
GET    /api/ipos/:id                 — full IPO detail
POST   /api/ipos                     — manual IPO entry
GET    /api/ipos/:id/financials       — computed metrics
GET    /api/ipos/:id/peers            — peer comparison table
GET    /api/ipos/:id/subscription     — time-series subscription data
GET    /api/ipos/:id/gmp              — GMP history + trend
GET    /api/ipos/:id/ai-analysis      — Claude analyst output
GET    /api/ipos/:id/score            — explainable score breakdown
POST   /api/ipos/:id/report/generate  — trigger HTML report generation
GET    /api/ipos/:id/report           — serve HTML report
GET    /api/ipos/:id/documents        — list of downloaded docs
GET    /api/backtest                  — backtest results summary
GET    /api/health                    — service health
```

## Tasks
- [ ] Consistent error response format: `{ error, code, message }`
- [ ] Cursor-based pagination on all list endpoints
- [ ] Request logging middleware (structured JSON logs)
- [ ] API documentation (OpenAPI auto-generated from Go structs)
- [ ] Rate limiting on AI endpoints (prevent runaway Claude API costs)
- [ ] Integration tests for every endpoint (Go `httptest`)

## Done when
All endpoints return correct responses. Integration tests pass. OpenAPI docs accessible at `/api/docs`.

---

# Master Checklist (V1)

- [ ] Project Foundation — `docker-compose up` works
- [ ] Database Schema — migrations run clean, outcomes table included
- [ ] IPO Calendar Monitor — auto-detects new IPOs
- [ ] Document Downloader — DRHPs saved to storage
- [ ] PDF Parser — structured financials extracted
- [ ] Financial Metrics Calculator — all ratios computed and tested
- [ ] Peer Comparison — valuation benchmarked against sector
- [ ] Subscription & GMP Tracker — live demand signals captured
- [ ] AI Document Analyzer — Claude returns red flags, not just summaries
- [ ] Explainable Scoring Engine — every score has a reason string
- [ ] HTML Report Generator — report auto-produced with AI findings
- [ ] Next.js Dashboard — all data visible end-to-end
- [ ] Backtesting Engine — scoring model validated against 50+ IPOs
- [ ] Complete API Layer — all endpoints documented and tested

---

# Next Immediate Actions

1. Scaffold folder structure and initialise Go module
2. Get `docker-compose up` running with Go + PostgreSQL + Redis
3. Write and run database migrations — include `ipo_outcomes` from day one
4. Build the NSE IPO calendar scraper — first real feature

---

# V2 Roadmap (After 50 IPOs Analyzed)

- Authentication (JWT + Google OAuth) — when you want to access it from anywhere
- Cloud deployment (Fly.io / Railway) — after the system is proven locally
- PDF report export — once HTML is battle-tested
- Portfolio tracking — applied, allotted, sold, XIRR per IPO
- Strategy comparison — "Apply >80 score" vs "Apply if GMP >20%" vs "Hold 1 year"
- AI Debate — Bull Analyst vs Bear Analyst generating a balanced investment memo

---

# V3 Roadmap (After the System Is Proven)

> Don't touch V3 until V2 is shipped and the backtesting engine has validated your scoring model. Both ideas below add real architectural complexity — they're worth it only once the core is solid.

## Knowledge Graph

Replace the flat sector → peers mapping with a proper knowledge graph (Neo4j or similar). Instead of a static list of competitors, the system understands relationships:

```
Hyundai IPO
    │
    ▼
Automobile Sector
    ├── Competitors: Maruti, Tata Motors, Mahindra
    ├── Dependencies: Steel Prices, Chip Supply
    ├── Tailwinds: EV Policy, PLI Scheme
    └── Risks: Fuel Price Sensitivity, Import Duties
```

This lets the AI answer questions like:
- "How sensitive is this IPO to steel prices?"
- "How does this compare structurally to Tata Motors at the time of its listing?"
- "Which macro factors have historically hurt this sector?"

That's a fundamentally different class of insight than anything a flat database can produce.

## Richer Data Sources

The DRHP is 400–700 pages and enough for V1 and V2. In V3, layer in additional sources to catch material changes that happen *after* the DRHP is filed:

- Annual reports and quarterly results of peer companies
- Earnings call transcripts (management tone analysis)
- Credit rating reports (CRISIL / ICRA / CARE)
- Industry reports referenced inside the DRHP itself
- News published between DRHP filing date and listing date

The last one matters most. A regulatory change, a major customer loss, or a sector-wide event in that window can completely change the risk profile — and the DRHP won't mention it.