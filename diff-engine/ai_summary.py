from __future__ import annotations

from typing import Iterable


def summarize_changes(items: Iterable[dict]) -> str:
    items = list(items)
    if not items:
        return "No changes were detected between these versions."

    try:
        from llama_index.core import Document, VectorStoreIndex  # type: ignore
        from langchain.prompts import PromptTemplate  # type: ignore
    except Exception:
        added = sum(1 for item in items if item.get("change_type") == "added")
        removed = sum(1 for item in items if item.get("change_type") == "removed")
        modified = sum(1 for item in items if item.get("change_type") in {"modified", "renamed"})
        return (
            f"Detected {added} added safeguards, {removed} removed safeguards, "
            f"and {modified} modified or renamed safeguards."
        )

    docs = [
        Document(
            text=(
                f"Type: {item.get('change_type')}; "
                f"Safeguard: {item.get('safeguard_new') or item.get('safeguard_old')}; "
                f"Old: {item.get('old_text', '')}; New: {item.get('new_text', '')}"
            )
        )
        for item in items
    ]

    index = VectorStoreIndex.from_documents(docs)
    query_engine = index.as_query_engine()

    prompt = PromptTemplate.from_template(
        "Summarize the key CIS safeguard changes in plain language, highlight impactful updates."
    )
    response = query_engine.query(prompt.template)
    return str(response)
