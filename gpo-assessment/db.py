from __future__ import annotations

import json
import os
from contextlib import contextmanager
from typing import Iterable


def get_db_connection():
    import psycopg2

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


@contextmanager
def db_cursor():
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            yield conn, cur


def upsert_policy_source(
    source_type: str,
    source_name: str,
    hostname: str,
    domain_name: str,
    raw_path: str,
    metadata: dict,
):
    with db_cursor() as (conn, cur):
        cur.execute(
            """
            INSERT INTO policy_sources (source_type, source_name, hostname, domain_name, raw_path, metadata)
            VALUES (%s, %s, %s, %s, %s, %s::jsonb)
            RETURNING id
            """,
            (source_type, source_name, hostname, domain_name, raw_path, json.dumps(metadata)),
        )
        source_id = int(cur.fetchone()[0])
        conn.commit()
        return source_id


def replace_policy_settings(source_id: int, settings: Iterable[dict]) -> int:
    with db_cursor() as (conn, cur):
        cur.execute("DELETE FROM policy_settings WHERE policy_source_id = %s", (source_id,))
        inserted = 0
        for setting in settings:
            cur.execute(
                """
                INSERT INTO policy_settings (
                    policy_source_id, setting_key, setting_name, canonical_type, scope,
                    value_text, value_number, value_bool, value_json
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s::jsonb)
                """,
                (
                    source_id,
                    setting.get("setting_key", ""),
                    setting.get("setting_name", ""),
                    setting.get("canonical_type", ""),
                    setting.get("scope", ""),
                    setting.get("value_text", ""),
                    setting.get("value_number"),
                    setting.get("value_bool"),
                    json.dumps(setting.get("value_json", {})),
                ),
            )
            inserted += 1
        conn.commit()
        return inserted
