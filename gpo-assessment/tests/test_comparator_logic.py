from comparator.compare import _evaluate


def test_exact_check():
    status, _ = _evaluate("exact", {"value_text": "Expected", "value_json": {}}, "Expected")
    assert status == "compliant"


def test_boolean_check():
    status, _ = _evaluate("boolean", {"value_bool": True, "value_text": "", "value_json": {}}, True)
    assert status == "compliant"


def test_numeric_threshold_check():
    status, _ = _evaluate("numeric_threshold", {"value_number": 12, "value_text": "", "value_json": {}}, {"min": 10})
    assert status == "compliant"
    status, _ = _evaluate("numeric_threshold", {"value_number": 8, "value_text": "", "value_json": {}}, {"min": 10})
    assert status == "noncompliant"


def test_set_membership_partial():
    status, _ = _evaluate("set_membership", {"value_json": {"values": ["users"]}, "value_text": ""}, {"values": ["users", "administrators"]})
    assert status == "partially_configured"

