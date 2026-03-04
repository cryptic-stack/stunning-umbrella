import pandas as pd

from cis_excel_parser import parse_excel


def test_parse_excel_normalizes_rows(tmp_path):
    path = tmp_path / "sample.csv"
    pd.DataFrame(
        [
            {
                "Control ID": "1",
                "Safeguard ID": "1.1",
                "Title": "Establish Inventory",
                "Description": "Create and maintain an inventory",
                "IG1": "yes",
                "IG2": "true",
                "IG3": "0",
            }
        ]
    ).to_csv(path, index=False)

    records = parse_excel(str(path), framework="CIS Controls", version="8")
    assert len(records) == 1
    assert records[0].safeguard_id == "1.1"
    assert records[0].ig1 is True
    assert records[0].ig2 is True
    assert records[0].ig3 is False
