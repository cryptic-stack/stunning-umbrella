import React, { useState } from "react";
import { Stack } from "@mui/material";
import GPOImport from "./GPOImport";
import GPOAssessment from "./GPOAssessment";

export default function GPOWorkflow({ apiBase, onOpenReports }) {
  const [benchmarkContext, setBenchmarkContext] = useState(null);
  const [refreshToken, setRefreshToken] = useState(0);

  return (
    <Stack spacing={2}>
      <GPOImport
        apiBase={apiBase}
        onBenchmarkContextChange={setBenchmarkContext}
        onPolicyImportQueued={() => setRefreshToken((value) => value + 1)}
      />
      <GPOAssessment apiBase={apiBase} benchmarkContext={benchmarkContext} refreshToken={refreshToken} onOpenReports={onOpenReports} />
    </Stack>
  );
}
