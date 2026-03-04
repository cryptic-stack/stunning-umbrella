from __future__ import annotations

import json
import os
from collections import defaultdict
from pathlib import Path

import psycopg2
import xlsxwriter
from jinja2 import Environment, FileSystemLoader


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


def fetch_report(report_id: int):
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT
                    dr.id,
                    dr.framework_id,
                    dr.version_a,
                    dr.version_b,
                    COALESCE(NULLIF(dr.control_level, ''), 'ALL') AS control_level,
                    dr.status,
                    dr.created_at,
                    COALESCE(f.name, '') AS framework_name,
                    COALESCE(va.version, '') AS version_a_label,
                    COALESCE(vb.version, '') AS version_b_label
                FROM diff_reports dr
                LEFT JOIN frameworks f ON f.id = dr.framework_id
                LEFT JOIN versions va ON va.id = dr.version_a
                LEFT JOIN versions vb ON vb.id = dr.version_b
                WHERE dr.id = %s
                """,
                (report_id,),
            )
            report = cur.fetchone()
            if not report:
                raise ValueError(f"Report {report_id} does not exist")

            cur.execute(
                """
                SELECT change_type, safeguard_old, safeguard_new, old_text, new_text, similarity
                FROM diff_items
                WHERE report_id = %s
                ORDER BY id ASC
                """,
                (report_id,),
            )
            items = cur.fetchall()

    return report, items


def build_report_name(framework: str, version_a: str, version_b: str, control_level: str) -> str:
    framework = (framework or "").strip()
    version_a = (version_a or "").strip()
    version_b = (version_b or "").strip()
    control_level = (control_level or "ALL").strip().upper()

    if framework:
        base = f"{framework} v{version_a} -> v{version_b}"
    else:
        base = f"v{version_a} -> v{version_b}"

    if control_level and control_level != "ALL":
        return f"{base} ({control_level})"
    return base


def export_report(report_id: int, output_dir: str) -> dict:
    out = Path(output_dir)
    out.mkdir(parents=True, exist_ok=True)

    report, rows = fetch_report(report_id)

    payload = {
        "report": {
            "id": report[0],
            "framework_id": report[1],
            "version_a": report[2],
            "version_b": report[3],
            "control_level": report[4],
            "status": report[5],
            "created_at": str(report[6]),
            "framework": report[7],
            "version_a_label": report[8],
            "version_b_label": report[9],
            "report_name": build_report_name(report[7], report[8], report[9], report[4]),
        },
        "items": [
            {
                "change_type": row[0],
                "safeguard_old": row[1],
                "safeguard_new": row[2],
                "old_text": row[3],
                "new_text": row[4],
                "similarity": float(row[5]),
            }
            for row in rows
        ],
    }

    json_path = out / f"cis_diff_report_{report_id}.json"
    json_path.write_text(json.dumps(payload, indent=2), encoding="utf-8")

    workbook_path = out / f"cis_diff_report_{report_id}.xlsx"
    workbook = xlsxwriter.Workbook(workbook_path.as_posix())
    grouped = defaultdict(list)
    for item in payload["items"]:
        grouped[item["change_type"]].append(item)

    sheet_map = {
        "added": "added_controls",
        "removed": "removed_controls",
        "modified": "modified_controls",
        "renamed": "renamed_controls",
    }

    for change_type, sheet_name in sheet_map.items():
        worksheet = workbook.add_worksheet(sheet_name)
        worksheet.write_row(0, 0, ["safeguard_old", "safeguard_new", "old_text", "new_text", "similarity"])
        for row_num, item in enumerate(grouped.get(change_type, []), start=1):
            worksheet.write_row(
                row_num,
                0,
                [item["safeguard_old"], item["safeguard_new"], item["old_text"], item["new_text"], item["similarity"]],
            )

    workbook.close()

    env = Environment(loader=FileSystemLoader(Path(__file__).parent / "templates"), autoescape=True)
    template = env.get_template("report.html.j2")
    html_content = template.render(report=payload["report"], items=payload["items"])

    html_path = out / f"cis_diff_report_{report_id}.html"
    html_path.write_text(html_content, encoding="utf-8")

    return {"json": str(json_path), "excel": str(workbook_path), "html": str(html_path)}


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Export diff report to JSON/Excel/HTML")
    parser.add_argument("report_id", type=int)
    parser.add_argument("--output-dir", default="/data/exports")
    args = parser.parse_args()

    print(json.dumps(export_report(args.report_id, args.output_dir), indent=2))
