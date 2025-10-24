# 🚀 AI Production Pipeline# AI Production Pipeline



**Hệ thống xử lý dữ liệu tài chính trẻ em tự động với AI****Production-grade AI batch processor for kids' financial behavior analysis** - Stable, configurable, and feature-complete with comprehensive logging and error handling.



## 📋 Tính năng## 📋 Tổng Quan



- ✅ **Tự động phát hiện tuần**: Tự động detect các tuần từ databasePipeline xử lý AI với tất cả tính năng production:

- ✅ **Phân tích xu hướng**: So sánh 2 tuần trước để phát hiện tăng/giảm- ✅ **Batch Processing**: Chia nhỏ xử lý theo lô có thể cấu hình

- ✅ **AI Reports**: Tạo báo cáo tự động bằng GPT-4o- ✅ **Parallel Execution**: Chạy song song có kiểm soát (semaphore pattern)

- ✅ **Tracking chi phí**: Theo dõi token usage và chi phí API- ✅ **Rate Limiting**: Token bucket algorithm, tuân thủ giới hạn API

- ✅ **Production-ready**: Batch processing, retry, rate limiting- ✅ **Auto Retry**: Exponential backoff với retry configurable

- ✅ **Table Formatting**: Bảng kết quả đẹp với box drawing characters

## 🏗️ Kiến trúc- ✅ **Comprehensive Logging**: Log đầy đủ từ code và API responses

- ✅ **Error Handling**: Production-ready error handling và recovery

```- ✅ **Zero Hardcode**: Tất cả config trong file YAML

Database (PostgreSQL)

    ↓## 🏗️ Kiến Trúc

Silver Layer (Transform + Enrich + Trends)

    ↓ kids_analysis_week_N.json```

Gold Layer (AI Processing)ai-production-pipeline/

    ↓ kids_reports_week_N.json├── main.go                          # Entry point - orchestrates everything

Reports├── config/

```│   └── config.yaml                  # All configuration (NO hardcode)

├── internal/

## 🚀 Quick Start│   └── processor/

│       ├── processor.go             # AI processor core (batch, parallel, retry)

### Option 1: Docker (Recommended)│       └── formatter.go             # Table formatter for results

├── data/

```bash│   ├── kids_analysis.json           # Input data

# 1. Copy environment file│   └── kids_report_*.json           # Output reports (timestamped)

cp .env.example .env├── logs/

│   └── ai_processor.log             # Application logs

# 2. Edit .env and add your OpenAI API key├── .env                             # API keys (git-ignored)

nano .env  # or use any editor└── README.md                        # This file

```

# 3. Start all services

docker-compose up -d## 🚀 Quick Start (3 bước)



# 4. Check logs### Bước 1: Setup Environment

docker-compose logs -f pipeline

```powershell

# 5. Stop servicescd d:\AI\ai-production-pipeline

docker-compose down

```# Tạo .env file với API key

"OPENAI_API_KEY=sk-your-actual-key-here" | Out-File -Encoding UTF8 .env

### Option 2: Local Development

# Install dependencies

```bashgo mod tidy

# 1. Install dependencies```

go mod download

### Bước 2: Cấu Hình (Optional)

# 2. Set environment variables

export OPENAI_API_KEY="sk-your-key"File `config/config.yaml` có tất cả settings:



# 3. Run pipeline```yaml

go run main.go# Chỉnh batch size theo nhu cầu

batch:

# Or build and run  size: 5                    # Items per batch

go build -o pipeline.exe  max_concurrent: 3          # Max parallel calls

./pipeline.exe

```# Chỉnh rate limit theo API tier

rate_limit:

## ⚙️ Configuration  requests_per_minute: 50    # Adjust to your limit



Edit `config/config.yaml`:# Retry configuration

retry:

```yaml  max_attempts: 3

database:  exponential_backoff: true

  host: "localhost"      # "postgres" for Docker```

  port: 5432

  user: "postgres"### Bước 3: Chạy

  password: "zseefvhu12"

  dbname: "testAI"```powershell

# Cách 1: Run trực tiếp

openai:go run main.go

  model: "gpt-4o"

  max_tokens: 4000# Cách 2: Build rồi chạy

  temperature: 1.0go build -o pipeline.exe

.\pipeline.exe

batch:```

  size: 10

  max_concurrent: 10## 📊 Output



