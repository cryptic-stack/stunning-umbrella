# GPO Assessment

## Overview

The `gpo-assessment` module imports Windows policy exports, normalizes settings, applies curated CIS benchmark mappings, evaluates compliance, and generates remediation-ready reports.

## Module layout

- `gpo-assessment/importer/`: source parsers (`gpresult_xml`, `gpmc_xml`, `secedit_inf`, `registry_pol`)
- `gpo-assessment/mapper/`: benchmark-to-policy mapping import
- `gpo-assessment/comparator/`: rule evaluation engine
- `gpo-assessment/reporter/`: Markdown/HTML/DOCX/CSV/XLSX/JSON exports
- `gpo-assessment/worker.py`: Redis queue worker (`gpo_jobs`)

## Supported checks

- `exact`
- `boolean`
- `numeric_threshold`
- `set_membership`

## Result statuses

- `compliant`
- `noncompliant`
- `unknown`
- `not_applicable`
- `partially_configured`

## API workflow

1. `POST /api/gpo/import`
2. `POST /api/gpo/mappings/import`
3. `POST /api/gpo/assess`
4. `GET /api/gpo/assessments/{id}`
5. `GET /api/gpo/assessments/{id}/report/{format}`

## Step-by-step (UI)

1. Open `GPO Workflow` tab.
2. In **Step 1: Import Policy Source**:
   Source type is auto-discovered from file content/extension.
   Choose policy source file (`gpresult.xml` first).
   Click `Queue Policy Import`.
3. In **Step 2: Import Benchmark Mapping**:
   Choose mapping CSV/JSON file.
   Select `Framework` and `Version` from dropdowns (optional but recommended).
   Click `Queue Mapping Import`.
4. In the same `GPO Workflow` page, use **Step 3: Run Assessment**:
   Select `Policy Source`, `Framework`, `Version`, and `Mapping Label` from dropdowns.
   Click `Queue Assessment`.
5. Open top-level `Reports` tab, then sub-tab `GPO Reports` (**Step 4**):
   Click `Refresh`.
   Select an assessment from dropdown.
   Click `Load Assessment Details`.
   Download exports from table links (`DOCX`, `XLSX`, etc.).

## Benchmark workflow alignment

The benchmark process uses the same simplified pattern:

1. Open `Benchmark Workflow`.
2. Upload/tag benchmark files in Step 1 section.
3. Run version comparison in Step 2 section (dropdown-led framework/version selection).
4. Open `Reports` -> `Diff Viewer` to review/export benchmark diffs.

## Notes

- `gpresult_xml` is the primary import path for milestone 1.
- `gpmc_xml` and `secedit_inf` are supported in the same canonical pipeline.
- `registry_pol` currently supports text-form exports (e.g. parsed output form), with binary parsing planned for a future enhancement.
