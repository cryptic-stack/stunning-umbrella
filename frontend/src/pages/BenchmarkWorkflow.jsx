import React from "react";
import { Stack } from "@mui/material";
import UploadBenchmarks from "./UploadBenchmarks";
import VersionComparison from "./VersionComparison";

export default function BenchmarkWorkflow({ apiBase, onReportCreated }) {
  return (
    <Stack spacing={2}>
      <UploadBenchmarks apiBase={apiBase} />
      <VersionComparison apiBase={apiBase} onReportCreated={onReportCreated} />
    </Stack>
  );
}

