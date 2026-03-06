import React, { useEffect, useState } from "react";
import axios from "axios";
import { Alert, Button, Link, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography } from "@mui/material";

const exportFormats = ["json", "md", "html", "csv", "xlsx", "docx"];

export default function GPOReports({ apiBase }) {
  const [rows, setRows] = useState([]);
  const [assessmentId, setAssessmentId] = useState("");
  const [result, setResult] = useState(null);
  const [error, setError] = useState("");

  const loadAssessments = async () => {
    setError("");
    try {
      const response = await axios.get(`${apiBase}/api/gpo/assessments`);
      setRows(response.data || []);
    } catch (err) {
      setError(err?.response?.data?.error || "Failed to load assessments");
      setRows([]);
    }
  };

  const loadAssessment = async () => {
    if (!assessmentId) {
      return;
    }
    setError("");
    try {
      const response = await axios.get(`${apiBase}/api/gpo/assessments/${assessmentId}`);
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
        <Typography variant="h6">GPO Assessments</Typography>
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

        <Typography variant="h6" sx={{ pt: 2 }}>Assessment Details</Typography>
        <Stack direction="row" spacing={1}>
          <TextField label="Assessment ID" value={assessmentId} onChange={(event) => setAssessmentId(event.target.value)} />
          <Button variant="contained" onClick={loadAssessment}>Load</Button>
        </Stack>
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

