# CIS Benchmark Intelligence

CIS Benchmark Intelligence ingests CIS benchmark source files (`.xlsx`, `.csv`, `.pdf`), normalizes controls/safeguards, computes version differences, and exposes APIs/UI for comparison and reporting.

## Stack

- Docker + Docker Compose
- Go (`api`, `collector`)
- Python (`parser`, `diff-engine`)
- React (`frontend`)
- PostgreSQL (`db/schema.sql`)
- Redis (parse and diff job queues)

## Quick Start

```bash
docker compose up --build
```

Services:

- API: `http://localhost:8080`
- Frontend: `http://localhost:3000`
- Postgres: `localhost:5432`
- Redis: `localhost:6379`

## Core API Endpoints

- `POST /api/upload` (multipart, accepts `xlsx|csv|pdf`)
- `GET /frameworks`
- `GET /frameworks/{id}/versions`
- `POST /compare`
- `GET /diff/{report_id}`
- `GET /health`

## Job Queues

- `parse_jobs`: upload ingestion and normalization jobs
- `diff_jobs`: version comparison jobs

## Export Formats

Diff reports can be exported to:

- JSON
- Excel (`added_controls`, `removed_controls`, `modified_controls`, `renamed_controls`)
- HTML

Use:

```bash
python diff-engine/reporting.py <report_id> --output-dir /data/exports
```

## Optional Extensions

- OpenSearch indexing: `diff-engine/opensearch_indexer.py`
- AI summaries: `diff-engine/ai_summary.py` (`langchain` + `llama-index` optional deps)

## Documentation

- [Architecture](docs/architecture.md)
- [API Usage](docs/api-usage.md)
- [Parser Format](docs/parser-format.md)
- [Diff Logic](docs/diff-logic.md)
- [Deployment](docs/deployment.md)
