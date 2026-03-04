# API Usage

## Authentication

- Disabled by default.
- Enable OIDC verification with:
  - `AUTH_ENABLED=true`
  - `OIDC_ISSUER_URL=<issuer-url>`
  - `OIDC_CLIENT_ID=<audience-client-id>`

## Endpoints

### `GET /`

Returns API status and a `frontend_url` field for the web app.

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
  "org_name": "Acme Security",
  "logo_url": "https://example.com/logo.png",
  "primary_color": "#0b7285",
  "secondary_color": "#f59f00",
  "support_email": "security@example.com"
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
