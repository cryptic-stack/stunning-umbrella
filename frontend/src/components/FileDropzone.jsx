import React, { useRef } from "react";
import { Box, Paper, Typography } from "@mui/material";

export default function FileDropzone({ onFileSelected, accepted }) {
  const inputRef = useRef(null);

  const handleDrop = (event) => {
    event.preventDefault();
    if (!event.dataTransfer.files?.length) {
      return;
    }
    onFileSelected(event.dataTransfer.files[0]);
  };

  return (
    <Paper
      onDragOver={(event) => event.preventDefault()}
      onDrop={handleDrop}
      onClick={() => inputRef.current?.click()}
      sx={{
        p: 3,
        border: "2px dashed #4f8a8b",
        textAlign: "center",
        cursor: "pointer",
      }}
    >
      <Typography variant="body1">Drag and drop a benchmark file, or click to browse.</Typography>
      <Box component="input" type="file" accept={accepted} ref={inputRef} hidden onChange={(event) => onFileSelected(event.target.files[0])} />
    </Paper>
  );
}
