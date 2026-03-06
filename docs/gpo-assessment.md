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

## Notes

- `gpresult_xml` is the primary import path for milestone 1.
- `gpmc_xml` and `secedit_inf` are supported in the same canonical pipeline.
- `registry_pol` currently supports text-form exports (e.g. parsed output form), with binary parsing planned for a future enhancement.
