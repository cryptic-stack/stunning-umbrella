from reporter.report import _render_html, _render_markdown


def test_report_renderers_include_core_fields():
    summary = {
        "assessment_run_id": 7,
        "source_name": "Current RSOP",
        "source_type": "gpresult_xml",
        "framework": "CIS Windows 11",
        "version": "2.0.0",
        "mapping_label": "example",
        "status": "completed",
        "counts": {"compliant": 2, "noncompliant": 1},
    }
    rows = [{"rule_id": "18.1.1", "setting_key": "turn_off_multicast_name_resolution", "status": "compliant", "details": "ok"}]
    markdown = _render_markdown(summary, rows)
    html = _render_html(summary, rows)

    assert "GPO Assessment Report #7" in markdown
    assert "turn_off_multicast_name_resolution" in markdown
    assert "<table" in html
    assert "18.1.1" in html

