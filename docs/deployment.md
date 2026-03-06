# Deployment

## Local Docker Deployment

From repo root:

```bash
docker compose up --build
```

## Service Build Files

- `docker/Dockerfile.api`
- `docker/Dockerfile.parser`
- `docker/Dockerfile.diff`
- `docker/Dockerfile.gpo`
- `frontend/Dockerfile`
- `docker/Dockerfile.collector` (optional collector profile)

## Data Volumes

- Uploaded files, downloads, and exports are stored in Docker volume `app_data`.
- Postgres data is stored in Docker volume `postgres_data`.
- No runtime benchmark/report database files are written into the repository by default.

## Network Exposure

- API and frontend ports are bound to `127.0.0.1` only.
- Postgres and Redis are not published to the host and are reachable only from other containers in the same compose project.

## Optional Collector

```bash
docker compose --profile collector up --build
```

## Testing

- Go API tests:

```bash
cd api && go test ./...
```

- Python tests:

```bash
pip install -r parser/requirements.txt
pip install -r diff-engine/requirements.txt
pip install -r gpo-assessment/requirements.txt
pytest
```
