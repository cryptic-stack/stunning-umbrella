import React, { useEffect, useState } from "react";
import axios from "axios";
import {
  Alert,
  Button,
  Checkbox,
  FormControlLabel,
  MenuItem,
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

export default function TestingCISBench({ apiBase }) {
  const [status, setStatus] = useState({ logged_in: false });
  const [cookiesText, setCookiesText] = useState("");
  const [browser, setBrowser] = useState("chrome");
  const [noVerifySSL, setNoVerifySSL] = useState(true);
  const [searchReq, setSearchReq] = useState(defaultSearch);
  const [searchResults, setSearchResults] = useState([]);
  const [downloadBenchmarkId, setDownloadBenchmarkId] = useState("");
  const [files, setFiles] = useState([]);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

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

  const clearMessages = () => {
    setMessage("");
    setError("");
  };

  const loginWithCookies = async () => {
    clearMessages();
    if (!cookiesText.trim()) {
      setError("Paste Netscape cookie content first.");
      return;
    }

    setBusy(true);
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/login`, {
        mode: "cookies",
        cookies_text: cookiesText,
        no_verify_ssl: noVerifySSL,
      });
      setMessage(response.data?.message || "Logged in to cis-bench.");
      setCookiesText("");
      await loadStatus();
    } catch (loginError) {
      setError(loginError?.response?.data?.error || "cis-bench login failed.");
    } finally {
      setBusy(false);
    }
  };

  const loginWithBrowser = async () => {
    clearMessages();
    setBusy(true);
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/login`, {
        mode: "browser",
        browser,
        no_verify_ssl: noVerifySSL,
      });
      setMessage(response.data?.message || "Attempted browser login.");
      await loadStatus();
    } catch (loginError) {
      setError(loginError?.response?.data?.error || "Browser login failed.");
    } finally {
      setBusy(false);
    }
  };

  const logout = async () => {
    clearMessages();
    setBusy(true);
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/logout`, {});
      setMessage(response.data?.message || "Logged out.");
      await loadStatus();
    } catch (logoutError) {
      setError(logoutError?.response?.data?.error || "Logout failed.");
    } finally {
      setBusy(false);
    }
  };

  const refreshCatalog = async () => {
    clearMessages();
    setBusy(true);
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/catalog/refresh`, {
        browser,
        no_verify_ssl: noVerifySSL,
      });
      setMessage(response.data?.message || "Catalog refresh complete.");
    } catch (refreshError) {
      setError(refreshError?.response?.data?.error || "Catalog refresh failed.");
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
      setError(searchError?.response?.data?.error || "Search failed.");
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
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/download`, {
        benchmark_id,
        formats: ["json"],
      });
      setMessage(response.data?.message || `Downloaded benchmark ${benchmark_id}.`);
      setDownloadBenchmarkId("");
      await loadFiles();
    } catch (downloadError) {
      setError(downloadError?.response?.data?.error || "Download failed.");
    } finally {
      setBusy(false);
    }
  };

  return (
    <Stack spacing={2}>
      <Typography variant="h6">CIS Bench (Testing)</Typography>
      {message && <Alert severity="success">{message}</Alert>}
      {error && <Alert severity="error">{error}</Alert>}

      <Paper sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Typography variant="subtitle1">Login</Typography>
          <Typography variant="body2">
            Use cookie login for containerized environments. Browser login may fail if no browser profile exists in the API container.
          </Typography>
          <FormControlLabel
            control={<Checkbox checked={noVerifySSL} onChange={(event) => setNoVerifySSL(event.target.checked)} />}
            label="Disable SSL verification (testing)"
          />
          <TextField
            select
            label="Browser (optional)"
            value={browser}
            onChange={(event) => setBrowser(event.target.value)}
            fullWidth
          >
            <MenuItem value="chrome">Chrome</MenuItem>
            <MenuItem value="firefox">Firefox</MenuItem>
            <MenuItem value="edge">Edge</MenuItem>
            <MenuItem value="safari">Safari</MenuItem>
          </TextField>
          <TextField
            label="Cookies Text (Netscape format)"
            value={cookiesText}
            onChange={(event) => setCookiesText(event.target.value)}
            multiline
            minRows={5}
            fullWidth
          />
          <Stack direction="row" spacing={1}>
            <Button variant="contained" onClick={loginWithCookies} disabled={busy}>
              Login with Cookies
            </Button>
            <Button variant="outlined" onClick={loginWithBrowser} disabled={busy}>
              Login with Browser
            </Button>
            <Button variant="outlined" onClick={logout} disabled={busy}>
              Logout
            </Button>
            <Button variant="outlined" onClick={loadStatus} disabled={busy}>
              Check Status
            </Button>
          </Stack>
          <Typography variant="body2">
            Logged in: <strong>{status.logged_in ? "Yes" : "No"}</strong>
          </Typography>
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
          </Stack>

          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Size</TableCell>
                <TableCell>Modified</TableCell>
                <TableCell>Download</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {files.map((file) => (
                <TableRow key={file.name}>
                  <TableCell>{file.name}</TableCell>
                  <TableCell>{file.size}</TableCell>
                  <TableCell>{new Date(file.modified_at).toLocaleString()}</TableCell>
                  <TableCell>
                    <Button
                      size="small"
                      variant="outlined"
                      component="a"
                      href={`${apiBase}/testing/cis-bench/files/${encodeURIComponent(file.name)}/download`}
                    >
                      Download
                    </Button>
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
