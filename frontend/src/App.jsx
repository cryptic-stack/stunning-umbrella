import React, { useMemo, useState } from "react";
import { AppBar, Box, Container, CssBaseline, Tab, Tabs, Toolbar, Typography } from "@mui/material";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import UploadBenchmarks from "./pages/UploadBenchmarks";
import VersionComparison from "./pages/VersionComparison";
import DiffViewer from "./pages/DiffViewer";
import Settings from "./pages/Settings";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";
const appTheme = createTheme({
  palette: {
    mode: "dark",
  },
});

export default function App() {
  const [tab, setTab] = useState(0);
  const [reportId, setReportId] = useState("");

  const views = useMemo(
    () => [
      <UploadBenchmarks key="upload" apiBase={API_BASE} />,
      <VersionComparison key="compare" apiBase={API_BASE} onReportCreated={setReportId} />,
      <DiffViewer key="diff" apiBase={API_BASE} reportId={reportId} onReportIdChange={setReportId} />,
      <Settings key="settings" apiBase={API_BASE} />,
    ],
    [reportId]
  );

  return (
    <ThemeProvider theme={appTheme}>
      <CssBaseline />
      <AppBar position="static" color="default">
        <Toolbar>
          <Typography variant="h6" sx={{ flexGrow: 1 }}>
            CIS Benchmark Intelligence
          </Typography>
        </Toolbar>
      </AppBar>
      <Container maxWidth="lg" sx={{ py: 3 }}>
        <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 2 }}>
          <Tabs value={tab} onChange={(_, value) => setTab(value)}>
            <Tab label="Upload Benchmarks" />
            <Tab label="Version Comparison" />
            <Tab label="Diff Viewer" />
            <Tab label="Settings" />
          </Tabs>
        </Box>
        {views[tab]}
      </Container>
    </ThemeProvider>
  );
}
