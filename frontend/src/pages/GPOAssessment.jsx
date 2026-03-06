import React, { useState } from "react";
import axios from "axios";
import { Alert, Button, Paper, Stack, TextField, Typography } from "@mui/material";

export default function GPOAssessment({ apiBase }) {
  const [policySourceId, setPolicySourceId] = useState("");
  const [frameworkId, setFrameworkId] = useState("");
  const [versionId, setVersionId] = useState("");
  const [mappingLabel, setMappingLabel] = useState("CIS Windows mapping");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const runAssessment = async () => {
    setMessage("");
    setError("");
    try {
      const response = await axios.post(`${apiBase}/api/gpo/assess`, {
        policy_source_id: Number(policySourceId),
        framework_id: frameworkId ? Number(frameworkId) : null,
        version_id: versionId ? Number(versionId) : null,
        mapping_label: mappingLabel,
      });
      setMessage(`Assessment queued: #${response.data.assessment_run_id}`);
    } catch (err) {
      setError(err?.response?.data?.error || "Failed to queue assessment");
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Run GPO Assessment</Typography>
        <TextField label="Policy Source ID" value={policySourceId} onChange={(event) => setPolicySourceId(event.target.value)} fullWidth />
        <TextField label="Framework ID (optional)" value={frameworkId} onChange={(event) => setFrameworkId(event.target.value)} fullWidth />
        <TextField label="Version ID (optional)" value={versionId} onChange={(event) => setVersionId(event.target.value)} fullWidth />
        <TextField label="Mapping Label" value={mappingLabel} onChange={(event) => setMappingLabel(event.target.value)} fullWidth />
        <Button variant="contained" onClick={runAssessment}>Queue Assessment</Button>
        {message && <Alert severity="success">{message}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </Paper>
  );
}

