from __future__ import annotations

import argparse
import json
import os
import time
from pathlib import Path
from typing import Iterable, List

import psycopg2
import redis

try:
    from .cis_excel_parser import parse_excel
    from .cis_pdf_parser import parse_pdf
    from .models import CanonicalSafeguard
except ImportError:  # pragma: no cover
    from cis_excel_parser import parse_excel
    from cis_pdf_parser import parse_pdf
    from models import CanonicalSafeguard


def get_db_connection():
    url = os.getenv("DATABASE_URL")
    if url and url.startswith("postgres://"):
        return psycopg2.connect(url)

    return psycopg2.connect(
        host=os.getenv("POSTGRES_HOST", "postgres"),
        port=int(os.getenv("POSTGRES_PORT", "5432")),
        dbname=os.getenv("POSTGRES_DB", "cisdb"),
        user=os.getenv("POSTGRES_USER", "cis"),
        password=os.getenv("POSTGRES_PASSWORD", "cis"),
    )


def normalize_file(path: str, framework: str, version: str) -> List[CanonicalSafeguard]:
    ext = Path(path).suffix.lower()
    if ext in {".xlsx", ".csv"}:
        return parse_excel(path, framework, version)
    if ext == ".pdf":
        return parse_pdf(path, framework, version)
    raise ValueError(f"Unsupported extension: {ext}")


def resolve_allowed_upload_path(path: str) -> str:
    upload_root = Path(os.getenv("UPLOAD_DIR", "/data/uploads")).resolve()
    resolved_path = Path(path).resolve()

    if not resolved_path.is_file():
        raise ValueError("uploaded file not found")
    if not resolved_path.is_relative_to(upload_root):
        raise ValueError("job file path is outside upload directory")

    return str(resolved_path)


def get_upload_context(upload_id: int) -> tuple[str, str, str]:
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT stored_path, framework, version
                FROM uploaded_files
                WHERE id = %s
                """,
                (upload_id,),
            )
            row = cur.fetchone()
            if not row:
                raise ValueError("upload record not found")
            stored_path, framework, version = row
            return str(stored_path or ""), str(framework or ""), str(version or "")


def ensure_framework_version(cur, framework: str, version: str, source_file: str) -> tuple[int, int]:
    cur.execute(
        """
        INSERT INTO frameworks (name)
        VALUES (%s)
        ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
        RETURNING id
        """,
        (framework,),
    )
    framework_id = cur.fetchone()[0]

    cur.execute(
        """
        INSERT INTO versions (framework_id, version, source_file)
        VALUES (%s, %s, %s)
        ON CONFLICT (framework_id, version)
        DO UPDATE SET source_file = EXCLUDED.source_file
        RETURNING id
        """,
        (framework_id, version, source_file),
    )
    version_id = cur.fetchone()[0]
    return framework_id, version_id


def ensure_schema_columns(cur) -> None:
    cur.execute("ALTER TABLE safeguards ADD COLUMN IF NOT EXISTS level TEXT DEFAULT ''")


def upsert_records(records: Iterable[CanonicalSafeguard], source_file: str, provided_version_id: int | None = None) -> int:
    records = list(records)
    if not records:
        return 0

    inserted = 0
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            ensure_schema_columns(cur)
            framework = records[0].framework
            version = records[0].version
            framework_id, version_id = ensure_framework_version(cur, framework, version, source_file)
            if provided_version_id:
                version_id = provided_version_id

            for record in records:
                control_id = record.control_id or record.safeguard_id.split(".", 1)[0]
                level = (record.level or "").strip().upper()
                safeguard_key = record.safeguard_id if not level else f"{record.safeguard_id}|{level}"
                cur.execute(
                    """
                    INSERT INTO controls (framework_id, version_id, control_id, title, description)
                    VALUES (%s, %s, %s, %s, %s)
                    ON CONFLICT (version_id, control_id)
                    DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description
                    RETURNING id
                    """,
                    (framework_id, version_id, control_id, record.title, record.description),
                )
                control_pk = cur.fetchone()[0]

                cur.execute(
                    """
                    INSERT INTO safeguards (
                        control_id, version_id, safeguard_id, title, description, level, ig1, ig2, ig3
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
                    ON CONFLICT (version_id, safeguard_id)
                    DO UPDATE SET
                      control_id = EXCLUDED.control_id,
                      title = EXCLUDED.title,
                      description = EXCLUDED.description,
                      level = EXCLUDED.level,
                      ig1 = EXCLUDED.ig1,
                      ig2 = EXCLUDED.ig2,
                      ig3 = EXCLUDED.ig3
                    """,
                    (
                        control_pk,
                        version_id,
                        safeguard_key,
                        record.title,
                        record.description,
                        level,
                        record.ig1,
                        record.ig2,
                        record.ig3,
                    ),
                )
                inserted += 1
    return inserted


def process_job(payload: dict) -> dict:
    framework = payload.get("framework") or "CIS Controls"
    version = payload.get("version") or "unknown"
    version_id = payload.get("version_id")
    upload_id = payload.get("upload_id")

    path = payload.get("file_path")
    if upload_id:
        stored_path, stored_framework, stored_version = get_upload_context(int(upload_id))
        path = stored_path
        if stored_framework:
            framework = stored_framework
        if stored_version:
            version = stored_version
    if not path:
        raise ValueError("job is missing upload path")

    safe_path = resolve_allowed_upload_path(path)

    records = normalize_file(safe_path, framework, version)
    inserted = upsert_records(records, safe_path, provided_version_id=version_id)
    return {"inserted": inserted, "records": len(records)}


def run_worker() -> None:
    redis_client = redis.Redis(host=os.getenv("REDIS_HOST", "redis"), port=int(os.getenv("REDIS_PORT", "6379")), decode_responses=True)

    while True:
        try:
            job = redis_client.blpop("parse_jobs", timeout=5)
        except Exception as exc:  # noqa: BLE001
            print(json.dumps({"status": "error", "error": f"redis unavailable: {exc}"}))
            time.sleep(2)
            continue
        if not job:
            continue

        _, payload = job
        try:
            data = json.loads(payload)
            result = process_job(data)
            print(json.dumps({"status": "ok", "result": result, "job": data}))
        except Exception as exc:  # noqa: BLE001
            print(json.dumps({"status": "error", "error": str(exc), "payload": payload}))
            time.sleep(1)


def main() -> None:
    parser = argparse.ArgumentParser(description="CIS benchmark parser")
    parser.add_argument("--file", help="Path to source file")
    parser.add_argument("--framework", default="CIS Controls")
    parser.add_argument("--version", default="unknown")
    parser.add_argument("--version-id", type=int, default=None)
    parser.add_argument("--output", choices=["json"], default="json")
    parser.add_argument("--worker", action="store_true")
    args = parser.parse_args()

    if args.worker:
        run_worker()
        return

    if not args.file:
        raise SystemExit("--file is required unless --worker is set")

    records = normalize_file(args.file, args.framework, args.version)
    upserted = upsert_records(records, args.file, args.version_id)

    if args.output == "json":
        print(json.dumps({
            "count": len(records),
            "inserted": upserted,
            "records": [record.model_dump() for record in records],
        }, indent=2))


if __name__ == "__main__":
    main()
