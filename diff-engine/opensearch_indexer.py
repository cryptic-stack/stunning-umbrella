from __future__ import annotations

import os
from typing import Dict, List


def index_diff_items(items: List[dict]) -> None:
    endpoint = os.getenv("OPENSEARCH_URL")
    index_name = os.getenv("OPENSEARCH_INDEX", "cis-diff-history")
    if not endpoint:
        return

    try:
        from opensearchpy import OpenSearch  # type: ignore
    except Exception as exc:  # noqa: BLE001
        raise RuntimeError("opensearch-py is required for OpenSearch indexing") from exc

    client = OpenSearch(endpoint)

    if not client.indices.exists(index=index_name):
        client.indices.create(index=index_name)

    for item in items:
        body: Dict[str, object] = {
            "framework": item.get("framework"),
            "version": item.get("version"),
            "safeguard_id": item.get("safeguard_new") or item.get("safeguard_old"),
            "change_type": item.get("change_type"),
            "description": item.get("new_text") or item.get("old_text"),
        }
        client.index(index=index_name, body=body)
