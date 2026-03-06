import React, { useEffect, useMemo, useState } from "react";
import axios from "axios";
import { Alert, Button, FormControl, InputLabel, MenuItem, Paper, Select, Stack, Typography } from "@mui/material";
import { fetchWorkflowCatalog } from "../api/workflowCatalog";

function extractApiError(err, fallbackMessage) {
  const status = err?.response?.status;
  const data = err?.response?.data;
  if (typeof data?.error === "string" && data.error.trim()) {
    return status ? `${data.error} (HTTP ${status})` : data.error;
  }
  if (typeof data === "string" && data.trim()) {
    const compact = data.replace(/\s+/g, " ").trim();
    return status ? `${compact} (HTTP ${status})` : compact;
  }
  if (err?.request && !err?.response) {
    return `${fallbackMessage}: API not reachable`;
  }
  if (err?.message) {
    return status ? `${fallbackMessage}: ${err.message} (HTTP ${status})` : `${fallbackMessage}: ${err.message}`;
  }
  return fallbackMessage;
}

export default function GPOAssessment({ apiBase, benchmarkContext, refreshToken }) {
  const [sources, setSources] = useState([]);
  const [frameworks, setFrameworks] = useState([]);
  const [versions, setVersions] = useState([]);
  const [policySourceId, setPolicySourceId] = useState("");
  const [frameworkId, setFrameworkId] = useState("");
  const [versionId, setVersionId] = useState("");
  const [controlLevel, setControlLevel] = useState("ALL");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const loadChoices = async (attempt = 1) => {
    try {
      const catalog = await fetchWorkflowCatalog(apiBase);
      const loadedSources = Array.isArray(catalog.gpo_sources) ? catalog.gpo_sources : [];
      const loadedFrameworks = Array.isArray(catalog.frameworks) ? catalog.frameworks : [];

      setSources(loadedSources);
      setFrameworks(loadedFrameworks);
      setPolicySourceId((previous) => {
        if (!loadedSources.length) {
          return "";
        }
        if (loadedSources.some((item) => String(item.id) === String(previous))) {
          return previous;
        }
        return String(loadedSources[0].id);
      });
      setError("");
      return;
    } catch {
      try {
        const [sourceRes, frameworkRes] = await Promise.all([
          axios.get(`${apiBase}/api/gpo/sources`),
          axios.get(`${apiBase}/api/frameworks`),
        ]);
        const loadedSources = Array.isArray(sourceRes.data) ? sourceRes.data : [];
        const loadedFrameworks = Array.isArray(frameworkRes.data) ? frameworkRes.data : [];
        setSources(loadedSources);
        setFrameworks(loadedFrameworks);
        setPolicySourceId((previous) => {
          if (!loadedSources.length) {
            return "";
          }
          if (loadedSources.some((item) => String(item.id) === String(previous))) {
            return previous;
          }
          return String(loadedSources[0].id);
        });
        setError("");
        return;
      } catch {
        if (attempt < 3) {
          await new Promise((resolve) => setTimeout(resolve, 700 * attempt));
          return loadChoices(attempt + 1);
        }
        setSources([]);
        setFrameworks([]);
        setPolicySourceId("");
        setError("Failed to load assessment context. Refresh and verify API services are healthy.");
      }
    }
  };

  useEffect(() => {
    loadChoices();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    loadChoices();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshToken, benchmarkContext]);

  useEffect(() => {
    const loadVersions = async () => {
      if (!frameworkId) {
        setVersions([]);
        setVersionId("");
        return;
      }
      try {
        const response = await axios.get(`${apiBase}/api/frameworks/${frameworkId}/versions`);
        setVersions(response.data || []);
      } catch {
        setVersions([]);
        setVersionId("");
      }
    };
    loadVersions();
  }, [apiBase, frameworkId]);

  useEffect(() => {
    if (!benchmarkContext?.framework) {
      setFrameworkId("");
      return;
    }
    const matchedFramework = frameworks.find(
      (item) => String(item.name || "").toLowerCase() === String(benchmarkContext.framework || "").toLowerCase()
    );
    setFrameworkId(matchedFramework ? String(matchedFramework.id) : "");
  }, [benchmarkContext, frameworks]);

  useEffect(() => {
    if (!benchmarkContext?.version || !versions.length) {
      setVersionId("");
      return;
    }
    const matchedVersion = versions.find((item) => String(item.version || "") === String(benchmarkContext.version || ""));
    setVersionId(matchedVersion ? String(matchedVersion.id) : "");
  }, [benchmarkContext, versions]);

  const latestSource = useMemo(() => {
    if (!policySourceId) {
      return null;
    }
    return sources.find((item) => String(item.id) === String(policySourceId)) || null;
  }, [policySourceId, sources]);

  const runAssessment = async () => {
    setMessage("");
    setError("");
    if (!policySourceId) {
      setError("Import a GPO source in Step 1 first.");
      return;
    }
    if (!frameworkId || !versionId) {
      setError("Select a benchmark in Step 2 so framework/version can be resolved.");
      return;
    }
    try {
      const response = await axios.post(`${apiBase}/api/gpo/assess`, {
        policy_source_id: Number(policySourceId),
        framework_id: Number(frameworkId),
        version_id: Number(versionId),
        control_level: controlLevel,
      });
      setMessage(`Assessment queued: #${response.data.assessment_run_id}`);
    } catch (err) {
      setError(extractApiError(err, "Failed to queue assessment"));
    }
  };

  const canQueueAssessment = Boolean(policySourceId && frameworkId && versionId);

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Step 3: Run Compare</Typography>
        <Alert severity="info">
          This compares the latest imported GPO settings against the benchmark selected in Step 2.
        </Alert>
        {latestSource ? (
          <Alert severity="success">
            Using latest policy source: #{latestSource.id} {latestSource.source_name || latestSource.source_type}
          </Alert>
        ) : (
          <Alert severity="warning">No policy source available yet. Import one in Step 1.</Alert>
        )}
        {benchmarkContext ? (
          <Alert severity="success">
            Using benchmark: #{benchmarkContext.uploadId} {benchmarkContext.framework || "Unmapped"}{" "}
            {benchmarkContext.version ? `v${benchmarkContext.version}` : "(no version)"}
          </Alert>
        ) : (
          <Alert severity="warning">No benchmark selected. Choose one in Step 2.</Alert>
        )}
        {benchmarkContext && (!frameworkId || !versionId) && (
          <Alert severity="warning">
            Could not resolve framework/version IDs for the selected benchmark. Re-tag that benchmark in Benchmark Workflow and refresh.
          </Alert>
        )}

        <FormControl fullWidth>
          <InputLabel id="cis-level-select-label">CIS Level</InputLabel>
          <Select
            labelId="cis-level-select-label"
            label="CIS Level"
            value={controlLevel}
            onChange={(event) => setControlLevel(event.target.value)}
          >
            <MenuItem value="ALL">All</MenuItem>
            <MenuItem value="L1">L1</MenuItem>
            <MenuItem value="L2">L2</MenuItem>
          </Select>
        </FormControl>

        <Stack direction="row" spacing={1}>
          <Button variant="outlined" onClick={loadChoices}>Refresh</Button>
          <Button variant="contained" onClick={runAssessment} disabled={!canQueueAssessment}>Run Compare</Button>
        </Stack>

        {message && <Alert severity="success">{message}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </Paper>
  );
}
