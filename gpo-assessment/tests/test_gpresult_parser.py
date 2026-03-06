from pathlib import Path

from importer.gpresult_xml import parse_gpresult_xml


def test_parse_gpresult_xml_normalizes_basic_policy_values(tmp_path):
    xml = """<?xml version="1.0" encoding="utf-8"?>
<Rsop>
  <Policy>
    <Name>Turn off multicast name resolution</Name>
    <State>Enabled</State>
  </Policy>
  <Policy>
    <Name>Minimum password length</Name>
    <State>14</State>
  </Policy>
  <Policy>
    <Name>Allow local log on</Name>
    <State>Administrators;Users</State>
  </Policy>
</Rsop>
"""
    path = tmp_path / "gpresult.xml"
    path.write_text(xml, encoding="utf-8")

    rows = parse_gpresult_xml(str(path))
    by_key = {row["setting_key"]: row for row in rows}

    assert by_key["turn_off_multicast_name_resolution"]["canonical_type"] == "boolean"
    assert by_key["turn_off_multicast_name_resolution"]["value_bool"] is True
    assert by_key["minimum_password_length"]["canonical_type"] == "numeric"
    assert by_key["minimum_password_length"]["value_number"] == 14.0
    assert by_key["allow_local_log_on"]["canonical_type"] == "set"
    assert by_key["allow_local_log_on"]["value_json"]["values"] == ["Administrators", "Users"]

