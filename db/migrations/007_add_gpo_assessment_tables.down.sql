DROP INDEX IF EXISTS idx_assessment_results_run_status;
DROP INDEX IF EXISTS idx_assessment_runs_status_created;
DROP INDEX IF EXISTS idx_benchmark_policy_rules_key;
DROP INDEX IF EXISTS idx_policy_settings_source_key;

DROP TABLE IF EXISTS assessment_results;
DROP TABLE IF EXISTS assessment_runs;
DROP TABLE IF EXISTS benchmark_policy_rules;
DROP TABLE IF EXISTS policy_settings;
DROP TABLE IF EXISTS policy_sources;
