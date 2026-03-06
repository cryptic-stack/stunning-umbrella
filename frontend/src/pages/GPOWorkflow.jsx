import React, { useState } from "react";
import { Stack } from "@mui/material";
import GPOImport from "./GPOImport";
import GPOAssessment from "./GPOAssessment";

export default function GPOWorkflow({ apiBase }) {
  const [benchmarkContext, setBenchmarkContext] = useState(null);

  return (
    <Stack spacing={2}>
      <GPOImport apiBase={apiBase} onBenchmarkContextChange={setBenchmarkContext} />
      <GPOAssessment apiBase={apiBase} benchmarkContext={benchmarkContext} />
    </Stack>
  );
}