rate_limit:### Console Output

  requests_per_minute: 500

``````

🚀 AI PRODUCTION PIPELINE STARTING

## 📊 Outputs✅ AI Processor initialized

✅ Rate limiter initialized

**Silver Layer** (`data/kids_analysis_week_N.json`):📖 Loading input data from: data/kids_analysis.json

- Dữ liệu đã transform với trends và statistics✅ Loaded 10 kids from input file

- 100 kids mỗi tuần🚀 Starting AI batch processing...

- Historical comparison (2 weeks)

📦 Processing batch 1/2

**Gold Layer** (`data/kids_reports_week_N.json`):✅ Item processed successfully (index: 0, tokens: 850)

- AI reports từ GPT-4o✅ Item processed successfully (index: 1, tokens: 920)

- Phân tích xu hướng, đề xuất cải thiện✅ Batch completed

- Format JSON chuẩn

📊 Processing Statistics

## 📈 Token Usage Report  total_items: 10 | successful: 10 | failed: 0 | success_rate: 100.00%



Tự động tracking và báo cáo:⚡ Performance Metrics

```  total_duration: 45.2s | average_per_item: 4.5s | total_retries: 0

Input tokens:  104,265 ($0.26)

Output tokens:  57,101 ($0.57)🎯 Token Usage

Total cost:     $0.83 USD  total_tokens: 8,750 | avg_tokens_per_item: 875

```

╔══════════════════════════════════════════════════════════════════════════════╗

## 🧪 Generate Test Data║                        FINAL PROCESSING SUMMARY                              ║

╠══════════════════════════════════════════════════════════════════════════════╣

```bash║  Total: 10   |  Success: 10   |  Failed: 0   |  Success Rate: 100.00%       ║

cd scripts╚══════════════════════════════════════════════════════════════════════════════╝

python generate_10kids_continuous_activity.py

```💾 Reports saved to: data/kids_report_20251016_165430.json

🎉 AI PRODUCTION PIPELINE COMPLETED SUCCESSFULLY

Tạo 10 kids với 6 tuần hoạt động liên tục.```



## 🐳 Docker Commands### File Output



```bash**`data/kids_report_20251016_165430.json`**:

# Build only```json

