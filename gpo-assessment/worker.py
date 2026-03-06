from __future__ import annotations

import json
import os
import time
from datetime import datetime, timezone

import redis

from comparator.compare import run_assessment
from db import db_cursor, replace_policy_settings, upsert_policy_source
from importer.detect import detect_source_type
from importer.gpmc_xml import parse_gpmc_xml
from importer.gpresult_xml import parse_gpresult_xml
from importer.registry_pol import parse_registry_pol
from importer.secedit_inf import parse_secedit_inf
from mapper.import_mapping import import_mapping_file
from reporter.report import export_assessment_report


def _import_policy_source(payload: dict) -> dict:
    source_path = payload.get("source_path") or ""
    if not source_path:
        raise ValueError("source_path is required")
    source_type = detect_source_type(
        path=source_path,
        declared_type=(payload.get("source_type") or ""),
    )
    source_name = (payload.get("source_name") or "").strip()

    if source_type == "gpresult_xml":
        settings = parse_gpresult_xml(source_path)
    elif source_type == "gpmc_xml":
        settings = parse_gpmc_xml(source_path)
    elif source_type == "secedit_inf":
        settings = parse_secedit_inf(source_path)
    elif source_type == "registry_pol":
        settings = parse_registry_pol(source_path)
    else:
        raise ValueError(f"unsupported source_type: {source_type}")

    source_id = upsert_policy_source(
        source_type=source_type,
        source_name=source_name or source_type,
        hostname=(payload.get("hostname") or "").strip(),
        domain_name=(payload.get("domain_name") or "").strip(),
        raw_path=source_path,
        metadata=payload.get("metadata") or {},
    )
    setting_count = replace_policy_settings(source_id, settings)
    return {"policy_source_id": source_id, "settings_imported": setting_count}


def _import_mapping(payload: dict) -> dict:
    mapping_path = payload.get("mapping_path") or ""
    if not mapping_path:
        raise ValueError("mapping_path is required")
    framework_id = payload.get("framework_id")
    version_id = payload.get("version_id")
    inserted = import_mapping_file(
        path=mapping_path,
        framework_id=int(framework_id) if framework_id else None,
        version_id=int(version_id) if version_id else None,
        source_label=(payload.get("mapping_label") or "").strip(),
    )
    return {"rules_imported": inserted}


def _create_assessment_run(payload: dict) -> int:
    control_level = (payload.get("control_level") or "ALL").strip().upper()
    if control_level not in {"ALL", "L1", "L2"}:
        control_level = "ALL"
    with db_cursor() as (conn, cur):
        cur.execute(
            """
            INSERT INTO assessment_runs (policy_source_id, framework_id, version_id, mapping_label, control_level, status, error)
            VALUES (%s, %s, %s, %s, %s, 'running', '')
            RETURNING id
            """,
            (
                int(payload.get("policy_source_id")),
                int(payload.get("framework_id")) if payload.get("framework_id") else None,
                int(payload.get("version_id")) if payload.get("version_id") else None,
                (payload.get("mapping_label") or "").strip(),
                control_level,
            ),
        )
        run_id = int(cur.fetchone()[0])
        conn.commit()
        return run_id


def _run_assessment(payload: dict) -> dict:
    if payload.get("assessment_run_id"):
        run_id = int(payload["assessment_run_id"])
        with db_cursor() as (conn, cur):
            cur.execute("UPDATE assessment_runs SET status = 'running', error = '' WHERE id = %s", (run_id,))
            conn.commit()
    else:
        run_id = _create_assessment_run(payload)

    result = run_assessment(run_id)
    exports = export_assessment_report(run_id, os.getenv("EXPORT_DIR", "/data/exports"))
    return {"assessment_run_id": run_id, "result": result, "exports": exports}


def _mark_assessment_failed(run_id: int, message: str) -> None:
    with db_cursor() as (conn, cur):
        cur.execute(
            "UPDATE assessment_runs SET status = 'failed', error = %s, completed_at = %s WHERE id = %s",
            (message[:500], datetime.now(timezone.utc), run_id),
        )
        conn.commit()


def process_job(payload: dict) -> dict:
    job_type = (payload.get("job_type") or "").strip().lower()
    if job_type == "import_policy_source":
        return _import_policy_source(payload)
    if job_type == "import_mapping":
        return _import_mapping(payload)
    if job_type == "run_assessment":
        return _run_assessment(payload)
    raise ValueError(f"unsupported job_type: {job_type}")


def run_worker() -> None:
    redis_client = redis.Redis(host=os.getenv("REDIS_HOST", "redis"), port=int(os.getenv("REDIS_PORT", "6379")), decode_responses=True)
    queue_name = os.getenv("GPO_QUEUE_NAME", "gpo_jobs")

    while True:
        try:
            job = redis_client.blpop(queue_name, timeout=5)
        except Exception as exc:  # noqa: BLE001
            print(json.dumps({"status": "error", "error": f"redis unavailable: {exc}"}))
            time.sleep(2)
            continue
        if not job:
            continue

        _, payload = job
        run_id = None
        try:
            data = json.loads(payload)
            if data.get("assessment_run_id"):
                run_id = int(data["assessment_run_id"])
            result = process_job(data)
            print(json.dumps({"status": "ok", "result": result, "job": data}))
        except Exception as exc:  # noqa: BLE001
            if run_id:
                _mark_assessment_failed(run_id, str(exc))
            print(json.dumps({"status": "error", "error": str(exc), "payload": payload}))
            time.sleep(1)


if __name__ == "__main__":
    run_worker()
