import React, { useEffect, useMemo, useState } from "react";
import axios from "axios";
import { Alert, Button, Chip, Divider, List, ListItemButton, Paper, Stack, TextField, Typography } from "@mui/material";
import DiffPane from "../components/DiffPane";

export default function DiffViewer({ apiBase, reportId, onReportIdChange }) {
  const [report, setReport] = useState(null);
  const [items, setItems] = useState([]);
  const [selectedItem, setSelectedItem] = useState(null);
  const [error, setError] = useState("");

  const loadReport = async () => {
    if (!reportId) {
      return;
    }

    setError("");
    try {
      const response = await axios.get(`${apiBase}/diff/${reportId}`);
      setReport(response.data.report);
      setItems(response.data.items || []);
      setSelectedItem((response.data.items || [])[0] || null);
    } catch (diffError) {
      setError(diffError?.response?.data?.error || "Failed to load diff report.");
    }
  };

  useEffect(() => {
    loadReport();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [reportId]);

  const summary = useMemo(() => {
    const map = { added: 0, removed: 0, modified: 0, renamed: 0 };
    for (const item of items) {
      map[item.change_type] = (map[item.change_type] || 0) + 1;
    }
    return map;
  }, [items]);

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Diff Viewer</Typography>
        <Stack direction="row" spacing={2}>
          <TextField
            label="Report ID"
            value={reportId}
            onChange={(event) => onReportIdChange(event.target.value)}
            sx={{ maxWidth: 220 }}
          />
          <Button variant="contained" onClick={loadReport}>
            Load Report
          </Button>
        </Stack>

        {report && (
          <Stack direction="row" spacing={1}>
            <Chip label={`Status: ${report.status}`} />
            <Chip label={`Added: ${summary.added || 0}`} />
            <Chip label={`Removed: ${summary.removed || 0}`} />
            <Chip label={`Modified: ${summary.modified || 0}`} />
            <Chip label={`Renamed: ${summary.renamed || 0}`} />
          </Stack>
        )}

        {error && <Alert severity="error">{error}</Alert>}

        <Stack direction={{ xs: "column", md: "row" }} spacing={2} divider={<Divider flexItem orientation="vertical" />}>
          <List sx={{ minWidth: 280, maxHeight: 420, overflow: "auto" }}>
            {items.map((item) => (
              <ListItemButton key={item.id} onClick={() => setSelectedItem(item)} selected={selectedItem?.id === item.id}>
                {item.change_type.toUpperCase()} {item.safeguard_new || item.safeguard_old}
              </ListItemButton>
            ))}
          </List>

          <DiffPane item={selectedItem} />
        </Stack>
      </Stack>
    </Paper>
  );
}
