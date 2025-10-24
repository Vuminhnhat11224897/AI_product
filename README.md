<!--
  Production README for AI Production Pipeline
  - Vietnamese-focused, concise, and action-oriented
-->

# AI Production Pipeline

Production-ready batch processor that analyzes kids' financial activity and generates AI reports (GPT-4o).
Designed for reliability, observability, and cost tracking.

Key features
- Automatic week detection from the database
- Silver layer: transform & enrich data with trends
- Gold layer: AI report generation with token tracking and cost estimation
- Rate limiting, retry with exponential backoff, and concurrent batching
- Docker & docker-compose ready

## Quick start (Docker) — recommended
1. Copy example env and edit:

```powershell
cp .env.example .env
# Edit .env to add OPENAI_API_KEY and DB settings
```

2. Start full stack (Postgres + pipeline + adminer):

```powershell
docker-compose up -d --build
docker-compose logs -f pipeline
```

3. Stop and cleanup:

```powershell
docker-compose down
```

## Quick start (local)
1. Create `.env` with `OPENAI_API_KEY` and DB credentials (see `.env.example`).
2. Install deps and build:

```powershell
go mod tidy
go build -o pipeline.exe main.go
```

3. Run:

```powershell
# Full run (all available weeks)
.\pipeline.exe

# Test mode: only process the last week (saves tokens)
$env:TEST_LAST_WEEK_ONLY = "true"; .\pipeline.exe
```

## 📁 Project Structure

```
ai-production-pipeline/
├── main.go                      # Entry point & orchestrator
├── Dockerfile                   # Multi-stage Go build
├── docker-compose.yml           # PostgreSQL + Pipeline + Adminer
├── .env.example                 # Environment template
├── .dockerignore                # Docker build exclusions
│
├── config/
│   └── config.yaml              # All configuration (NO hardcode)
│
├── internal/
│   ├── config/                  # Config loader
│   ├── constants/               # System constants
│   ├── errors/                  # Error types
│   ├── logger/                  # Logger setup
│   ├── weekmanager/             # Week detection & management
│   ├── silver/                  # Data transformation layer
│   │   ├── types.go
│   │   └── silver_layer.go
│   ├── gold/                    # AI report generation layer
│   │   └── gold_layer.go
│   └── processor/               # AI processing engine
│       ├── processor.go         # Core processor with batch/retry
│       └── token_tracker.go     # Token usage & cost tracking
│
├── prompts/
│   ├── vietnamese_financial_report.txt
│   └── system_message.txt
│
├── scripts/
│   ├── generate_10kids_continuous_activity.py
│   └── run_test_quick.bat
│
├── data/                        # Output directory
│   ├── kids_analysis_week_*.json    (Silver layer outputs)
│   └── kids_reports_week_*.json     (Gold layer outputs)
│
└── logs/                        # Runtime logs
    └── pipeline_*.log
```

## Configuration highlights
- All runtime settings live in `config/config.yaml` (batch sizes, concurrency, rate limits, retry).
- Secrets (OpenAI key) must be set via `.env` or environment variables. Do NOT commit `.env`.

Example important env vars (in `.env`):
- `OPENAI_API_KEY` — your API key
- `DATABASE_URL` or DB-specific vars used by `config/config.yaml`

## Test — last week only (recommended during development)
- Use environment variable `TEST_LAST_WEEK_ONLY=true` to limit processing to the latest week and save tokens.

PowerShell example:

```powershell
$env:TEST_LAST_WEEK_ONLY = "true"
.\pipeline.exe
$env:TEST_LAST_WEEK_ONLY = ""  # clear after
```

There is also a helper script: `scripts\run_test_quick.bat` and `scripts\test_last_week.ps1` to automate the build+run.

## Token tracking & cost estimation
- Token usage is tracked per-request and aggregated per-week.
- Pricing used (configurable): GPT-4o input $2.50 / 1M tokens, output $10.00 / 1M tokens.
- Logs include a per-week breakdown and a total estimated cost.

## Troubleshooting (common)
- If DB connection fails: ensure Postgres is running (`docker-compose ps`) and `.env` DB values are correct.
- If OpenAI errors occur: check `OPENAI_API_KEY` and rate limits; reduce `batch.max_concurrent` in `config/config.yaml`.
- If build fails: run `go mod tidy` and `go mod download`.

## Production checklist
- [x] Config in YAML (no hardcode)
- [x] Token usage tracking & cost report
- [x] Graceful shutdown with context cancellation
- [x] Retry + rate-limiting
- [x] Docker-ready deployment

## Support
Open an issue in the repository with logs (`logs/pipeline_*.log`) and a short description.

---

**Production-ready** 🚀 — Minimal config, maximum observability.

