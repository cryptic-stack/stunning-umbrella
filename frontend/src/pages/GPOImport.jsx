import React, { useEffect, useState } from "react";
import axios from "axios";
import { Alert, Button, FormControl, InputLabel, MenuItem, Paper, Select, Stack, TextField, Typography } from "@mui/material";

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
    return `${fallbackMessage}: API not reachable. Verify API is running at ${window.location.protocol}//${window.location.hostname}:8080`;
  }
  if (err?.message) {
    return status ? `${fallbackMessage}: ${err.message} (HTTP ${status})` : `${fallbackMessage}: ${err.message}`;
  }
  return fallbackMessage;
}

export default function GPOImport({ apiBase, onBenchmarkContextChange }) {
  const [sourceName, setSourceName] = useState("Current RSOP");
  const [sourceFile, setSourceFile] = useState(null);
  const [uploads, setUploads] = useState([]);
  const [selectedUploadId, setSelectedUploadId] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const canQueueSource = Boolean(sourceFile);
  const selectedUpload = uploads.find((item) => String(item.id) === String(selectedUploadId));

  const loadUploads = async () => {
    try {
      const response = await axios.get(`${apiBase}/uploads`);
      const rows = response.data || [];
      setUploads(rows);
      if (!selectedUploadId && rows.length > 0) {
        setSelectedUploadId(String(rows[0].id));
      }
    } catch {
      setUploads([]);
    }
  };

  useEffect(() => {
    loadUploads();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiBase]);

  useEffect(() => {
    if (!onBenchmarkContextChange) {
      return;
    }
    if (!selectedUpload) {
      onBenchmarkContextChange(null);
      return;
    }
    onBenchmarkContextChange({
      uploadId: selectedUpload.id,
      framework: selectedUpload.framework || selectedUpload.suggested_framework || "",
      version: selectedUpload.version || selectedUpload.suggested_version || "",
      filename: selectedUpload.filename || "",
    });
  }, [onBenchmarkContextChange, selectedUpload]);

  const importSource = async () => {
    setMessage("");
    setError("");
    if (!sourceFile) {
      setError("Choose a policy source file.");
      return;
    }
    try {
      const formData = new FormData();
      formData.append("source_name", sourceName);
      formData.append("file", sourceFile);
      const response = await axios.post(`${apiBase}/api/gpo/import`, formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      setMessage(response.data.message || "GPO import queued");
    } catch (err) {
      setError(extractApiError(err, "Failed to queue GPO import"));
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Step 1: Import Policy Source</Typography>
        <Alert severity="info">Source type is auto-discovered from the uploaded file.</Alert>
        <Typography variant="caption" color="text.secondary">API endpoint: {apiBase}</Typography>
        <TextField label="Source Name" value={sourceName} onChange={(event) => setSourceName(event.target.value)} fullWidth />
        <Button component="label" variant="outlined">
          {sourceFile ? `Selected: ${sourceFile.name}` : "Choose Policy Source File"}
          <input type="file" hidden accept=".xml,.inf,.pol,.txt" onChange={(event) => setSourceFile(event.target.files?.[0] || null)} />
        </Button>
        {!sourceFile && <Alert severity="warning">Step 1 requires a policy source file before queueing import.</Alert>}
        <Button variant="contained" onClick={importSource} disabled={!canQueueSource}>Queue Policy Import</Button>

        <Typography variant="h6" sx={{ pt: 2 }}>Step 2: Select Uploaded Benchmark</Typography>
        <Alert severity="info">Benchmark files are managed in Benchmark Workflow. Pick one uploaded benchmark to scope assessment defaults.</Alert>
        <Stack direction="row" spacing={1}>
          <Button variant="outlined" onClick={loadUploads}>Refresh Benchmarks</Button>
        </Stack>
        <FormControl fullWidth>
          <InputLabel id="benchmark-select-label">Uploaded Benchmark</InputLabel>
          <Select
            labelId="benchmark-select-label"
            label="Uploaded Benchmark"
            value={selectedUploadId}
            onChange={(event) => setSelectedUploadId(event.target.value)}
          >
            {uploads.map((item) => (
              <MenuItem key={item.id} value={String(item.id)}>
                #{item.id} {item.framework || item.suggested_framework || "Unmapped"} {item.version ? `v${item.version}` : ""} - {item.filename}
              </MenuItem>
            ))}
            {uploads.length === 0 && (
              <MenuItem value="" disabled>
                No uploaded benchmarks found
              </MenuItem>
            )}
          </Select>
        </FormControl>
        {selectedUpload && (
          <Alert severity="success">
            Selected benchmark #{selectedUpload.id}: {selectedUpload.framework || selectedUpload.suggested_framework || "Unmapped"}{" "}
            {selectedUpload.version ? `v${selectedUpload.version}` : "(no version)"}.
          </Alert>
        )}
        {!selectedUpload && <Alert severity="warning">Select an uploaded benchmark to continue with assessment defaults.</Alert>}

        {message && <Alert severity="success">{message}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </Paper>
  );
}
