import React, { useMemo } from "react";
import { createPatch } from "diff";
import { Diff, Hunk, parseDiff } from "react-diff-view";
import { Box, Paper, Stack, Typography } from "@mui/material";

export default function DiffPane({
  item,
  oldVersionLabel = "Old",
  newVersionLabel = "New",
  swapSides = false,
}) {
  const originalText = swapSides ? item?.new_text || "" : item?.old_text || "";
  const modifiedText = swapSides ? item?.old_text || "" : item?.new_text || "";

  const { files } = useMemo(() => {
    if (!item) {
      return { files: [] };
    }

    const sanitizeFiles = (input) =>
      (input || [])
        .map((file) => ({
          ...file,
          hunks: (file?.hunks || []).filter((hunk) => Array.isArray(hunk?.changes) && hunk.changes.length > 0),
        }))
        .filter((file) => file.hunks.length > 0);

    try {
      const patch = createPatch(
        item.safeguard_new || item.safeguard_old || "safeguard",
        originalText,
        modifiedText,
        oldVersionLabel,
        newVersionLabel
      );

      // Primary parse mode can fail on some malformed hunks; retry with default parser options.
      try {
        const parsed = sanitizeFiles(parseDiff(patch, { nearbySequences: "zip" }));
        if (parsed.length > 0) {
          return { files: parsed };
        }
      } catch {
        // Fall through to default parsing mode.
      }

      try {
        const parsed = sanitizeFiles(parseDiff(patch));
        if (parsed.length > 0) {
          return { files: parsed };
        }
      } catch {
        // Keep Monaco diff as fallback below.
      }

      return { files: [] };
    } catch {
      return { files: [] };
    }
  }, [item, originalText, modifiedText, oldVersionLabel, newVersionLabel]);

  if (!item) {
    return <Typography variant="body2">Select a diff item to inspect details.</Typography>;
  }

  return (
    <Box sx={{ flex: 1, minWidth: 0 }}>
      <Typography variant="subtitle1" sx={{ mb: 1 }}>
        {item.change_type.toUpperCase()} {item.safeguard_new || item.safeguard_old}
      </Typography>
      <Stack direction="row" spacing={2} sx={{ mb: 1 }}>
        <Typography variant="body2">
          Old: <strong>{oldVersionLabel}</strong>
        </Typography>
        <Typography variant="body2">
          New: <strong>{newVersionLabel}</strong>
        </Typography>
      </Stack>
      <Box sx={{ mt: 2, display: "grid", gap: 1.5, gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" } }}>
        <Paper variant="outlined" sx={{ p: 1.5, maxHeight: 220, overflow: "auto" }}>
          <Typography variant="caption" color="text.secondary">
            Old Text ({oldVersionLabel})
          </Typography>
          <Typography variant="body2" sx={{ whiteSpace: "pre-wrap", mt: 0.5 }}>
            {originalText || "No old text for this change."}
          </Typography>
        </Paper>
        <Paper variant="outlined" sx={{ p: 1.5, maxHeight: 220, overflow: "auto" }}>
          <Typography variant="caption" color="text.secondary">
            New Text ({newVersionLabel})
          </Typography>
          <Typography variant="body2" sx={{ whiteSpace: "pre-wrap", mt: 0.5 }}>
            {modifiedText || "No new text for this change."}
          </Typography>
        </Paper>
      </Box>
      {files.length > 0 && (
        <Box sx={{ mt: 2, maxHeight: 420, overflow: "auto", border: "1px solid #ddd", borderRadius: 1 }}>
          {files.map((file, fileIndex) => (
            <Diff key={`${file.oldRevision || "old"}-${file.newRevision || "new"}-${fileIndex}`} viewType="split" diffType={file.type} hunks={file.hunks}>
              {(hunks) => hunks.map((hunk, hunkIndex) => <Hunk key={`${fileIndex}-${hunkIndex}`} hunk={hunk} />)}
            </Diff>
          ))}
        </Box>
      )}
    </Box>
  );
}
