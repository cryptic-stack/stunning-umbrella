import React, { useEffect, useState } from "react";
import axios from "axios";
import { Alert, Button, FormControl, InputLabel, MenuItem, Paper, Select, Stack, Typography } from "@mui/material";

export default function GPOAssessment({ apiBase }) {
  const [sources, setSources] = useState([]);
  const [frameworks, setFrameworks] = useState([]);
  const [versions, setVersions] = useState([]);
  const [mappingLabels, setMappingLabels] = useState([]);
  const [policySourceId, setPolicySourceId] = useState("");
  const [frameworkId, setFrameworkId] = useState("");
  const [versionId, setVersionId] = useState("");
  const [mappingLabel, setMappingLabel] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const loadChoices = async () => {
    try {
      const [sourceRes, frameworkRes, mappingRes] = await Promise.all([
        axios.get(`${apiBase}/api/gpo/sources`),
        axios.get(`${apiBase}/frameworks`),
        axios.get(`${apiBase}/api/gpo/mappings`),
      ]);
      const loadedSources = sourceRes.data || [];
      const loadedMappings = mappingRes.data || [];
      setSources(loadedSources);
      setFrameworks(frameworkRes.data || []);
      setMappingLabels([...new Set(loadedMappings.map((item) => item.source_label).filter(Boolean))]);
      if (!policySourceId && loadedSources.length > 0) {
        setPolicySourceId(String(loadedSources[0].id));
      }
      if (!mappingLabel && loadedMappings.length > 0) {
        setMappingLabel(loadedMappings[0].source_label || "");
      }
    } catch {
      setSources([]);
      setFrameworks([]);
      setMappingLabels([]);
    }
  };

  useEffect(() => {
    loadChoices();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const loadVersions = async () => {
      if (!frameworkId) {
        setVersions([]);
        setVersionId("");
        return;
      }
      try {
        const response = await axios.get(`${apiBase}/frameworks/${frameworkId}/versions`);
        setVersions(response.data || []);
      } catch {
        setVersions([]);
      }
    };
    loadVersions();
  }, [apiBase, frameworkId]);

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
      setError(err?.response?.data?.error || "Failed to queue assessment");
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Step 3: Run Assessment</Typography>

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
            <MenuItem value=""><em>Any mapping</em></MenuItem>
            {mappingLabels.map((label) => (
              <MenuItem key={label} value={label}>{label}</MenuItem>
            ))}
          </Select>
        </FormControl>

        <Button variant="contained" onClick={runAssessment}>Queue Assessment</Button>
        {message && <Alert severity="success">{message}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </Paper>
  );
}

