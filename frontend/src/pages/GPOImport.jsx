import React, { useState } from "react";
import axios from "axios";
import { Alert, Button, Paper, Stack, TextField, Typography } from "@mui/material";

export default function GPOImport({ apiBase }) {
  const [sourceType, setSourceType] = useState("gpresult_xml");
  const [sourceName, setSourceName] = useState("Current RSOP");
  const [sourceFile, setSourceFile] = useState(null);
  const [sourcePath, setSourcePath] = useState("");
  const [mappingPath, setMappingPath] = useState("");
  const [mappingFile, setMappingFile] = useState(null);
  const [frameworkId, setFrameworkId] = useState("");
  const [versionId, setVersionId] = useState("");
  const [mappingLabel, setMappingLabel] = useState("CIS Windows mapping");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const importSource = async () => {
    setMessage("");
    setError("");
    try {
      let response;
      if (sourceFile) {
        const formData = new FormData();
        formData.append("source_type", sourceType);
        formData.append("source_name", sourceName);
        formData.append("file", sourceFile);
        response = await axios.post(`${apiBase}/api/gpo/import`, formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
      } else {
        response = await axios.post(`${apiBase}/api/gpo/import`, {
          source_type: sourceType,
          source_name: sourceName,
          source_path: sourcePath,
        });
      }
      setMessage(response.data.message || "GPO import queued");
    } catch (err) {
      setError(err?.response?.data?.error || "Failed to queue GPO import");
    }
  };

  const importMapping = async () => {
    setMessage("");
    setError("");
    try {
      let response;
      if (mappingFile) {
        const formData = new FormData();
        formData.append("mapping_label", mappingLabel);
        if (frameworkId) {
          formData.append("framework_id", frameworkId);
        }
        if (versionId) {
          formData.append("version_id", versionId);
        }
        formData.append("file", mappingFile);
        response = await axios.post(`${apiBase}/api/gpo/mappings/import`, formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
      } else {
        response = await axios.post(`${apiBase}/api/gpo/mappings/import`, {
          mapping_path: mappingPath,
          framework_id: frameworkId ? Number(frameworkId) : null,
          version_id: versionId ? Number(versionId) : null,
          mapping_label: mappingLabel,
        });
      }
      setMessage(response.data.message || "Mapping import queued");
    } catch (err) {
      setError(err?.response?.data?.error || "Failed to queue mapping import");
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">GPO Import</Typography>
        <TextField label="Source Type" value={sourceType} onChange={(event) => setSourceType(event.target.value)} helperText="gpresult_xml, gpmc_xml, secedit_inf, registry_pol" fullWidth />
        <TextField label="Source Name" value={sourceName} onChange={(event) => setSourceName(event.target.value)} fullWidth />
        <Button component="label" variant="outlined">
          {sourceFile ? `Selected: ${sourceFile.name}` : "Choose Source File"}
          <input type="file" hidden onChange={(event) => setSourceFile(event.target.files?.[0] || null)} />
        </Button>
        <TextField label="Or Source File Path (optional)" value={sourcePath} onChange={(event) => setSourcePath(event.target.value)} fullWidth />
        <Button variant="contained" onClick={importSource}>Queue GPO Import</Button>

        <Typography variant="h6" sx={{ pt: 2 }}>Curated Mapping Import</Typography>
        <Button component="label" variant="outlined">
          {mappingFile ? `Selected: ${mappingFile.name}` : "Choose Mapping File"}
          <input type="file" hidden onChange={(event) => setMappingFile(event.target.files?.[0] || null)} />
        </Button>
        <TextField label="Or Mapping File Path (.csv/.json)" value={mappingPath} onChange={(event) => setMappingPath(event.target.value)} fullWidth />
        <TextField label="Framework ID (optional)" value={frameworkId} onChange={(event) => setFrameworkId(event.target.value)} fullWidth />
        <TextField label="Version ID (optional)" value={versionId} onChange={(event) => setVersionId(event.target.value)} fullWidth />
        <TextField label="Mapping Label" value={mappingLabel} onChange={(event) => setMappingLabel(event.target.value)} fullWidth />
        <Button variant="contained" onClick={importMapping}>Queue Mapping Import</Button>

        {message && <Alert severity="success">{message}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </Paper>
  );
}
