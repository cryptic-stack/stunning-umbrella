from diff_algorithms import compare_safeguards
from diff_models import SafeguardRecord


def test_compare_detects_renamed():
    left = {
        "1.1": SafeguardRecord(
            safeguard_id="1.1",
            title="Inventory",
            description="Establish inventory for enterprise assets",
            ig1=True,
            ig2=False,
            ig3=False,
        )
    }
    right = {
        "1.2": SafeguardRecord(
            safeguard_id="1.2",
            title="Asset inventory",
            description="Establish inventory for enterprise assets",
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
            ig1=True,
            ig2=False,
            ig3=False,
        )
    }

    results = compare_safeguards(left, right)
    assert len(results) == 1
    assert results[0].change_type == "modified"
