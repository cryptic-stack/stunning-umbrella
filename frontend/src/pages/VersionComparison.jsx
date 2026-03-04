import React, { useEffect, useMemo, useState } from "react";
import axios from "axios";
import { Alert, Button, MenuItem, Paper, Stack, TextField, Typography } from "@mui/material";

export default function VersionComparison({ apiBase, onReportCreated }) {
  const [frameworks, setFrameworks] = useState([]);
  const [frameworkId, setFrameworkId] = useState("");
  const [versions, setVersions] = useState([]);
  const [versionA, setVersionA] = useState("");
  const [versionB, setVersionB] = useState("");
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    axios
      .get(`${apiBase}/frameworks`)
      .then((response) => setFrameworks(response.data))
      .catch(() => setFrameworks([]));
  }, [apiBase]);

  useEffect(() => {
    if (!frameworkId) {
      setVersions([]);
      return;
    }

    axios
      .get(`${apiBase}/frameworks/${frameworkId}/versions`)
      .then((response) => setVersions(response.data))
      .catch(() => setVersions([]));
  }, [apiBase, frameworkId]);

  const options = useMemo(() => versions.map((item) => item.version), [versions]);

  const runCompare = async () => {
    setError("");
    setStatus("");

    try {
      const response = await axios.post(`${apiBase}/compare`, {
        framework_id: Number(frameworkId),
        version_a: versionA,
        version_b: versionB,
      });
      onReportCreated(String(response.data.report_id));
      setStatus(`Diff job queued. Report ID: ${response.data.report_id}`);
    } catch (compareError) {
      setError(compareError?.response?.data?.error || "Comparison failed.");
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Version Comparison</Typography>
        <TextField
          select
          label="Framework"
          value={frameworkId}
          onChange={(event) => setFrameworkId(event.target.value)}
          fullWidth
        >
          {frameworks.map((framework) => (
            <MenuItem key={framework.id} value={framework.id}>
              {framework.name}
            </MenuItem>
          ))}
        </TextField>
        <TextField select label="Version A" value={versionA} onChange={(event) => setVersionA(event.target.value)} fullWidth>
          {options.map((value) => (
            <MenuItem key={`a-${value}`} value={value}>
              {value}
            </MenuItem>
          ))}
        </TextField>
        <TextField select label="Version B" value={versionB} onChange={(event) => setVersionB(event.target.value)} fullWidth>
          {options.map((value) => (
            <MenuItem key={`b-${value}`} value={value}>
              {value}
            </MenuItem>
          ))}
        </TextField>
        <Button variant="contained" onClick={runCompare} disabled={!frameworkId || !versionA || !versionB}>
          Run Comparison
        </Button>
        {status && <Alert severity="success">{status}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </Paper>
  );
}
