# API Usage

## Authentication

- Enabled by default.
- Disable only for local development with:
  - `AUTH_ENABLED=false`
- OIDC verification requires:
  - `OIDC_ISSUER_URL=<issuer-url>`
  - `OIDC_CLIENT_ID=<audience-client-id>`

## Endpoints

### `GET /`

Returns API status metadata.

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

If `version` is omitted, the API attempts to auto-detect it from filename (for example `v8.0.0`).
If `framework` is omitted or generic, the API infers benchmark type from filename and fuzzy-matches
against existing framework names. If similarity is `>= 95%`, it reuses the existing framework type.

### `GET /uploads`

Returns recent uploaded benchmark files with current framework/version tags.
Internal storage paths are not returned.

### `PUT /uploads/{id}/tag`

Updates an uploaded file's framework/version metadata and re-enqueues parse.

Request body:

```json
{
  "framework": "CIS Controls",
  "version": "4.0.0"
}
```

Use empty values to force auto-tag inference:

```json
{
  "framework": "",
  "version": ""
}
```

### `DELETE /uploads/{id}?purge=true`

Deletes uploaded file metadata and file from disk.

- `purge=true` also deletes parsed version data if no other uploads reference that same framework/version.

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
  "version_b": "8",
  "control_level": "ALL"
}
```

or with framework id:

```json
{
  "framework_id": 1,
  "version_a": "7",
  "version_b": "8",
  "control_level": "L1"
}
```

`control_level` options:

- `ALL` (default)
- `L1`
- `L2`

Response:

```json
{
  "report_id": 12,
  "status": "queued"
}
```

### `GET /diff/{report_id}`

Returns diff report metadata and diff items.

### `PATCH /diff/items/{item_id}/review`

Updates review state for a diff item.

Request body (any one or both fields):

```json
{
  "reviewed": true,
  "review_comment": "Validated with security engineering."
}
```

### `GET /reports`

Returns recent diff reports with framework/version labels and item counts.

### `GET /reports/{report_id}/download/{format}`

Downloads generated report export in one of:

- `json`
- `xlsx`
- `html`

## Settings Endpoints

### `GET /settings/branding`

Returns org branding settings.

### `PUT /settings/branding`

Upserts org branding settings.

```json
{
  "org_name": "",
  "logo_url": "",
  "primary_color": "",
  "secondary_color": "",
  "support_email": ""
}
```

### `GET /settings/roles`

Lists all roles.

### `POST /settings/roles`

Creates a role.

```json
{
  "name": "Security Analyst",
  "description": "Analyze benchmark deltas"
}
```

### `PUT /settings/roles/{id}`

Updates role name/description.

### `DELETE /settings/roles/{id}`

Deletes a non-system role not assigned to users.

### `GET /settings/users`

Lists users with role mapping.

### `POST /settings/users`

Creates a user.

```json
{
  "email": "analyst@example.com",
  "display_name": "Diff Analyst",
  "role_id": 2,
  "is_active": true
}
```

### `PUT /settings/users/{id}`

Updates user properties.

Use `"clear_role": true` to remove role assignment.

### `DELETE /settings/users/{id}`

Deletes a user.

## GPO Assessment Endpoints

### `POST /api/gpo/import`

Queues import of policy source exports.

```json
{
  "source_name": "Current RSOP",
  "source_path": "/data/uploads/gpresult.xml"
}
```

`source_type` is optional. If omitted, the service auto-discovers file type from extension/content.
Supported values when explicitly provided:

- `gpresult_xml`
- `gpmc_xml`
- `secedit_inf`
- `registry_pol`

### `POST /api/gpo/mappings/import`

Queues curated benchmark-to-policy mapping import from CSV or JSON.

```json
{
  "mapping_path": "/data/mappings/cis_windows_example_mapping.csv",
  "framework_id": 1,
  "version_id": 2,
  "mapping_label": "CIS Windows 11 v2.0.0"
}
```

### `POST /api/gpo/assess`

Queues a policy-vs-benchmark assessment run.

```json
{
  "policy_source_id": 1,
  "framework_id": 1,
  "version_id": 2,
  "mapping_label": "CIS Windows 11 v2.0.0"
}
```

### `GET /api/gpo/assessments`

Lists recent assessment runs.

### `GET /api/gpo/assessments/{assessment_id}`

Returns assessment metadata and detailed result rows.

### `GET /api/gpo/assessments/{assessment_id}/report/{format}`

Downloads generated exports in one of:

- `json`
- `md`
- `html`
- `csv`
- `xlsx`
- `docx`
