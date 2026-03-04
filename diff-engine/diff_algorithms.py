from __future__ import annotations

import difflib
from typing import Dict, List

from deepdiff import DeepDiff
from rapidfuzz import fuzz

from diff_models import DiffResult, SafeguardRecord


def _combined_text(item: SafeguardRecord) -> str:
    return f"{item.title}\n{item.description}".strip()


def _similarity(a: SafeguardRecord, b: SafeguardRecord) -> float:
    ratio = fuzz.ratio(a.description or a.title, b.description or b.title)
    return float(ratio)


def compare_safeguards(
    safeguards_a: Dict[str, SafeguardRecord],
    safeguards_b: Dict[str, SafeguardRecord],
    rename_threshold: float = 85.0,
) -> List[DiffResult]:
    results: List[DiffResult] = []

    keys_a = set(safeguards_a.keys())
    keys_b = set(safeguards_b.keys())

    added = set(keys_b - keys_a)
    removed = set(keys_a - keys_b)

    renamed_pairs: list[tuple[str, str, float]] = []
    for old_id in list(removed):
        old_item = safeguards_a[old_id]
        best_new = None
        best_score = 0.0
        for new_id in list(added):
            score = _similarity(old_item, safeguards_b[new_id])
            if score > best_score:
                best_score = score
                best_new = new_id
        if best_new and best_score >= rename_threshold:
            renamed_pairs.append((old_id, best_new, best_score))
            removed.discard(old_id)
            added.discard(best_new)

    for old_id, new_id, score in renamed_pairs:
        old_item = safeguards_a[old_id]
        new_item = safeguards_b[new_id]

        # Skip rename-only noise when text is effectively identical.
        if score >= 100.0:
            continue

        results.append(
            DiffResult(
                change_type="renamed",
                safeguard_old=old_id,
                safeguard_new=new_id,
                old_text=_combined_text(old_item),
                new_text=_combined_text(new_item),
                similarity=score,
            )
        )

    for sid in sorted(added):
        item = safeguards_b[sid]
        results.append(
            DiffResult(
                change_type="added",
                safeguard_old="",
                safeguard_new=sid,
                old_text="",
                new_text=_combined_text(item),
                similarity=0.0,
            )
        )

    for sid in sorted(removed):
        item = safeguards_a[sid]
        results.append(
            DiffResult(
                change_type="removed",
                safeguard_old=sid,
                safeguard_new="",
                old_text=_combined_text(item),
                new_text="",
                similarity=0.0,
            )
        )

    shared = keys_a & keys_b
    for sid in sorted(shared):
        left = safeguards_a[sid]
        right = safeguards_b[sid]
        diff = DeepDiff(left.__dict__, right.__dict__, ignore_order=True)
        if diff:
            ratio = difflib.SequenceMatcher(None, _combined_text(left), _combined_text(right)).ratio() * 100
            results.append(
                DiffResult(
                    change_type="modified",
                    safeguard_old=sid,
                    safeguard_new=sid,
                    old_text=_combined_text(left),
                    new_text=_combined_text(right),
                    similarity=float(round(ratio, 2)),
                )
            )

    return results
