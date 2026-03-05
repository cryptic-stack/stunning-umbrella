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
  const browserExtractionSupported = false;
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

  const clearMessages = () => {
    setMessage("");
    setError("");
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
      setMessage(response.data?.message || "Generated session cookie from browser.");
      await loadStatus();
    } catch (loginError) {
      setError(
        extractApiError(
          loginError,
          "Failed to generate cookie session from browser. In Docker, use exported/pasted cookies from your host browser."
        )
      );
    } finally {
      setBusy(false);
    }
  };

  const exportSavedCookies = async () => {
    clearMessages();
    setBusy(true);
    try {
      const response = await axios.get(`${apiBase}/testing/cis-bench/cookies/export`);
      const cookieText = response.data?.cookies_text || "";
      setCookiesText(cookieText);
      setMessage(cookieText ? "Exported saved session cookies into the editor." : "No cookie content found.");
    } catch (exportError) {
      setError(exportError?.response?.data?.error || "Failed to export saved cookies.");
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
      await loadStatus();
    } catch (loginError) {
      setError(extractApiError(loginError, "cis-bench login failed."));
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
        browser,
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
    try {
      const response = await axios.post(`${apiBase}/testing/cis-bench/download`, {
        benchmark_id,
        formats: ["json"],
      });
      setMessage(response.data?.message || `Downloaded benchmark ${benchmark_id}.`);
      setDownloadBenchmarkId("");
      await loadFiles();
    } catch (downloadError) {
      setError(extractApiError(downloadError, "Download failed."));
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
            Workflow: 1) Open CIS login page, 2) sign in, 3) export cookies from your browser, 4) paste them here and click Use Pasted Cookies.
          </Typography>
          <Typography variant="body2" color="warning.main">
            Browser cookie extraction from inside the API container is not supported in this deployment. Use exported cookies from your host browser.
          </Typography>
          <Typography variant="body2">
            Supported cookie input formats: Netscape cookie text, JSON cookie export, or raw "Cookie: name=value; ..." header.
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
            <Button variant="outlined" onClick={openWorkbenchLogin} disabled={busy}>
              Open CIS Login Page
            </Button>
            <Button
              variant="contained"
              onClick={loginWithBrowser}
              disabled={busy || !browserExtractionSupported}
              title="Disabled: API runs in container and cannot read host browser profile/cookies."
            >
              Generate Cookie Session (Unsupported Here)
            </Button>
            <Button variant="outlined" onClick={exportSavedCookies} disabled={busy}>
              Export Saved Cookies
            </Button>
            <Button variant="outlined" onClick={copyCookiesToClipboard} disabled={busy}>
              Copy Cookies
            </Button>
          </Stack>
          <Stack direction="row" spacing={1}>
            <Button variant="contained" onClick={loginWithCookies} disabled={busy}>
              Use Pasted Cookies
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
