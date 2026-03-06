import React, { useEffect, useState } from "react";
import axios from "axios";
import {
  Alert,
  Button,
  Checkbox,
  FormControl,
  FormControlLabel,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import FileDropzone from "../components/FileDropzone";
import { fetchWorkflowCatalog } from "../api/workflowCatalog";

export default function UploadBenchmarks({ apiBase }) {
  const [framework, setFramework] = useState("CIS Controls");
  const [frameworks, setFrameworks] = useState([]);
  const [useCustomFramework, setUseCustomFramework] = useState(false);
  const [customFramework, setCustomFramework] = useState("");
  const [version, setVersion] = useState("");
  const [releaseDate, setReleaseDate] = useState("");
  const [isUploading, setIsUploading] = useState(false);
  const [uploads, setUploads] = useState([]);
  const [tagEdits, setTagEdits] = useState({});
  const [purgeOnDelete, setPurgeOnDelete] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const extractApiError = (requestError, fallback) => {
    const data = requestError?.response?.data;
    const status = requestError?.response?.status;
    if (typeof data?.error === "string" && data.error.trim()) {
      return status ? `${data.error} (HTTP ${status})` : data.error;
    }
    if (requestError?.request && !requestError?.response) {
      return `${fallback}: API not reachable at ${apiBase}`;
    }
    return fallback;
  };

  const requestWithFallback = async (method, primaryPath, secondaryPath, data, config = {}) => {
    try {
      return await axios({ method, url: `${apiBase}${primaryPath}`, data, ...config });
    } catch (primaryError) {
      if (primaryError?.response?.status !== 404 || !secondaryPath) {
        throw primaryError;
      }
      return axios({ method, url: `${apiBase}${secondaryPath}`, data, ...config });
    }
  };

  const loadUploads = async () => {
    try {
      const catalog = await fetchWorkflowCatalog(apiBase);
      const rows = catalog.uploads || [];
      setUploads(rows);
      setFrameworks(catalog.frameworks || []);
      setTagEdits((prev) => {
        const next = { ...prev };
        for (const row of rows) {
          if (!next[row.id]) {
            next[row.id] = {
              framework: row.framework || row.suggested_framework || "",
              version: row.version || row.suggested_version || "",
            };
          }
        }
        return next;
      });
    } catch {
      try {
        const [uploadRes, frameworkRes] = await Promise.all([
          axios.get(`${apiBase}/uploads`),
          axios.get(`${apiBase}/frameworks`),
        ]);
        const rows = uploadRes.data || [];
        setUploads(rows);
        setFrameworks(frameworkRes.data || []);
      } catch {
        setUploads([]);
        setFrameworks([]);
      }
    }
  };

  useEffect(() => {
    loadUploads();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const uploadSelectedFiles = async (selectedFiles) => {
    const files = Array.from(selectedFiles || []);
    if (files.length === 0) {
      return;
    }

    setError("");
    setMessage(`Uploading ${files.length} file(s)...`);
    setIsUploading(true);

    let uploadedCount = 0;
    const failures = [];

    const selectedFramework = useCustomFramework ? customFramework : framework;
    for (const selectedFile of files) {
      const formData = new FormData();
      formData.append("file", selectedFile);
      formData.append("framework", selectedFramework);
      formData.append("version", version);
      if (releaseDate) {
        formData.append("release_date", releaseDate);
      }

      try {
        await requestWithFallback("post", "/api/upload", "/upload", formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
        uploadedCount += 1;
      } catch (uploadError) {
        const reason = extractApiError(uploadError, "upload failed");
        failures.push(`${selectedFile.name}: ${reason}`);
      }
    }

    await loadUploads();
    setIsUploading(false);

    if (uploadedCount > 0) {
      setMessage(`Uploaded ${uploadedCount} of ${files.length} file(s).`);
    } else {
      setMessage("");
    }

    if (failures.length > 0) {
      const preview = failures.slice(0, 3).join(" | ");
      const suffix = failures.length > 3 ? ` (+${failures.length - 3} more)` : "";
      setError(preview + suffix);
    }
  };

  const saveTag = async (uploadId) => {
    const edit = tagEdits[uploadId] || {};
    setError("");
    setMessage("");

    try {
      const response = await requestWithFallback("put", `/uploads/${uploadId}/tag`, null, {
        framework: edit.framework,
        version: edit.version,
      });
      const matchInfo = response.data.matched_framework ? " (matched existing type)" : "";
      setMessage(`Saved metadata for upload #${uploadId}${matchInfo}`);
      loadUploads();
    } catch (tagError) {
      setError(extractApiError(tagError, "Failed to save metadata"));
    }
  };

  const autoTag = async (uploadId) => {
    setError("");
    setMessage("");

    try {
      const response = await requestWithFallback("put", `/uploads/${uploadId}/tag`, null, {
        framework: "",
        version: "",
      });
      const matchInfo = response.data.matched_framework ? " (matched existing type)" : "";
      setMessage(`Auto-tagged upload #${uploadId}${matchInfo}`);
      loadUploads();
    } catch (tagError) {
      setError(extractApiError(tagError, "Auto-tag failed"));
    }
  };

  const deleteUpload = async (uploadId) => {
    if (!window.confirm(`Delete upload #${uploadId}?`)) {
      return;
    }

    setError("");
    setMessage("");
    try {
      const response = await requestWithFallback("delete", `/uploads/${uploadId}`, null, null, {
        params: { purge: purgeOnDelete },
      });
      const purgeInfo = response.data.purged_version ? " and purged parsed version data" : "";
      setMessage(`Deleted upload #${uploadId}${purgeInfo}.`);
      loadUploads();
    } catch (deleteError) {
      setError(extractApiError(deleteError, "Delete failed"));
    }
  };

  const reparseUpload = async (uploadId) => {
    setError("");
    setMessage("");
    try {
      const response = await requestWithFallback("post", `/uploads/${uploadId}/reparse`, null, null);
      const warning = response?.data?.warning ? ` (${response.data.warning})` : "";
      setMessage(`Queued parse for upload #${uploadId}${warning}`);
    } catch (reparseError) {
      setError(extractApiError(reparseError, "Failed to queue parse"));
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Upload Benchmarks</Typography>
        <FormControl fullWidth>
          <InputLabel id="framework-select-label">Framework</InputLabel>
          <Select
            labelId="framework-select-label"
            label="Framework"
            value={useCustomFramework ? "__custom__" : framework}
            onChange={(event) => {
              const value = event.target.value;
              if (value === "__custom__") {
                setUseCustomFramework(true);
                return;
              }
              setUseCustomFramework(false);
              setFramework(value);
            }}
          >
            <MenuItem value="CIS Controls">CIS Controls</MenuItem>
            {frameworks.map((item) => (
              <MenuItem key={item.id} value={item.name}>
                {item.name}
              </MenuItem>
            ))}
            <MenuItem value="__custom__">Custom Framework</MenuItem>
          </Select>
        </FormControl>
        {useCustomFramework && (
          <TextField
            label="Custom Framework"
            value={customFramework}
            onChange={(event) => setCustomFramework(event.target.value)}
            fullWidth
          />
        )}
        <TextField label="Version (optional, auto-detected if blank)" value={version} onChange={(event) => setVersion(event.target.value)} fullWidth />
        <TextField
          label="Release Date"
          value={releaseDate}
          onChange={(event) => setReleaseDate(event.target.value)}
          type="date"
          InputLabelProps={{ shrink: true }}
          fullWidth
        />
        <FileDropzone onFilesSelected={uploadSelectedFiles} accepted=".xlsx,.xlsm,.csv,.pdf" disabled={isUploading} />
        {message && <Alert severity="success">{message}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}

        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Typography variant="h6">Uploaded Benchmarks</Typography>
          <Stack direction="row" spacing={1} alignItems="center">
            <FormControlLabel
              control={<Checkbox checked={purgeOnDelete} onChange={(event) => setPurgeOnDelete(event.target.checked)} />}
              label="Purge parsed data on delete"
            />
            <Button variant="outlined" onClick={loadUploads}>
              Refresh
            </Button>
          </Stack>
        </Stack>

        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>ID</TableCell>
              <TableCell>File</TableCell>
              <TableCell>Type Match</TableCell>
              <TableCell>Framework</TableCell>
              <TableCell>Version</TableCell>
              <TableCell>Uploaded</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {uploads.map((row) => (
              <TableRow key={row.id}>
                <TableCell>{row.id}</TableCell>
                <TableCell>{row.filename}</TableCell>
                <TableCell>
                  {row.suggested_framework}
                </TableCell>
                <TableCell sx={{ minWidth: 210 }}>
                  <TextField
                    size="small"
                    value={tagEdits[row.id]?.framework || ""}
                    onChange={(event) =>
                      setTagEdits((prev) => ({
                        ...prev,
                        [row.id]: { ...(prev[row.id] || {}), framework: event.target.value },
                      }))
                    }
                    fullWidth
                  />
                </TableCell>
                <TableCell sx={{ minWidth: 140 }}>
                  <TextField
                    size="small"
                    value={tagEdits[row.id]?.version || ""}
                    onChange={(event) =>
                      setTagEdits((prev) => ({
                        ...prev,
                        [row.id]: { ...(prev[row.id] || {}), version: event.target.value },
                      }))
                    }
                    fullWidth
                  />
                </TableCell>
                <TableCell>{new Date(row.created_at).toLocaleString()}</TableCell>
                <TableCell>
                  <Stack direction="row" spacing={1}>
                    <Button size="small" variant="outlined" onClick={() => autoTag(row.id)}>
                      Auto Tag
                    </Button>
                    <Button size="small" variant="outlined" onClick={() => saveTag(row.id)}>
                      Save
                    </Button>
                    <Button size="small" variant="outlined" onClick={() => reparseUpload(row.id)}>
                      Reparse
                    </Button>
                    <Button size="small" color="error" variant="outlined" onClick={() => deleteUpload(row.id)}>
                      Delete
                    </Button>
                  </Stack>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Stack>
    </Paper>
  );
}
