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


def test_parse_benchmark_style_rows_with_levels(tmp_path):
    path = tmp_path / "benchmark.csv"
    pd.DataFrame(
        [
            {
                "Section #": "1.1",
                "Recommendation #": "1.1.1",
                "Title": "(L1) Ensure test setting is configured",
                "Description": "Benchmark recommendation text",
                "v8 IG1": "X",
                "v8 IG2": "",
                "v8 IG3": "",
            },
            {
                "Section #": "1.1",
                "Recommendation #": "1.1.2",
                "Title": "(L2) Ensure stronger setting is configured",
                "Description": "Second recommendation",
                "v8 IG1": "",
                "v8 IG2": "X",
                "v8 IG3": "",
            },
        ]
    ).to_csv(path, index=False)

    records = parse_excel(str(path), framework="CIS Windows Benchmark", version="4.0.0")
    assert len(records) == 2
    assert records[0].safeguard_id == "1.1.1"
    assert records[0].control_id == "1.1"
    assert records[0].level == "L1"
    assert records[1].level == "L2"
