import React, { useMemo, useState } from "react";
import { Alert, AppBar, Box, Container, CssBaseline, Tab, Tabs, Toolbar, Typography } from "@mui/material";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import Settings from "./pages/Settings";
import ReportsHub from "./pages/ReportsHub";
import GPOWorkflow from "./pages/GPOWorkflow";
import BenchmarkWorkflow from "./pages/BenchmarkWorkflow";

function resolveApiBase() {
  const configured = String(import.meta.env.VITE_API_BASE_URL || "").trim();
  if (configured) {
    return configured.replace(/\/+$/, "");
  }
  if (typeof window !== "undefined" && window?.location?.hostname) {
    return `${window.location.protocol}//${window.location.hostname}:8080`;
  }
  return "http://localhost:8080";
}

const API_BASE = resolveApiBase();
const appTheme = createTheme({
  palette: {
    mode: "dark",
  },
});

export default function App() {
  const [tab, setTab] = useState(0);
  const [reportId, setReportId] = useState("");
  const mixedProtocol = typeof window !== "undefined" && window.location.protocol === "https:" && API_BASE.startsWith("http://");

  const views = useMemo(
    () => [
      <BenchmarkWorkflow key="benchmark-workflow" apiBase={API_BASE} onReportCreated={setReportId} />,
      <GPOWorkflow key="gpo-workflow" apiBase={API_BASE} />,
      <ReportsHub key="reports" apiBase={API_BASE} reportId={reportId} onReportIdChange={setReportId} />,
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
        {mixedProtocol && (
          <Alert severity="warning" sx={{ mb: 2 }}>
            API base is HTTP while the UI is HTTPS. Use http://localhost or configure VITE_API_BASE_URL to an HTTPS API endpoint.
          </Alert>
        )}
        <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 2 }}>
          <Tabs value={tab} onChange={(_, value) => setTab(value)}>
            <Tab label="Benchmark Workflow" />
            <Tab label="GPO Workflow" />
            <Tab label="Reports" />
            <Tab label="Settings" />
          </Tabs>
        </Box>
        {views[tab]}
      </Container>
    </ThemeProvider>
  );
}
