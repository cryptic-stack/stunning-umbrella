from __future__ import annotations

import argparse
import json
import os
import time
from typing import Dict

import psycopg2
import redis
from psycopg2.extras import RealDictCursor

from diff_algorithms import compare_safeguards
from diff_models import SafeguardRecord


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


def load_safeguards(version_id: int) -> Dict[str, SafeguardRecord]:
    with get_db_connection() as conn:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute(
                """
                SELECT safeguard_id, title, description, ig1, ig2, ig3
                FROM safeguards
                WHERE version_id = %s
                """,
                (version_id,),
            )
            rows = cur.fetchall()

    return {
        row["safeguard_id"]: SafeguardRecord(
            safeguard_id=row["safeguard_id"],
            title=row.get("title") or "",
            description=row.get("description") or "",
            ig1=bool(row.get("ig1")),
            ig2=bool(row.get("ig2")),
            ig3=bool(row.get("ig3")),
        )
        for row in rows
    }


def persist_diff(report_id: int, results) -> int:
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("DELETE FROM diff_items WHERE report_id = %s", (report_id,))

            for item in results:
                cur.execute(
                    """
                    INSERT INTO diff_items (
                        report_id,
                        change_type,
                        safeguard_old,
                        safeguard_new,
                        old_text,
                        new_text,
                        similarity
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s)
                    """,
                    (
                        report_id,
                        item.change_type,
                        item.safeguard_old,
                        item.safeguard_new,
                        item.old_text,
                        item.new_text,
                        item.similarity,
                    ),
                )

            cur.execute("UPDATE diff_reports SET status = 'completed', error = NULL WHERE id = %s", (report_id,))
        conn.commit()

    return len(results)


def mark_failed(report_id: int, message: str) -> None:
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("UPDATE diff_reports SET status = 'failed', error = %s WHERE id = %s", (message[:500], report_id))
        conn.commit()


def process_job(payload: dict) -> dict:
    report_id = int(payload["report_id"])
    version_a_id = int(payload["version_a_id"])
    version_b_id = int(payload["version_b_id"])

    left = load_safeguards(version_a_id)
    right = load_safeguards(version_b_id)
    results = compare_safeguards(left, right)
    count = persist_diff(report_id, results)
    return {"report_id": report_id, "items": count}


def run_worker() -> None:
    redis_client = redis.Redis(host=os.getenv("REDIS_HOST", "redis"), port=int(os.getenv("REDIS_PORT", "6379")), decode_responses=True)

    while True:
        try:
            job = redis_client.blpop("diff_jobs", timeout=5)
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
            report_id = None
            try:
                report_id = int(json.loads(payload).get("report_id"))
            except Exception:  # noqa: BLE001
                report_id = None
            if report_id:
                mark_failed(report_id, str(exc))
            print(json.dumps({"status": "error", "error": str(exc), "payload": payload}))
            time.sleep(1)


def main() -> None:
    parser = argparse.ArgumentParser(description="CIS diff engine")
    parser.add_argument("--framework-id", type=int)
    parser.add_argument("--version-a-id", type=int)
    parser.add_argument("--version-b-id", type=int)
    parser.add_argument("--report-id", type=int)
    parser.add_argument("--worker", action="store_true")
    args = parser.parse_args()

    if args.worker:
        run_worker()
        return

    if not all([args.framework_id, args.version_a_id, args.version_b_id, args.report_id]):
        raise SystemExit("Provide --framework-id --version-a-id --version-b-id --report-id or run with --worker")

    payload = {
        "framework_id": args.framework_id,
        "version_a_id": args.version_a_id,
        "version_b_id": args.version_b_id,
        "report_id": args.report_id,
    }
    result = process_job(payload)
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
