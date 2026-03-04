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

- `GET /` (API service info + frontend URL)
- `POST /api/upload` (multipart, accepts `xlsx|csv|pdf`)
- `GET /uploads`
- `PUT /uploads/{id}/tag`
- `DELETE /uploads/{id}?purge=true`
- `GET /frameworks`
- `GET /frameworks/{id}/versions`
- `POST /compare`
- `POST /compare` accepts optional `control_level` (`ALL`, `L1`, `L2`)
- `GET /diff/{report_id}`
- `GET /reports`
- `GET /reports/{report_id}/download/{json|xlsx|html}`
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
