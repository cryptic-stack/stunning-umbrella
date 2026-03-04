import React, { useState } from "react";
import axios from "axios";
import { Alert, Box, Button, Paper, Stack, TextField, Typography } from "@mui/material";
import FileDropzone from "../components/FileDropzone";

export default function UploadBenchmarks({ apiBase }) {
  const [framework, setFramework] = useState("CIS Controls");
  const [version, setVersion] = useState("");
  const [releaseDate, setReleaseDate] = useState("");
  const [file, setFile] = useState(null);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const submit = async () => {
    if (!file) {
      setError("Select a file before uploading.");
      return;
    }

    const formData = new FormData();
    formData.append("file", file);
    formData.append("framework", framework);
    formData.append("version", version);
    if (releaseDate) {
      formData.append("release_date", releaseDate);
    }

    setError("");
    setMessage("");

    try {
      const response = await axios.post(`${apiBase}/api/upload`, formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      setMessage(`Uploaded: ${response.data.path}`);
    } catch (uploadError) {
      setError(uploadError?.response?.data?.error || "Upload failed.");
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Upload Benchmarks</Typography>
        <TextField label="Framework" value={framework} onChange={(event) => setFramework(event.target.value)} fullWidth />
        <TextField label="Version" value={version} onChange={(event) => setVersion(event.target.value)} fullWidth />
        <TextField
          label="Release Date"
          value={releaseDate}
          onChange={(event) => setReleaseDate(event.target.value)}
          type="date"
          InputLabelProps={{ shrink: true }}
          fullWidth
        />
        <FileDropzone onFileSelected={setFile} accepted=".xlsx,.csv,.pdf" />
        <Box>
          <Button variant="contained" onClick={submit}>
            Upload
          </Button>
        </Box>
        {message && <Alert severity="success">{message}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </Paper>
  );
}
