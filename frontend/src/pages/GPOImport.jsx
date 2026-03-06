import React, { useEffect, useState } from "react";
import axios from "axios";
import { Alert, Button, FormControl, InputLabel, MenuItem, Paper, Select, Stack, TextField, Typography } from "@mui/material";

const sourceTypes = [
  { value: "gpresult_xml", label: "GPResult XML" },
  { value: "gpmc_xml", label: "GPMC XML" },
  { value: "secedit_inf", label: "Secedit INF" },
  { value: "registry_pol", label: "Registry.pol" },
];

export default function GPOImport({ apiBase }) {
  const [sourceType, setSourceType] = useState("gpresult_xml");
  const [sourceName, setSourceName] = useState("Current RSOP");
  const [sourceFile, setSourceFile] = useState(null);
  const [mappingFile, setMappingFile] = useState(null);
  const [frameworks, setFrameworks] = useState([]);
  const [versions, setVersions] = useState([]);
  const [frameworkId, setFrameworkId] = useState("");
  const [versionId, setVersionId] = useState("");
  const [mappingLabel, setMappingLabel] = useState("CIS Windows mapping");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    const loadFrameworks = async () => {
      try {
        const response = await axios.get(`${apiBase}/frameworks`);
        setFrameworks(response.data || []);
      } catch {
        setFrameworks([]);
      }
    };
    loadFrameworks();
  }, [apiBase]);

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

  const importSource = async () => {
    setMessage("");
    setError("");
    if (!sourceFile) {
      setError("Choose a policy source file.");
      return;
    }
    try {
      const formData = new FormData();
      formData.append("source_type", sourceType);
      formData.append("source_name", sourceName);
      formData.append("file", sourceFile);
      const response = await axios.post(`${apiBase}/api/gpo/import`, formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      setMessage(response.data.message || "GPO import queued");
    } catch (err) {
      setError(err?.response?.data?.error || "Failed to queue GPO import");
    }
  };

  const importMapping = async () => {
    setMessage("");
    setError("");
    if (!mappingFile) {
      setError("Choose a mapping file.");
      return;
    }
    try {
      const formData = new FormData();
      formData.append("mapping_label", mappingLabel);
      if (frameworkId) {
        formData.append("framework_id", frameworkId);
      }
      if (versionId) {
        formData.append("version_id", versionId);
      }
      formData.append("file", mappingFile);
      const response = await axios.post(`${apiBase}/api/gpo/mappings/import`, formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      setMessage(response.data.message || "Mapping import queued");
    } catch (err) {
      setError(err?.response?.data?.error || "Failed to queue mapping import");
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Step 1: Import Policy Source</Typography>
        <FormControl fullWidth>
          <InputLabel id="source-type-label">Source Type</InputLabel>
          <Select labelId="source-type-label" label="Source Type" value={sourceType} onChange={(event) => setSourceType(event.target.value)}>
            {sourceTypes.map((item) => (
              <MenuItem key={item.value} value={item.value}>{item.label}</MenuItem>
            ))}
          </Select>
        </FormControl>
        <TextField label="Source Name" value={sourceName} onChange={(event) => setSourceName(event.target.value)} fullWidth />
        <Button component="label" variant="outlined">
          {sourceFile ? `Selected: ${sourceFile.name}` : "Choose Policy Source File"}
          <input type="file" hidden onChange={(event) => setSourceFile(event.target.files?.[0] || null)} />
        </Button>
        <Button variant="contained" onClick={importSource}>Queue Policy Import</Button>

        <Typography variant="h6" sx={{ pt: 2 }}>Step 2: Import Benchmark Mapping</Typography>
        <Button component="label" variant="outlined">
          {mappingFile ? `Selected: ${mappingFile.name}` : "Choose Mapping CSV/JSON"}
          <input type="file" hidden onChange={(event) => setMappingFile(event.target.files?.[0] || null)} />
        </Button>
        <FormControl fullWidth>
          <InputLabel id="framework-label">Framework</InputLabel>
          <Select labelId="framework-label" label="Framework" value={frameworkId} onChange={(event) => setFrameworkId(event.target.value)}>
            <MenuItem value=""><em>Any framework</em></MenuItem>
            {frameworks.map((item) => (
              <MenuItem key={item.id} value={String(item.id)}>{item.name}</MenuItem>
            ))}
          </Select>
        </FormControl>
        <FormControl fullWidth disabled={!frameworkId}>
          <InputLabel id="version-label">Version</InputLabel>
          <Select labelId="version-label" label="Version" value={versionId} onChange={(event) => setVersionId(event.target.value)}>
            <MenuItem value=""><em>Any version</em></MenuItem>
            {versions.map((item) => (
              <MenuItem key={item.id} value={String(item.id)}>{item.version}</MenuItem>
            ))}
          </Select>
        </FormControl>
        <TextField label="Mapping Label" value={mappingLabel} onChange={(event) => setMappingLabel(event.target.value)} fullWidth />
        <Button variant="contained" onClick={importMapping}>Queue Mapping Import</Button>

        {message && <Alert severity="success">{message}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </Paper>
  );
}

