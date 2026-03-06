# Architecture

## High-Level Flow

1. User uploads benchmark file in UI (`frontend`).
2. Go API stores file under `/data/uploads` and inserts metadata.
3. API enqueues parsing message to Redis `parse_jobs` (references `upload_id`; parser resolves file path from database).
4. Python parser worker normalizes benchmark controls/safeguards and writes canonical records to PostgreSQL.
5. User requests compare between versions.
6. API inserts `diff_reports` row and pushes job to Redis `diff_jobs`.
7. Python diff worker compares safeguards and writes `diff_items`.
8. UI loads report and shows side-by-side diff.
9. Optional reporting module exports JSON/Excel/HTML.
10. Admin imports Windows policy exports (gpresult/GPMC/secedit/registry.pol) through API.
11. API enqueues GPO jobs to Redis `gpo_jobs`.
12. Python `gpo-assessment` worker normalizes policy settings, applies benchmark mappings, evaluates compliance, and writes assessment results.
13. Worker exports remediation reports in Markdown/HTML/DOCX/CSV/XLSX/JSON.

## Services

- `api` (Go, Gin, GORM, Redis client)
- `parser` (Python, pandas/openpyxl/pdfplumber/pydantic)
- `diff-engine` (Python, deepdiff/rapidfuzz/difflib)
- `gpo-assessment` (Python, XML/INF/policy parsing + mapping + comparator + report exporter)
- `postgres` (persistent storage)
- `redis` (queues)
- `frontend` (React + MUI + Monaco + react-diff-view)
- `collector` (Go + colly + cron, optional profile)

## Data Model

Primary entities:

- `frameworks`
- `versions`
- `controls`
- `safeguards`
- `diff_reports`
- `diff_items`
- `uploaded_files`
- `policy_sources`
- `policy_settings`
- `benchmark_policy_rules`
- `assessment_runs`
- `assessment_results`
