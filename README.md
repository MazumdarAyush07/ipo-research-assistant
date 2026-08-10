# IPO Research Assistant

An AI-powered system designed to analyze IPOs by extracting financials from DRHPs, tracking subscriptions and GMP, benchmarking peers, and evaluating risk through Claude AI. 

## Local Setup

### Prerequisites
- Docker and Docker Compose
- Go 1.22
- Node.js 18+ (for frontend development)
- Python 3.11 (for pdf parser development)

### Getting Started

1. **Clone the repository:**
   ```bash
   git clone https://github.com/MazumdarAyush07/ipo-research.git
   cd ipo-research
   ```

2. **Environment Configuration:**
   Copy `.env.example` to `.env` and fill in your details.
   - `CLAUDE_API_KEY` (Used for Risk Evaluation)
   - `GEMINI_API_KEY` (Used for DRHP Financial Extraction)
   ```bash
   cp .env.example .env
   ```

3. **Start the services via Docker Compose:**
   ```bash
   docker-compose up --build
   ```

### Architecture
- **Backend:** Go / Fiber API Server (Port: `8080`)
- **PDF Parser Sidecar:** Python / FastAPI (Port: `8000`)
- **Frontend Dashboard:** Next.js (Port: `3000`)
- **Database:** PostgreSQL (Port: `5432`)
- **Cache / Message Queue:** Redis (Port: `6379`)
