from diff_algorithms import compare_safeguards
from diff_engine import filter_by_level
from diff_models import SafeguardRecord


def test_compare_detects_renamed():
    left = {
        "1.1": SafeguardRecord(
            safeguard_id="1.1",
            title="Inventory",
            description="Establish inventory for enterprise assets",
            level="L1",
            ig1=True,
            ig2=False,
            ig3=False,
        )
    }
    right = {
        "1.2": SafeguardRecord(
            safeguard_id="1.2",
            title="Asset inventory",
            description="Establish inventory for enterprise assets and systems",
            level="L1",
            ig1=True,
            ig2=False,
            ig3=False,
        )
    }

    results = compare_safeguards(left, right)
    assert len(results) == 1
    assert results[0].change_type == "renamed"
    assert results[0].similarity >= 85


def test_compare_detects_modified():
    left = {
        "1.1": SafeguardRecord(
            safeguard_id="1.1",
            title="Inventory",
            description="Old description",
            level="L1",
            ig1=True,
            ig2=False,
            ig3=False,
        )
    }
    right = {
        "1.1": SafeguardRecord(
            safeguard_id="1.1",
            title="Inventory",
            description="New description",
            level="L1",
            ig1=True,
            ig2=False,
            ig3=False,
        )
    }

    results = compare_safeguards(left, right)
    assert len(results) == 1
    assert results[0].change_type == "modified"


def test_compare_skips_renamed_with_identical_text():
    left = {
        "1.1|L1": SafeguardRecord(
            safeguard_id="1.1|L1",
            title="Inventory",
            description="Same text",
            level="L1",
            ig1=True,
            ig2=False,
            ig3=False,
        )
    }
    right = {
        "1.2|L1": SafeguardRecord(
            safeguard_id="1.2|L1",
            title="Inventory",
            description="Same text",
            level="L1",
            ig1=True,
            ig2=False,
            ig3=False,
        )
    }

    results = compare_safeguards(left, right)
    assert len(results) == 0


def test_filter_by_level():
    records = {
        "1.1|L1": SafeguardRecord("1.1|L1", "A", "A", "L1", True, False, False),
        "1.2|L2": SafeguardRecord("1.2|L2", "B", "B", "L2", True, False, False),
    }
    l1 = filter_by_level(records, "L1")
    l2 = filter_by_level(records, "L2")
    assert len(l1) == 1
    assert len(l2) == 1
