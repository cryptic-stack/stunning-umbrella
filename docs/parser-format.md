# Parser Format

## Canonical JSON

Each parsed safeguard is normalized into:

```json
{
  "framework": "CIS Controls",
  "version": "8",
  "control_id": "1",
  "safeguard_id": "1.1",
  "title": "Establish and Maintain Detailed Enterprise Asset Inventory",
  "description": "...",
  "ig1": true,
  "ig2": true,
  "ig3": false
}
```

## Parser Modules

- `cis_excel_parser.py`
  - Handles `.xlsx` and `.csv`
  - Resolves common alias headers (control id, safeguard id, description, IG flags)
- `cis_pdf_parser.py`
  - Extracts line text with `pdfplumber`
  - Detects safeguards using regex pattern `\d+\.\d+`

## Worker

`parser.py --worker`:

- Waits on Redis `parse_jobs`
- Normalizes file content
- Upserts framework/version/control/safeguard rows
