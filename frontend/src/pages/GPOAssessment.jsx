import React, { useEffect, useState } from "react";
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
  const [mappings, setMappings] = useState([]);
  const [policySourceId, setPolicySourceId] = useState("");
  const [frameworkId, setFrameworkId] = useState("");
  const [versionId, setVersionId] = useState("");
  const [mappingLabel, setMappingLabel] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const loadChoices = async () => {
    try {
      const catalog = await fetchWorkflowCatalog(apiBase);
      const loadedSources = catalog.gpo_sources || [];
      const loadedMappings = catalog.gpo_mappings || [];
      setSources(loadedSources);
      setFrameworks(catalog.frameworks || []);
      setMappings(loadedMappings);
      if (loadedSources.length > 0 && (!policySourceId || !loadedSources.some((item) => String(item.id) === String(policySourceId)))) {
        setPolicySourceId(String(loadedSources[0].id));
      }
      if (!mappingLabel && loadedMappings.length > 0) {
        setMappingLabel(loadedMappings[0].source_label || "");
      }
    } catch {
      try {
        const [sourceRes, frameworkRes, mappingRes] = await Promise.all([
          axios.get(`${apiBase}/api/gpo/sources`),
          axios.get(`${apiBase}/api/frameworks`),
          axios.get(`${apiBase}/api/gpo/mappings`),
        ]);
        const loadedSources = sourceRes.data || [];
        const loadedMappings = mappingRes.data || [];
        setSources(loadedSources);
        setFrameworks(frameworkRes.data || []);
        setMappings(loadedMappings);
      } catch {
        setSources([]);
        setFrameworks([]);
        setMappings([]);
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
    const selectedMapping = mappings.find((item) => item.source_label === mappingLabel);
    if (!selectedMapping) {
      return;
    }
    if (!frameworkId && selectedMapping.framework_id) {
      setFrameworkId(String(selectedMapping.framework_id));
    }
  }, [mappingLabel, mappings, frameworkId]);

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
      }
    };
    loadVersions();
  }, [apiBase, frameworkId]);

  useEffect(() => {
    const selectedMapping = mappings.find((item) => item.source_label === mappingLabel);
    if (!selectedMapping) {
      return;
    }
    if (!versionId && selectedMapping.version_id) {
      setVersionId(String(selectedMapping.version_id));
    }
  }, [mappingLabel, mappings, versionId]);

  useEffect(() => {
    if (!benchmarkContext) {
      return;
    }

    if (!frameworkId && benchmarkContext.framework) {
      const matchedFramework = frameworks.find(
        (item) => String(item.name || "").toLowerCase() === String(benchmarkContext.framework || "").toLowerCase()
      );
      if (matchedFramework) {
        setFrameworkId(String(matchedFramework.id));
      }
    }
  }, [benchmarkContext, frameworks, frameworkId]);

  useEffect(() => {
    if (!benchmarkContext?.version) {
      return;
    }
    if (!versionId) {
      const matchedVersion = versions.find((item) => String(item.version) === String(benchmarkContext.version));
      if (matchedVersion) {
        setVersionId(String(matchedVersion.id));
      }
    }
  }, [benchmarkContext, versions, versionId]);

  useEffect(() => {
    if (!benchmarkContext || !mappings.length || mappingLabel) {
      return;
    }

    const selectedFramework = frameworks.find((item) => String(item.id) === String(frameworkId));
    const frameworkName = selectedFramework?.name || benchmarkContext.framework || "";
    const versionText = benchmarkContext.version || "";

    const candidate = mappings.find((item) => {
      const sameFramework = !item.framework_id || frameworks.some((row) => String(row.id) === String(item.framework_id) && String(row.name).toLowerCase() === String(frameworkName).toLowerCase());
      const sameVersion = !item.version_id || versions.some((row) => String(row.id) === String(item.version_id) && String(row.version) === String(versionText));
      return sameFramework && sameVersion;
    });

    if (candidate?.source_label) {
      setMappingLabel(candidate.source_label);
    }
  }, [benchmarkContext, frameworks, frameworkId, mappingLabel, mappings, versions]);

  const runAssessment = async () => {
    setMessage("");
    setError("");
    if (!policySourceId) {
      setError("Select a policy source.");
      return;
    }
    try {
      const response = await axios.post(`${apiBase}/api/gpo/assess`, {
        policy_source_id: Number(policySourceId),
        framework_id: frameworkId ? Number(frameworkId) : null,
        version_id: versionId ? Number(versionId) : null,
        mapping_label: mappingLabel || "",
      });
      setMessage(`Assessment queued: #${response.data.assessment_run_id}`);
    } catch (err) {
      setError(extractApiError(err, "Failed to queue assessment"));
    }
  };

  const canQueueAssessment = Boolean(policySourceId);
  const mappingLabels = [...new Set(mappings.map((item) => item.source_label).filter((value) => value !== null && value !== undefined))];

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Step 3: Run Assessment</Typography>
        {benchmarkContext && (
          <Alert severity="info">
            Benchmark context from Step 2: #{benchmarkContext.uploadId} {benchmarkContext.framework || "Unmapped"}{" "}
            {benchmarkContext.version ? `v${benchmarkContext.version}` : "(no version)"}.
          </Alert>
        )}
        <Stack direction="row" spacing={1}>
          <Button variant="outlined" onClick={loadChoices}>Refresh Sources/Mappings</Button>
        </Stack>
        {!policySourceId && <Alert severity="warning">Step 3 requires a Policy Source selection.</Alert>}
        {!mappingLabel && <Alert severity="info">Mapping Label is optional. If left blank, assessment uses selected framework/version rules.</Alert>}

        <FormControl fullWidth>
          <InputLabel id="source-select-label">Policy Source</InputLabel>
          <Select labelId="source-select-label" label="Policy Source" value={policySourceId} onChange={(event) => setPolicySourceId(event.target.value)}>
            {sources.map((item) => (
              <MenuItem key={item.id} value={String(item.id)}>
                #{item.id} {item.source_name || item.source_type}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        <FormControl fullWidth>
          <InputLabel id="framework-select-label">Framework</InputLabel>
          <Select labelId="framework-select-label" label="Framework" value={frameworkId} onChange={(event) => setFrameworkId(event.target.value)}>
            <MenuItem value=""><em>Any framework</em></MenuItem>
            {frameworks.map((item) => (
              <MenuItem key={item.id} value={String(item.id)}>{item.name}</MenuItem>
            ))}
          </Select>
        </FormControl>

        <FormControl fullWidth disabled={!frameworkId}>
          <InputLabel id="version-select-label">Version</InputLabel>
          <Select labelId="version-select-label" label="Version" value={versionId} onChange={(event) => setVersionId(event.target.value)}>
            <MenuItem value=""><em>Any version</em></MenuItem>
            {versions.map((item) => (
              <MenuItem key={item.id} value={String(item.id)}>{item.version}</MenuItem>
            ))}
          </Select>
        </FormControl>

        <FormControl fullWidth>
          <InputLabel id="mapping-select-label">Mapping Label</InputLabel>
          <Select labelId="mapping-select-label" label="Mapping Label" value={mappingLabel} onChange={(event) => setMappingLabel(event.target.value)}>
            <MenuItem value=""><em>Any mapping label</em></MenuItem>
            {mappingLabels.map((label) => (
              <MenuItem key={label || "__empty__"} value={label || ""}>{label || "(Unlabeled mapping)"}</MenuItem>
            ))}
          </Select>
        </FormControl>

        <Button variant="contained" onClick={runAssessment} disabled={!canQueueAssessment}>Queue Assessment</Button>
        {message && <Alert severity="success">{message}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </Paper>
  );
}
