import React, { useEffect, useState } from "react";
import axios from "axios";
import { Alert, Button, FormControl, InputLabel, Link, MenuItem, Paper, Select, Stack, Table, TableBody, TableCell, TableHead, TableRow, Typography } from "@mui/material";

const exportFormats = ["json", "md", "html", "csv", "xlsx", "docx"];

export default function GPOReports({ apiBase }) {
  const [rows, setRows] = useState([]);
  const [selectedAssessmentId, setSelectedAssessmentId] = useState("");
  const [result, setResult] = useState(null);
  const [error, setError] = useState("");

  const loadAssessments = async () => {
    setError("");
    try {
      const response = await axios.get(`${apiBase}/api/gpo/assessments`);
      const list = response.data || [];
      setRows(list);
      if (!selectedAssessmentId && list.length > 0) {
        setSelectedAssessmentId(String(list[0].id));
      }
    } catch (err) {
      setError(err?.response?.data?.error || "Failed to load assessments");
      setRows([]);
    }
  };

  const loadAssessment = async () => {
    if (!selectedAssessmentId) {
      return;
    }
    setError("");
    try {
      const response = await axios.get(`${apiBase}/api/gpo/assessments/${selectedAssessmentId}`);
      setResult(response.data);
    } catch (err) {
      setError(err?.response?.data?.error || "Failed to load assessment details");
      setResult(null);
    }
  };

  useEffect(() => {
    loadAssessments();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Step 4: Review + Export</Typography>
        <Stack direction="row" spacing={1}>
          <Button variant="outlined" onClick={loadAssessments}>Refresh</Button>
        </Stack>

        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>ID</TableCell>
              <TableCell>Policy Source</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Created</TableCell>
              <TableCell>Exports</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {rows.map((row) => (
              <TableRow key={row.id}>
                <TableCell>{row.id}</TableCell>
                <TableCell>{row.policy_source_id}</TableCell>
                <TableCell>{row.status}</TableCell>
                <TableCell>{new Date(row.created_at).toLocaleString()}</TableCell>
                <TableCell>
                  <Stack direction="row" spacing={1}>
                    {exportFormats.map((format) => (
                      <Link key={format} href={`${apiBase}/api/gpo/assessments/${row.id}/report/${format}`} target="_blank" rel="noreferrer">
                        {format.toUpperCase()}
                      </Link>
                    ))}
                  </Stack>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>

        <FormControl fullWidth>
          <InputLabel id="assessment-select-label">Assessment</InputLabel>
          <Select labelId="assessment-select-label" label="Assessment" value={selectedAssessmentId} onChange={(event) => setSelectedAssessmentId(event.target.value)}>
            {rows.map((row) => (
              <MenuItem key={row.id} value={String(row.id)}>
                #{row.id} - {row.status}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <Button variant="contained" onClick={loadAssessment}>Load Assessment Details</Button>

        {result && (
          <Alert severity="info">
            Assessment #{result.assessment?.id} has {(result.results || []).length} result item(s).
          </Alert>
        )}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </Paper>
  );
}