docker-compose build{

  "generated_at": "2025-10-16T16:54:30Z",

# Start services  "total_reports": 10,

docker-compose up -d  "reports": [

    {

# View logs      "nickname": "Alice",

docker-compose logs -f pipeline      "financial_health": "Excellent financial habits...",

      "spending_pattern": "Balanced spending across categories...",

# Restart pipeline      "savings_behavior": "Strong savings discipline...",

docker-compose restart pipeline      "recommendations": "1. Continue current...",

      "activity_level": "Highly engaged...",

# Stop all      "overall_score": 8.5,

docker-compose down      "generated_at": "2025-10-16T16:54:30Z"

    }

# Stop and remove volumes  ]

docker-compose down -v}

```

# Access database

docker exec -it ai-production-db psql -U postgres -d testAI## ⚙️ Configuration Chi Tiết



# Database admin UI### OpenAI Settings

# Open http://localhost:8080 (Adminer)

``````yaml

openai:

## 📁 Project Structure  model: "gpt-4o-mini"         # Model name

  max_tokens: 2000             # Response limit

```  temperature: 0.7             # Creativity (0.0-1.0)

ai-production-pipeline/  timeout_seconds: 60          # Request timeout

├── main.go                    # Entry point```

├── Dockerfile                 # Docker build config

├── docker-compose.yml         # Multi-container setup### Batch & Concurrency

├── .dockerignore             # Docker ignore rules

├── .env.example              # Environment template```yaml

│batch:

├── config/  size: 5                      # Items per batch (smaller = more frequent status updates)

│   └── config.yaml           # Main configuration  max_concurrent: 3            # Max parallel API calls (depends on rate limit)

│```

├── internal/

│   ├── config/               # Config loader**Khuyến nghị:**

│   ├── constants/            # Shared constants- `size: 5-10` cho datasets nhỏ (<50 items)

│   ├── weekmanager/          # Week detection- `max_concurrent: 3-5` để không vượt rate limit

│   ├── silver/               # Data transformation- Tăng `max_concurrent` nếu bạn có tier cao hơn

│   │   ├── types.go

│   │   └── silver_layer.go### Rate Limiting

│   ├── gold/                 # AI processing

│   │   └── gold_layer.go```yaml

│   └── processor/            # AI enginerate_limit:

│       ├── processor.go  requests_per_minute: 50      # Must match your API tier

│       └── token_tracker.go```

│

├── prompts/**OpenAI Tiers:**

│   ├── vietnamese_financial_report.txt- Free: 3 req/min

│   └── system_message.txt- Tier 1: 60 req/min

│- Tier 2: 3,500 req/min

├── scripts/

│   └── generate_10kids_continuous_activity.py### Retry Strategy

│

├── data/                     # Outputs (mounted in Docker)```yaml

│   ├── kids_analysis_week_*.jsonretry:

│   └── kids_reports_week_*.json  max_attempts: 3                    # Total attempts (1 original + 3 retries)

│  initial_delay_seconds: 2           # First retry delay

└── logs/                     # Application logs  max_delay_seconds: 10              # Max delay cap

    └── pipeline_*.log  exponential_backoff: true          # 2s → 4s → 8s (capped at 10s)

``````



## 🔧 Troubleshooting### Logging



### Database connection error```yaml

```bashlogging:

# Check PostgreSQL is running  level: "info"                      # debug, info, warn, error

docker-compose ps  output: "both"                     # console, file, both

  log_dir: "logs"

# Check logs  log_file: "ai_processor.log"

docker-compose logs postgres  json_format: false                 # Set true for structured logs

  include_caller: true               # Show file:line in logs

# Test connection```

docker exec -it ai-production-db psql -U postgres -d testAI -c "SELECT 1"

```## 🔧 Xử Lý Lỗi



### OpenAI API errors### Rate Limit Errors

- Verify API key in `.env`

- Check rate limits (500 RPM for Tier 1)Pipeline tự động:

- Monitor token usage in logs1. **Token bucket** kiểm soát tốc độ requests

2. **Exponential backoff** khi gặp 429 errors

### Out of memory3. **Retry** tự động với delay tăng dần

- Reduce `batch.max_concurrent` in config.yaml

- Increase Docker memory limit### Network Errors



## 💰 Cost Estimation```

⚠️ Request failed, retrying...

**GPT-4o Pricing:**  attempt: 1/3 | error: connection timeout | retry_in: 2s

- Input: $2.50 / 1M tokens⚠️ Request failed, retrying...

- Output: $10.00 / 1M tokens  attempt: 2/3 | error: connection timeout | retry_in: 4s

✅ Item processed successfully (index: 5, retries: 2)

**Typical run (100 kids × 7 weeks = 700 AI calls):**```

- Input: ~150K tokens → **$0.37**

- Output: ~80K tokens → **$0.80**### API Errors

- **Total: ~$1.17 per full run**

Các lỗi API được log đầy đủ:

## 📝 License```

❌ Item processing failed after all retries

MIT  index: 7 | retries: 3 | error: API error: insufficient_quota

```

## 🤝 Support

### Context Cancellation

For issues or questions, create an issue on GitHub.

Nhấn `Ctrl+C` để graceful shutdown:
```
🛑 Received interrupt signal, shutting down gracefully...
⚠️ Gold layer processing was cancelled
```

## 📈 Performance Monitoring

### Token Tracking

```yaml
monitoring:
  track_token_usage: true      # Log token consumption
```

Output:
```
🎯 Token Usage
  total_tokens: 8,750
  avg_tokens_per_item: 875
```

**Cost Estimation** (GPT-4o-mini):
- Input: $0.15 / 1M tokens
- Output: $0.60 / 1M tokens
- 10 kids × 875 tokens ≈ $0.005

### Timing Metrics

```yaml
monitoring:
  track_timing: true           # Log processing times
```

Output:
```
⚡ Performance Metrics
  total_duration: 45.2s
  average_per_item: 4.5s
```

### Progress Updates

```yaml
monitoring:
  show_progress: true          # Real-time progress
```

Output:
```
📊 Progress update: 3/10 (30.0%)
📊 Progress update: 6/10 (60.0%)
📊 Progress update: 10/10 (100.0%)
```

## 🛠️ Troubleshooting

### Issue: "OPENAI_API_KEY not found"

```powershell
# Check if .env exists
Get-Content .env

# If not, create it:
"OPENAI_API_KEY=sk-your-key" | Out-File -Encoding UTF8 .env
```

### Issue: Rate limit errors

Giảm `max_concurrent` trong config:
```yaml
batch:
  max_concurrent: 2      # Reduce from 3 to 2
```

### Issue: Timeout errors

Tăng timeout:
```yaml
openai:
  timeout_seconds: 120   # Increase from 60 to 120
```

### Issue: Missing packages

```powershell
go mod tidy
go mod download
```

## 🔍 Architecture Details

### 1. Main Flow (`main.go`)

```go
Load Config → Setup Logger → Load Data → Process Batch → Save Results
```

- Graceful shutdown với context cancellation
- Comprehensive error handling
- Config-driven, no hardcode

### 2. AI Processor (`internal/processor/processor.go`)

**Core Features:**

```go
// Batch processing with controlled concurrency
func ProcessBatch(ctx, items, promptTemplate) []ProcessResult

// Single item with retry logic
func processItemWithRetry(ctx, index, item, promptTemplate) ProcessResult

// Rate-limited API call
func callOpenAI(ctx, prompt) (output, usage, error)
```

**Token Bucket Rate Limiter:**
```go
type RateLimiter struct {
    tokens     chan struct{}      // Token bucket
    refillRate time.Duration      // Refill interval
}
```

### 3. Table Formatter (`internal/processor/formatter.go`)

```go
// Display results as table
func FormatResultsTable(results []ProcessResult)

// Calculate and display summary
func calculateSummary(results) ResultSummary
```

## 📝 Best Practices

### 1. Config Management

- ✅ **DO**: Đặt tất cả settings trong `config.yaml`
- ✅ **DO**: API keys trong `.env` (git-ignored)
- ❌ **DON'T**: Hardcode values trong code

### 2. Error Handling

- ✅ **DO**: Log đầy đủ với context
- ✅ **DO**: Retry với exponential backoff
- ❌ **DON'T**: Silent failures

### 3. Performance

- ✅ **DO**: Monitor token usage
- ✅ **DO**: Adjust batch size theo dataset
- ✅ **DO**: Respect API rate limits
- ❌ **DON'T**: Set `max_concurrent` quá cao

### 4. Production Deployment

- ✅ **DO**: Set `logging.output: "file"` hoặc `"both"`
- ✅ **DO**: Enable `json_format` cho log aggregation
- ✅ **DO**: Monitor logs directory size
- ❌ **DON'T**: Commit `.env` file

## 🔐 Security

### Environment Variables

```powershell
# .env file structure
OPENAI_API_KEY=sk-proj-...
```

### .gitignore

```
.env
*.log
logs/
data/kids_report_*.json
*.exe
```

## 📚 Dependencies

```go
require (
    github.com/joho/godotenv v1.5.1      // .env loading
    github.com/sirupsen/logrus v1.9.3    // Structured logging
    gopkg.in/yaml.v3 v3.0.1              // YAML parsing
)
```

## 🎯 Production Checklist

- [x] ✅ Không có hardcoded values
- [x] ✅ Tất cả config trong YAML
- [x] ✅ API keys trong environment variables
- [x] ✅ Comprehensive logging
- [x] ✅ Error handling và retry logic
- [x] ✅ Rate limiting
- [x] ✅ Graceful shutdown
- [x] ✅ Progress monitoring
- [x] ✅ Token usage tracking
- [x] ✅ Performance metrics
- [x] ✅ Formatted output (table + JSON)
- [x] ✅ Context cancellation support

## 📞 Support

Các lỗi thường gặp đã được xử lý tự động:
- ✅ Rate limit → Automatic throttling
- ✅ Network errors → Retry with backoff
- ✅ API errors → Logged với full context
- ✅ Timeout → Configurable timeouts
- ✅ Invalid data → Graceful skip với warning

---

**Ready for Production** 🚀 - Code ổn định, đầy đủ tính năng, dễ maintain!
