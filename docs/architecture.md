# Architecture

## High-Level Flow

1. User uploads benchmark file in UI (`frontend`).
2. Go API stores file under `/data/uploads` and inserts metadata.
3. API enqueues parsing message to Redis `parse_jobs`.
4. Python parser worker normalizes benchmark controls/safeguards and writes canonical records to PostgreSQL.
5. User requests compare between versions.
6. API inserts `diff_reports` row and pushes job to Redis `diff_jobs`.
7. Python diff worker compares safeguards and writes `diff_items`.
8. UI loads report and shows side-by-side diff.
9. Optional reporting module exports JSON/Excel/HTML.

## Services

- `api` (Go, Gin, GORM, Redis client)
- `parser` (Python, pandas/openpyxl/pdfplumber/pydantic)
- `diff-engine` (Python, deepdiff/rapidfuzz/difflib)
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
