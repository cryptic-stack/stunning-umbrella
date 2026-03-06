from importer.gpmc_xml import parse_gpmc_xml
from importer.registry_pol import parse_registry_pol
from importer.secedit_inf import parse_secedit_inf


def test_parse_secedit_inf(tmp_path):
    path = tmp_path / "secedit.inf"
    path.write_text("[System Access]\nMinimumPasswordLength = 14\n", encoding="utf-8")
    rows = parse_secedit_inf(str(path))
    assert rows[0]["setting_key"] == "minimumpasswordlength"
    assert rows[0]["value_number"] == 14.0


def test_parse_gpmc_xml(tmp_path):
    path = tmp_path / "gpmc.xml"
    path.write_text("<Root><Policy><Name>Test Policy</Name><State>Enabled</State></Policy></Root>", encoding="utf-8")
    rows = parse_gpmc_xml(str(path))
    assert rows[0]["setting_key"] == "test_policy"
    assert rows[0]["value_bool"] is True


def test_parse_registry_pol_text(tmp_path):
    path = tmp_path / "registry.pol"
    path.write_text("RequireSecuritySignature=1\n", encoding="utf-8")
    rows = parse_registry_pol(str(path))
    assert rows[0]["setting_key"] == "requiresecuritysignature"
    assert rows[0]["value_bool"] is True

