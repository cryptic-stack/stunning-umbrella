import React, { useMemo } from "react";
import { createPatch } from "diff";
import DiffEditor from "@monaco-editor/react";
import { Diff, Hunk, parseDiff } from "react-diff-view";
import { Box, Typography } from "@mui/material";

export default function DiffPane({ item }) {
  const files = useMemo(() => {
    if (!item) {
      return [];
    }
    const patch = createPatch(
      item.safeguard_new || item.safeguard_old || "safeguard",
      item.old_text || "",
      item.new_text || "",
      "OLD",
      "NEW"
    );
    return parseDiff(patch, { nearbySequences: "zip" });
  }, [item]);

  if (!item) {
    return <Typography variant="body2">Select a diff item to inspect details.</Typography>;
  }

  return (
    <Box sx={{ flex: 1, minWidth: 0 }}>
      <Typography variant="subtitle1" sx={{ mb: 1 }}>
        {item.change_type.toUpperCase()} {item.safeguard_new || item.safeguard_old}
      </Typography>
      <DiffEditor
        height="220px"
        language="markdown"
        original={item.old_text || ""}
        modified={item.new_text || ""}
        options={{ renderSideBySide: true, readOnly: true, minimap: { enabled: false } }}
      />
      <Box sx={{ mt: 2, maxHeight: 260, overflow: "auto", border: "1px solid #ddd", borderRadius: 1 }}>
        {files.map((file) => (
          <Diff key={file.oldRevision + file.newRevision} viewType="split" diffType={file.type} hunks={file.hunks}>
            {(hunks) => hunks.map((hunk) => <Hunk key={hunk.content} hunk={hunk} />)}
          </Diff>
        ))}
      </Box>
    </Box>
  );
}
