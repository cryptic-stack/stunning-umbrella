import React, { useMemo, useState } from "react";
import { Box, Tab, Tabs } from "@mui/material";
import DiffViewer from "./DiffViewer";
import GPOReports from "./GPOReports";

export default function ReportsHub({ apiBase, reportId, onReportIdChange }) {
  const [tab, setTab] = useState(0);

  const views = useMemo(
    () => [
      <DiffViewer key="diff" apiBase={apiBase} reportId={reportId} onReportIdChange={onReportIdChange} />,
      <GPOReports key="gpo-reports" apiBase={apiBase} />,
    ],
    [apiBase, onReportIdChange, reportId]
  );

  return (
    <>
      <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 2 }}>
        <Tabs value={tab} onChange={(_, value) => setTab(value)}>
          <Tab label="Diff Viewer" />
          <Tab label="GPO Reports" />
        </Tabs>
      </Box>
      {views[tab]}
    </>
  );
}

