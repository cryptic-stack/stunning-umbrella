import pytest

import parser as parser_worker


def test_resolve_allowed_upload_path_allows_file_in_upload_dir(tmp_path, monkeypatch):
    upload_dir = tmp_path / "uploads"
    upload_dir.mkdir(parents=True, exist_ok=True)
    benchmark_file = upload_dir / "benchmark.xlsx"
    benchmark_file.write_text("test")

    monkeypatch.setenv("UPLOAD_DIR", str(upload_dir))

    resolved = parser_worker.resolve_allowed_upload_path(str(benchmark_file))
    assert resolved == str(benchmark_file.resolve())


def test_resolve_allowed_upload_path_blocks_outside_file(tmp_path, monkeypatch):
    upload_dir = tmp_path / "uploads"
    upload_dir.mkdir(parents=True, exist_ok=True)
    outside_file = tmp_path / "outside.xlsx"
    outside_file.write_text("test")

    monkeypatch.setenv("UPLOAD_DIR", str(upload_dir))

    with pytest.raises(ValueError, match="outside upload directory"):
        parser_worker.resolve_allowed_upload_path(str(outside_file))


def test_process_job_uses_upload_context_path(tmp_path, monkeypatch):
    upload_dir = tmp_path / "uploads"
    upload_dir.mkdir(parents=True, exist_ok=True)
    benchmark_file = upload_dir / "benchmark.xlsx"
    benchmark_file.write_text("test")

    monkeypatch.setenv("UPLOAD_DIR", str(upload_dir))

    def fake_get_upload_context(upload_id):
        assert upload_id == 99
        return str(benchmark_file), "CIS Test", "1.2.0"

    captured = {}

    def fake_normalize(path, framework, version):
        captured["path"] = path
        captured["framework"] = framework
        captured["version"] = version
        return []

    monkeypatch.setattr(parser_worker, "get_upload_context", fake_get_upload_context)
    monkeypatch.setattr(parser_worker, "normalize_file", fake_normalize)
    monkeypatch.setattr(parser_worker, "upsert_records", lambda records, source_file, provided_version_id=None: 0)

    result = parser_worker.process_job(
        {
            "upload_id": 99,
            "file_path": "/tmp/untrusted.xlsx",
            "framework": "Untrusted",
            "version": "0.0.1",
            "version_id": 123,
        }
    )

    assert result == {"inserted": 0, "records": 0}
    assert captured["path"] == str(benchmark_file.resolve())
    assert captured["framework"] == "CIS Test"
    assert captured["version"] == "1.2.0"


def test_process_job_applies_upload_metadata_when_version_id_present(tmp_path, monkeypatch):
    upload_dir = tmp_path / "uploads"
    upload_dir.mkdir(parents=True, exist_ok=True)
    benchmark_file = upload_dir / "benchmark.csv"
    benchmark_file.write_text("ref,title\n1.1.1,test")

    monkeypatch.setenv("UPLOAD_DIR", str(upload_dir))

    def fake_get_upload_context(upload_id):
        assert upload_id == 7
        return str(benchmark_file), "CIS Server 2019", "1.3.0"

    class DummyRecord:
        def __init__(self):
            self.framework = "CSV Embedded Framework"
            self.version = "v999"

    records = [DummyRecord()]
    captured = {}

    monkeypatch.setattr(parser_worker, "get_upload_context", fake_get_upload_context)
    monkeypatch.setattr(parser_worker, "normalize_file", lambda *_: records)

    def fake_upsert(parsed_records, source_file, provided_version_id=None):
        captured["framework"] = parsed_records[0].framework
        captured["version"] = parsed_records[0].version
        captured["version_id"] = provided_version_id
        captured["source_file"] = source_file
        return 1

    monkeypatch.setattr(parser_worker, "upsert_records", fake_upsert)

    result = parser_worker.process_job({"upload_id": 7, "version_id": 55})
    assert result == {"inserted": 1, "records": 1}
    assert captured["framework"] == "CIS Server 2019"
    assert captured["version"] == "1.3.0"
    assert captured["version_id"] == 55
    assert captured["source_file"] == str(benchmark_file.resolve())
