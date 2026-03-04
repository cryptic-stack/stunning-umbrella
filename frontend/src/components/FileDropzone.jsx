import React, { useRef } from "react";
import { Box, Paper, Typography } from "@mui/material";

export default function FileDropzone({ onFilesSelected, accepted, disabled = false }) {
  const inputRef = useRef(null);

  const handleDrop = (event) => {
    event.preventDefault();
    if (disabled) {
      return;
    }
    if (!event.dataTransfer.files?.length) {
      return;
    }
    onFilesSelected(Array.from(event.dataTransfer.files));
  };

  return (
    <Paper
      onDragOver={(event) => event.preventDefault()}
      onDrop={handleDrop}
      onClick={() => {
        if (!disabled) {
          inputRef.current?.click();
        }
      }}
      sx={{
        p: 3,
        border: "2px dashed #4f8a8b",
        textAlign: "center",
        cursor: disabled ? "not-allowed" : "pointer",
        opacity: disabled ? 0.7 : 1,
      }}
    >
      <Typography variant="body1">
        Drag and drop benchmark files, or click to browse. Files upload automatically after selection.
      </Typography>
      <Box
        component="input"
        type="file"
        accept={accepted}
        ref={inputRef}
        hidden
        multiple
        onChange={(event) => {
          const files = Array.from(event.target.files || []);
          if (files.length > 0) {
            onFilesSelected(files);
          }
          event.target.value = "";
        }}
      />
    </Paper>
  );
}
