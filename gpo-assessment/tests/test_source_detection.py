from importer.detect import detect_source_type


def test_detect_source_type_by_declared_type(tmp_path):
    path = tmp_path / "sample.xml"
    path.write_text("<Rsop></Rsop>", encoding="utf-8")
    assert detect_source_type(str(path), "gpmc_xml") == "gpmc_xml"


def test_detect_gpresult_xml_from_content(tmp_path):
    path = tmp_path / "gpresult.xml"
    path.write_text("<Rsop><ComputerResults></ComputerResults></Rsop>", encoding="utf-8")
    assert detect_source_type(str(path)) == "gpresult_xml"


def test_detect_gpmc_xml_from_content(tmp_path):
    path = tmp_path / "gpmc.xml"
    path.write_text("<GPO><Policy></Policy></GPO>", encoding="utf-8")
    assert detect_source_type(str(path)) == "gpmc_xml"


def test_detect_secedit_inf_from_extension(tmp_path):
    path = tmp_path / "security.inf"
    path.write_text("[System Access]\nMinimumPasswordLength = 14\n", encoding="utf-8")
    assert detect_source_type(str(path)) == "secedit_inf"


def test_detect_registry_pol_from_binary_signature(tmp_path):
    path = tmp_path / "Registry.pol"
    path.write_bytes(b"PReg\x01\x00\x00\x00")
    assert detect_source_type(str(path)) == "registry_pol"
