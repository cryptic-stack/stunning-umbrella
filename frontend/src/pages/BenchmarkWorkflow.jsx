import React from "react";
import { Stack } from "@mui/material";
import UploadBenchmarks from "./UploadBenchmarks";
import VersionComparison from "./VersionComparison";
import CISBench from "./CISBench";

export default function BenchmarkWorkflow({ apiBase, onReportCreated }) {
  return (
    <Stack spacing={2}>
      <UploadBenchmarks apiBase={apiBase} />
      <VersionComparison apiBase={apiBase} onReportCreated={onReportCreated} />
      <CISBench apiBase={apiBase} />
    </Stack>
  );
}
