# Diff Logic

## Inputs

- `framework_id`
- `version_a`
- `version_b`

## Algorithm Steps

1. Load safeguards for both versions.
2. Key records by `safeguard_id`.
3. Detect baseline changes:
   - `added` (in B only)
   - `removed` (in A only)
   - `modified` (same ID, different fields)
4. Detect likely renames:
   - Compare removed-vs-added candidates with `rapidfuzz.ratio(description_a, description_b)`
   - If score `>= 85`, mark as `renamed`
5. Persist to:
   - `diff_reports` (`status = completed|failed`)
   - `diff_items`

## Libraries

- `deepdiff` for object-level change detection
- `rapidfuzz` for fuzzy similarity
- `difflib` for text similarity percentage
