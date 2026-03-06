import React from "react";
import { Stack } from "@mui/material";
import GPOImport from "./GPOImport";
import GPOAssessment from "./GPOAssessment";

export default function GPOWorkflow({ apiBase }) {
  return (
    <Stack spacing={2}>
      <GPOImport apiBase={apiBase} />
      <GPOAssessment apiBase={apiBase} />
    </Stack>
  );
}

