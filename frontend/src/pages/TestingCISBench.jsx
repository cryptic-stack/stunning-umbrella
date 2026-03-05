import React, { useEffect, useRef, useState } from "react";
import axios from "axios";
import {
  Alert,
  Button,
  Checkbox,
  Chip,
  Collapse,
  FormControlLabel,
  LinearProgress,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";

const defaultSearch = {
  query: "",
  platform_type: "",
  latest: true,
  limit: 25,
};

const downloadFormatOptions = [
  { id: "csv", label: "CSV" },
];

export default function TestingCISBench({ apiBase }) {
  const [status, setStatus] = useState({ logged_in: false });
  const [cookiesText, setCookiesText] = useState("");
  const [manualCookieEditorOpen, setManualCookieEditorOpen] = useState(false);
  const [noVerifySSL, setNoVerifySSL] = useState(true);
  const [searchReq, setSearchReq] = useState(defaultSearch);
  const [searchResults, setSearchResults] = useState([]);
  const [downloadBenchmarkId, setDownloadBenchmarkId] = useState("");
  const [downloadFormats, setDownloadFormats] = useState(["csv"]);
  const [files, setFiles] = useState([]);
  const [downloadProgress, setDownloadProgress] = useState({
    active: false,
    value: 0,
    label: "",
    phase: "idle",
  });
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const progressIntervalRef = useRef(null);
  const progressCompletionRef = useRef(null);

  const extractApiError = (requestError, fallback) => {
    const data = requestError?.response?.data;
    if (!data) {
      return fallback;
    }
    const parts = [data.error, data.hint, data.stderr].filter((part) => Boolean(String(part || "").trim()));
    return parts.length > 0 ? parts.join(" | ") : fallback;
  };

  const loadStatus = async () => {
    try {
      const response = await axios.get(`${apiBase}/testing/cis-bench/status`);
      setStatus(response.data || { logged_in: false });
    } catch (statusError) {
      setStatus({ logged_in: false });
      setError(statusError?.response?.data?.error || "Failed to load cis-bench status.");
    }
  };

  const loadFiles = async () => {
    try {
      const response = await axios.get(`${apiBase}/testing/cis-bench/files`);
      setFiles(response.data || []);
    } catch {
      setFiles([]);
    }
  };

  useEffect(() => {
    loadStatus();
    loadFiles();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiBase]);

  useEffect(
    () => () => {
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
      }
      if (progressCompletionRef.current) {
        clearTimeout(progressCompletionRef.current);
      }
    },
    []
  );

  const clearMessages = () => {
    setMessage("");
    setError("");
  };

  const formatFileSize = (bytes) => {
    const value = Number(bytes || 0);
    if (value < 1024) {
      return `${value} B`;
    }
    const units = ["KB", "MB", "GB", "TB"];
    let size = value / 1024;
    let index = 0;
    while (size >= 1024 && index < units.length - 1) {
      size /= 1024;
      index += 1;
    }
    return `${size.toFixed(size >= 10 ? 1 : 2)} ${units[index]}`;
  };

  const startDownloadProgress = (label) => {
    if (progressIntervalRef.current) {
      clearInterval(progressIntervalRef.current);
    }
    if (progressCompletionRef.current) {
      clearTimeout(progressCompletionRef.current);
    }
    setDownloadProgress({ active: true, value: 6, label, phase: "downloading" });
    progressIntervalRef.current = setInterval(() => {
      setDownloadProgress((current) => {
        if (!current.active) {
          return current;
        }
        if (current.value >= 90) {
          return { ...current, phase: "finalizing" };
        }
        const increment = Math.max(1, Math.round((100 - current.value) * 0.08));
        const nextValue = Math.min(90, current.value + increment);
        return { ...current, value: nextValue, phase: "downloading" };
      });
    }, 700);
  };

  const finishDownloadProgress = () => {
    if (progressIntervalRef.current) {
      clearInterval(progressIntervalRef.current);
      progressIntervalRef.current = null;
    }
    setDownloadProgress((current) => ({ ...current, value: 100, phase: "complete" }));
    progressCompletionRef.current = setTimeout(() => {
      setDownloadProgress({ active: false, value: 0, label: "", phase: "idle" });
      progressCompletionRef.current = null;
    }, 500);
  };

  const loginWithCookiesInput = async (cookieInput, successMessage = "Logged in to cis-bench.") => {
    clearMessages();
    const cookiePayload = String(cookieInput || "").trim();
    if (!cookiePayload) {
      setError("Cookie input is empty.");
      return;
    }

    setBusy(true);
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/login`, {
        mode: "cookies",
        cookies_text: cookiePayload,
        no_verify_ssl: noVerifySSL,
      });
      setMessage(response.data?.message || successMessage);
      await loadStatus();
    } catch (loginError) {
      setError(extractApiError(loginError, "cis-bench login failed."));
    } finally {
      setBusy(false);
    }
  };

  const loginWithCookies = async () => {
    await loginWithCookiesInput(cookiesText, "Connected using pasted cookies.");
  };

  const pasteFromClipboardAndConnect = async () => {
    clearMessages();
    if (!navigator?.clipboard?.readText) {
      setError("Clipboard access is not available in this browser. Paste manually in the editor.");
      return;
    }

    try {
      const text = await navigator.clipboard.readText();
      if (!text || !text.trim()) {
        setError("Clipboard is empty.");
        return;
      }
      setCookiesText(text);
      setManualCookieEditorOpen(true);
      await loginWithCookiesInput(text, "Connected using cookies from clipboard.");
    } catch {
      setError("Failed to read clipboard. Paste manually in the editor.");
    }
  };

  const importCookieFileAndConnect = async (event) => {
    const selected = event.target.files?.[0];
    event.target.value = "";
    if (!selected) {
      return;
    }

    clearMessages();
    try {
      const text = await selected.text();
      setCookiesText(text);
      setManualCookieEditorOpen(true);
      await loginWithCookiesInput(text, `Connected using cookie file: ${selected.name}`);
    } catch {
      setError(`Failed to read cookie file: ${selected.name}`);
    }
  };

  const exportSavedCookies = async () => {
    clearMessages();
    setBusy(true);
    try {
      const response = await axios.get(`${apiBase}/testing/cis-bench/cookies/export`);
      const cookieText = response.data?.cookies_text || "";
      setCookiesText(cookieText);
      setManualCookieEditorOpen(true);
      setMessage(cookieText ? "Loaded saved cookies into editor." : "No saved cookie content found.");
    } catch (exportError) {
      setError(extractApiError(exportError, "Failed to load saved cookies."));
    } finally {
      setBusy(false);
    }
  };

  const copyCookiesToClipboard = async () => {
    clearMessages();
    if (!cookiesText.trim()) {
      setError("No cookie text available to copy.");
      return;
    }
    try {
      await navigator.clipboard.writeText(cookiesText);
      setMessage("Cookie text copied to clipboard.");
    } catch {
      setError("Clipboard copy failed.");
    }
  };

  const openWorkbenchLogin = () => {
    window.open("https://workbench.cisecurity.org/", "_blank", "noopener,noreferrer");
    setMessage("Opened CIS WorkBench login page in a new tab.");
    setError("");
  };

  const logout = async () => {
    clearMessages();
    setBusy(true);
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/logout`, {});
      setMessage(response.data?.message || "Logged out.");
      await loadStatus();
    } catch (logoutError) {
      setError(extractApiError(logoutError, "Logout failed."));
    } finally {
      setBusy(false);
    }
  };

  const refreshCatalog = async () => {
    clearMessages();
    setBusy(true);
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/catalog/refresh`, {
        browser: "chrome",
        no_verify_ssl: noVerifySSL,
      });
      setMessage(response.data?.message || "Catalog refresh complete.");
    } catch (refreshError) {
      setError(extractApiError(refreshError, "Catalog refresh failed."));
    } finally {
      setBusy(false);
    }
  };

  const runSearch = async () => {
    clearMessages();
    setBusy(true);
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/search`, searchReq);
      setSearchResults(response.data?.results || []);
      setMessage(`Found ${response.data?.count || 0} benchmark(s).`);
    } catch (searchError) {
      setError(extractApiError(searchError, "Search failed."));
      setSearchResults([]);
    } finally {
      setBusy(false);
    }
  };

  const downloadBenchmark = async (benchmarkID) => {
    clearMessages();
    const benchmark_id = String(benchmarkID || "").trim();
    if (!benchmark_id) {
      setError("Benchmark ID is required.");
      return;
    }

    setBusy(true);
    startDownloadProgress(`Downloading benchmark ${benchmark_id}`);
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/download`, {
        benchmark_id,
        formats: downloadFormats,
      });
      setMessage(response.data?.message || `Downloaded benchmark ${benchmark_id}.`);
      setDownloadBenchmarkId("");
      await loadFiles();
    } catch (downloadError) {
      setError(extractApiError(downloadError, "Download failed."));
    } finally {
      finishDownloadProgress();
      setBusy(false);
    }
  };

  const deleteDownload = async (name) => {
    clearMessages();
    if (!window.confirm(`Delete ${name}?`)) {
      return;
    }

    setBusy(true);
    try {
      const response = await axios.delete(`${apiBase}/testing/cis-bench/files/${encodeURIComponent(name)}`);
      setMessage(response.data?.message || `Deleted ${name}.`);
      await loadFiles();
    } catch (deleteError) {
      setError(extractApiError(deleteError, "Failed to delete file."));
    } finally {
      setBusy(false);
    }
  };

  const deleteAllDownloads = async () => {
    clearMessages();
    if (!window.confirm("Delete all downloaded benchmark files?")) {
      return;
    }

    setBusy(true);
    try {
      const response = await axios.delete(`${apiBase}/testing/cis-bench/files`, { params: { all: true } });
      setMessage(response.data?.message || "Deleted downloaded files.");
      await loadFiles();
    } catch (deleteError) {
      setError(extractApiError(deleteError, "Failed to delete downloaded files."));
    } finally {
      setBusy(false);
    }
  };

  const toggleDownloadFormat = (formatID) => {
    setDownloadFormats((current) => {
      if (current.includes(formatID)) {
        if (current.length === 1) {
          return current;
        }
        return current.filter((value) => value !== formatID);
      }
      return [...current, formatID];
    });
  };

  return (
    <Stack spacing={2}>
      <Typography variant="h6">CIS Bench (Testing)</Typography>
      {message && <Alert severity="success">{message}</Alert>}
      {error && <Alert severity="error">{error}</Alert>}

      <Paper sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Stack direction="row" justifyContent="space-between" alignItems="center">
            <Typography variant="subtitle1">Connect CIS WorkBench</Typography>
            <Chip label={status.logged_in ? "Connected" : "Not Connected"} color={status.logged_in ? "success" : "default"} size="small" />
          </Stack>
          <Typography variant="body2">
            1) Open and sign in to CIS WorkBench in your browser. 2) Import cookies from clipboard or file. 3) Session connects automatically.
          </Typography>
          <Typography variant="body2" color="warning.main">
            Browser profile extraction is disabled for Docker deployments. Import cookie text from your host browser instead.
          </Typography>

          <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
            <Button variant="outlined" onClick={openWorkbenchLogin} disabled={busy}>
              Open CIS WorkBench
            </Button>
            <Button variant="contained" onClick={pasteFromClipboardAndConnect} disabled={busy}>
              Paste Clipboard and Connect
            </Button>
            <Button variant="contained" component="label" disabled={busy}>
              Import Cookie File and Connect
              <input type="file" hidden accept=".txt,.json,.cookie,.cookies" onChange={importCookieFileAndConnect} />
            </Button>
          </Stack>

          <Button variant="text" onClick={() => setManualCookieEditorOpen((open) => !open)} sx={{ alignSelf: "flex-start" }}>
            {manualCookieEditorOpen ? "Hide Manual Cookie Editor" : "Show Manual Cookie Editor"}
          </Button>

          <Collapse in={manualCookieEditorOpen}>
            <Stack spacing={1.5}>
              <Typography variant="body2">
                Supported formats: Netscape cookie text, JSON cookie export, or raw `Cookie: name=value; ...` header.
              </Typography>
              <TextField
                label="Cookie Input"
                value={cookiesText}
                onChange={(event) => setCookiesText(event.target.value)}
                multiline
                minRows={6}
                fullWidth
              />
              <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
                <Button variant="contained" onClick={loginWithCookies} disabled={busy}>
                  Connect Using Editor Text
                </Button>
                <Button variant="outlined" onClick={exportSavedCookies} disabled={busy}>
                  Load Saved Cookies
                </Button>
                <Button variant="outlined" onClick={copyCookiesToClipboard} disabled={busy}>
                  Copy Editor Text
                </Button>
              </Stack>
              <FormControlLabel
                control={<Checkbox checked={noVerifySSL} onChange={(event) => setNoVerifySSL(event.target.checked)} />}
                label="Disable SSL verification (testing)"
              />
            </Stack>
          </Collapse>

          <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
            <Button variant="outlined" onClick={loadStatus} disabled={busy}>
              Refresh Status
            </Button>
            <Button variant="outlined" onClick={logout} disabled={busy}>
              Logout
            </Button>
          </Stack>
        </Stack>
      </Paper>

      <Paper sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Stack direction="row" justifyContent="space-between" alignItems="center">
            <Typography variant="subtitle1">Search Benchmarks</Typography>
            <Button variant="outlined" onClick={refreshCatalog} disabled={busy}>
              Refresh Catalog
            </Button>
          </Stack>
          <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
            <TextField
              label="Query"
              value={searchReq.query}
              onChange={(event) => setSearchReq((prev) => ({ ...prev, query: event.target.value }))}
              fullWidth
            />
            <TextField
              label="Platform Type"
              value={searchReq.platform_type}
              onChange={(event) => setSearchReq((prev) => ({ ...prev, platform_type: event.target.value }))}
              fullWidth
            />
            <TextField
              type="number"
              label="Limit"
              value={searchReq.limit}
              onChange={(event) => setSearchReq((prev) => ({ ...prev, limit: Number(event.target.value || 0) }))}
              sx={{ width: 140 }}
            />
          </Stack>
          <FormControlLabel
            control={
              <Checkbox
                checked={searchReq.latest}
                onChange={(event) => setSearchReq((prev) => ({ ...prev, latest: event.target.checked }))}
              />
            }
            label="Latest only"
          />
          <Stack spacing={0.5}>
            <Typography variant="body2">Download File Types</Typography>
            <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
              {downloadFormatOptions.map((option) => (
                <FormControlLabel
                  key={option.id}
                  control={
                    <Checkbox
                      checked={downloadFormats.includes(option.id)}
                      onChange={() => toggleDownloadFormat(option.id)}
                    />
                  }
                  label={option.label}
                />
              ))}
            </Stack>
          </Stack>
          <Button variant="contained" onClick={runSearch} disabled={busy}>
            Search
          </Button>

          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>Title</TableCell>
                <TableCell>Version</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Action</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {searchResults.map((row) => (
                <TableRow key={String(row.benchmark_id)}>
                  <TableCell>{row.benchmark_id}</TableCell>
                  <TableCell>{row.title}</TableCell>
                  <TableCell>{row.version || ""}</TableCell>
                  <TableCell>{row.status || ""}</TableCell>
                  <TableCell>
                    <Button size="small" variant="outlined" onClick={() => downloadBenchmark(row.benchmark_id)} disabled={busy}>
                      Download
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Stack>
      </Paper>

      <Paper sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Typography variant="subtitle1">Downloaded Files</Typography>
          {downloadProgress.active && (
            <Stack spacing={0.5}>
              <Typography variant="body2">
                {downloadProgress.phase === "finalizing" ? `${downloadProgress.label} (finalizing...)` : downloadProgress.label}
              </Typography>
              <LinearProgress
                variant={downloadProgress.phase === "finalizing" ? "indeterminate" : "determinate"}
                value={downloadProgress.value}
              />
            </Stack>
          )}
          <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
            <TextField
              label="Benchmark ID"
              value={downloadBenchmarkId}
              onChange={(event) => setDownloadBenchmarkId(event.target.value)}
              sx={{ width: 220 }}
            />
            <Button variant="contained" onClick={() => downloadBenchmark(downloadBenchmarkId)} disabled={busy}>
              Download by ID
            </Button>
            <Button variant="outlined" onClick={loadFiles} disabled={busy}>
              Refresh Files
            </Button>
            <Button variant="outlined" color="error" onClick={deleteAllDownloads} disabled={busy || files.length === 0}>
              Delete All
            </Button>
          </Stack>

          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Size</TableCell>
                <TableCell>Modified</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {files.map((file) => (
                <TableRow key={file.name}>
                  <TableCell>{file.name}</TableCell>
                  <TableCell>{formatFileSize(file.size)}</TableCell>
                  <TableCell>{new Date(file.modified_at).toLocaleString()}</TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={1}>
                      <Button
                        size="small"
                        variant="outlined"
                        component="a"
                        href={`${apiBase}/testing/cis-bench/files/${encodeURIComponent(file.name)}/download`}
                      >
                        Download
                      </Button>
                      <Button size="small" variant="outlined" color="error" onClick={() => deleteDownload(file.name)} disabled={busy}>
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
    </Stack>
  );
}
