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
- `frontend/Dockerfile`
- `docker/Dockerfile.collector` (optional collector profile)

## Data Volumes

- Uploaded files: `./data/uploads`
- Collected downloads: `./data/downloads`
- Postgres data: `./storage/postgres`

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
pytest
```
