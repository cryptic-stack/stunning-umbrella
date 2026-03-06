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

UI workflow tabs:

- `Benchmark Workflow` (upload + tag + compare)
- `Reports` (Diff Viewer + GPO Reports)
- `GPO Workflow` (policy import + mapping import + assessment)

Services:

- API: `http://localhost:8080`
- Frontend (HTTP): `http://localhost` (port 80)
- Frontend (alt port): `http://localhost:443`

Runtime data (uploads, exports, database state) is stored in Docker named volumes, not in repository directories.
Postgres and Redis are internal-only service endpoints on the Docker network and are not published to the host.

## Authentication and RBAC

The API enforces OIDC bearer-token authentication and role-based authorization for protected endpoints.

Required API environment variables for internet-exposed deployments:

- `OIDC_ISSUER_URL` (OIDC issuer URL)
- `OIDC_CLIENT_ID` (expected token audience/client)
- `CORS_ALLOWED_ORIGINS` (comma-separated explicit origins, e.g. `https://app.example.com`)

Optional:

- `AUTH_ENABLED` (`true` by default; set `false` only for local/dev)
- `AUTH_BOOTSTRAP_ADMIN_EMAIL` (auto-provisions/updates this email as an active Admin user at startup)
- `UPLOAD_MAX_BYTES` (multipart upload limit in bytes, default `20971520`)

Role model:

- `Viewer`: read-only report/framework/upload access
- `Reviewer`: Viewer + create comparisons and review diff items
- `Admin`: full access, including uploads, report deletion, and settings/user/role management

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

## GPO Assessment Endpoints

- `POST /api/gpo/import` (queue policy source import with automatic source-type discovery; optional `source_type` override supports `gpresult_xml`, `gpmc_xml`, `secedit_inf`, `registry_pol`)
- `POST /api/gpo/mappings/import` (queue curated benchmark mapping import from CSV/JSON)
- `POST /api/gpo/assess` (queue assessment run)
- `GET /api/gpo/assessments`
- `GET /api/gpo/assessments/{assessment_id}`
- `GET /api/gpo/assessments/{assessment_id}/report/{json|md|html|csv|xlsx|docx}`

Milestone 1 workflow:

1. Upload CIS benchmark files as before.
2. Collect `gpresult` XML and import with `POST /api/gpo/import`.
3. Import curated mapping CSV/JSON via `POST /api/gpo/mappings/import`.
4. Run assessment via `POST /api/gpo/assess`.
5. Export remediation-ready reports (DOCX/XLSX/CSV/HTML/Markdown/JSON).

Example mapping file: `data/mappings/cis_windows_example_mapping.csv`
PowerShell collection scripts: `scripts/windows/collect-gpresult.ps1`, `scripts/windows/collect-secedit.ps1`

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
- [GPO Assessment](docs/gpo-assessment.md)
