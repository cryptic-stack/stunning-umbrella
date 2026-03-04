# API Usage

## Authentication

- Disabled by default.
- Enable OIDC verification with:
  - `AUTH_ENABLED=true`
  - `OIDC_ISSUER_URL=<issuer-url>`
  - `OIDC_CLIENT_ID=<audience-client-id>`

## Endpoints

### `POST /api/upload`

Multipart form fields:

- `file` (required): `.xlsx`, `.csv`, `.pdf`
- `framework` (optional): framework name (example `CIS Controls`)
- `version` (optional): version label (example `8`)
- `release_date` (optional): `YYYY-MM-DD`

Example:

```bash
curl -X POST http://localhost:8080/api/upload \
  -F "framework=CIS Controls" \
  -F "version=8" \
  -F "file=@./cis_v8.xlsx"
```

### `GET /frameworks`

Returns all frameworks.

### `GET /frameworks/{id}/versions`

Returns all versions for a framework.

### `POST /compare`

Request body:

```json
{
  "framework": "CIS Controls",
  "version_a": "7",
  "version_b": "8"
}
```

or with framework id:

```json
{
  "framework_id": 1,
  "version_a": "7",
  "version_b": "8"
}
```

Response:

```json
{
  "report_id": 12,
  "status": "queued"
}
```

### `GET /diff/{report_id}`

Returns diff report metadata and diff items.
