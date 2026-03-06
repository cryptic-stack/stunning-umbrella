from mapper.import_mapping import _normalize_check_type, _parse_expected_value


def test_mapping_expected_value_parser_handles_json_and_scalars():
    assert _parse_expected_value('{"min":14}') == {"min": 14}
    assert _parse_expected_value("true") is True
    assert _parse_expected_value("10") == 10
    assert _parse_expected_value("text-value") == "text-value"


def test_mapping_check_type_normalization():
    assert _normalize_check_type("equals") == "exact"
    assert _normalize_check_type("bool") == "boolean"
    assert _normalize_check_type("threshold") == "numeric_threshold"
    assert _normalize_check_type("contains_all") == "set_membership"

