import React, { useEffect, useMemo, useState } from "react";
import axios from "axios";
import { Alert, Button, MenuItem, Paper, Stack, TextField, Typography } from "@mui/material";

export default function VersionComparison({ apiBase, onReportCreated }) {
  const [frameworks, setFrameworks] = useState([]);
  const [uploads, setUploads] = useState([]);
  const [frameworkId, setFrameworkId] = useState("");
  const [versions, setVersions] = useState([]);
  const [versionA, setVersionA] = useState("");
  const [versionB, setVersionB] = useState("");
  const [controlLevel, setControlLevel] = useState("ALL");
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");

  const refreshCatalog = async () => {
    try {
      const [frameworkRes, uploadRes] = await Promise.all([axios.get(`${apiBase}/frameworks`), axios.get(`${apiBase}/uploads`)]);
      setFrameworks(frameworkRes.data || []);
      setUploads(uploadRes.data || []);
    } catch {
      setFrameworks([]);
      setUploads([]);
    }
  };

  useEffect(() => {
    refreshCatalog();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiBase]);

  useEffect(() => {
    if (!frameworkId) {
      setVersions([]);
      return;
    }

    axios
      .get(`${apiBase}/frameworks/${frameworkId}/versions`)
      .then((response) => setVersions(response.data || []))
      .catch(() => setVersions([]));
  }, [apiBase, frameworkId]);

  const selectedFramework = useMemo(
    () => frameworks.find((framework) => String(framework.id) === String(frameworkId)),
    [frameworks, frameworkId]
  );

  const options = useMemo(() => {
    const set = new Set((versions || []).map((item) => item.version));
    if (selectedFramework?.name) {
      for (const upload of uploads) {
        if (upload.framework === selectedFramework.name && upload.version) {
          set.add(upload.version);
        }
      }
    }
    return Array.from(set);
  }, [versions, uploads, selectedFramework]);

  const runCompare = async () => {
    setError("");
    setStatus("");

    try {
      const response = await axios.post(`${apiBase}/compare`, {
        framework_id: Number(frameworkId),
        version_a: versionA,
        version_b: versionB,
        control_level: controlLevel,
      });
      onReportCreated(String(response.data.report_id));
      const label = response.data.report_name || `Report ${response.data.report_id}`;
      setStatus(`Diff job queued (${controlLevel}): ${label}`);
    } catch (compareError) {
      setError(compareError?.response?.data?.error || "Comparison failed.");
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Typography variant="h6">Version Comparison</Typography>
          <Button variant="outlined" onClick={refreshCatalog}>
            Refresh Benchmarks
          </Button>
        </Stack>

        <TextField
          select
          label="Framework"
          value={frameworkId}
          onChange={(event) => {
            setFrameworkId(event.target.value);
            setVersionA("");
            setVersionB("");
          }}
          fullWidth
        >
          {frameworks.map((framework) => (
            <MenuItem key={framework.id} value={framework.id}>
              {framework.name}
            </MenuItem>
          ))}
        </TextField>
        <TextField
          select
          label="Version A"
          value={versionA}
          onChange={(event) => setVersionA(event.target.value)}
          helperText={options.length === 0 ? "No versions found. Upload or tag benchmarks first." : ""}
          fullWidth
        >
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
        <TextField select label="Control Level" value={controlLevel} onChange={(event) => setControlLevel(event.target.value)} fullWidth>
          <MenuItem value="ALL">All</MenuItem>
          <MenuItem value="L1">L1</MenuItem>
          <MenuItem value="L2">L2</MenuItem>
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
