import React, { useEffect, useMemo, useState } from "react";
import axios from "axios";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import {
  Alert,
  Box,
  Button,
  Checkbox,
  Chip,
  FormControl,
  FormControlLabel,
  InputLabel,
  LinearProgress,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  MenuItem,
  Paper,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import DiffPane from "../components/DiffPane";

const CHANGE_TYPES = ["added", "removed", "modified", "renamed"];

function compactText(value, limit = 120) {
  const normalized = String(value || "")
    .replace(/\s+/g, " ")
    .trim();
  if (normalized.length <= limit) {
    return normalized;
  }
  return `${normalized.slice(0, limit - 1)}...`;
}

export default function DiffViewer({ apiBase, reportId, onReportIdChange }) {
  const [report, setReport] = useState(null);
  const [reportName, setReportName] = useState("");
  const [items, setItems] = useState([]);
  const [selectedItem, setSelectedItem] = useState(null);
  const [reports, setReports] = useState([]);
  const [versionLabels, setVersionLabels] = useState({ a: "", b: "" });
  const [reviewFilter, setReviewFilter] = useState("all");
  const [typeFilter, setTypeFilter] = useState("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [reviewDraft, setReviewDraft] = useState({ reviewed: false, review_comment: "" });
  const [savingReview, setSavingReview] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const compareVersionLabels = (left, right) => {
    const a = String(left || "").trim();
    const b = String(right || "").trim();
    if (a === b) {
      return 0;
    }

    const parse = (value) => value.split(".").map((part) => Number.parseInt(part, 10));
    const aParts = parse(a);
    const bParts = parse(b);
    const numericA = aParts.every((n) => Number.isFinite(n));
    const numericB = bParts.every((n) => Number.isFinite(n));

    if (numericA && numericB) {
      const length = Math.max(aParts.length, bParts.length);
      for (let index = 0; index < length; index += 1) {
        const av = aParts[index] ?? 0;
        const bv = bParts[index] ?? 0;
        if (av < bv) {
          return -1;
        }
        if (av > bv) {
          return 1;
        }
      }
      return 0;
    }

    return a.localeCompare(b, undefined, { numeric: true, sensitivity: "base" });
  };

  const loadReports = async () => {
    try {
      const response = await axios.get(`${apiBase}/reports`);
      setReports(response.data || []);
    } catch {
      setReports([]);
    }
  };

  const loadReport = async (id = reportId) => {
    if (!id) {
      return;
    }

    setError("");
    try {
      const response = await axios.get(`${apiBase}/diff/${id}`);
      const loadedItems = response.data.items || [];
      setReport(response.data.report);
      setReportName(response.data.report_name || "");
      setItems(loadedItems);
      setVersionLabels({
        a: response.data.version_a_label || "",
        b: response.data.version_b_label || "",
      });
      setSelectedItem((current) => loadedItems.find((item) => item.id === current?.id) || loadedItems[0] || null);
    } catch (diffError) {
      setError(diffError?.response?.data?.error || "Failed to load diff report.");
    }
  };

  const applyUpdatedItem = (updatedItem) => {
    setItems((previous) => previous.map((item) => (item.id === updatedItem.id ? { ...item, ...updatedItem } : item)));
    setSelectedItem((previous) => (previous && previous.id === updatedItem.id ? { ...previous, ...updatedItem } : previous));
  };

  const saveItemReview = async (itemId, payload, showSuccess = true) => {
    try {
      const response = await axios.patch(`${apiBase}/diff/items/${itemId}/review`, payload);
      applyUpdatedItem(response.data.item);
      if (showSuccess) {
        setMessage("Review notes saved.");
      }
      return response.data.item;
    } catch (reviewError) {
      setError(reviewError?.response?.data?.error || "Failed to save review.");
      return null;
    }
  };

  const saveSelectedReview = async () => {
    if (!selectedItem) {
      return;
    }
    setError("");
    setMessage("");
    setSavingReview(true);
    await saveItemReview(
      selectedItem.id,
      { reviewed: reviewDraft.reviewed, review_comment: reviewDraft.review_comment },
      true
    );
    setSavingReview(false);
  };

  const toggleReviewedInline = async (event, item) => {
    event.stopPropagation();
    setError("");
    setMessage("");
    await saveItemReview(item.id, { reviewed: event.target.checked }, false);
  };

  const deleteReport = async (id) => {
    if (!id) {
      return;
    }
    if (!window.confirm(`Delete report #${id}?`)) {
      return;
    }

    setError("");
    try {
      await axios.delete(`${apiBase}/reports/${id}`);
      if (String(reportId) === String(id)) {
        onReportIdChange("");
        setReport(null);
        setReportName("");
        setItems([]);
        setSelectedItem(null);
      }
      await loadReports();
    } catch (deleteError) {
      setError(deleteError?.response?.data?.error || "Failed to delete report.");
    }
  };

  useEffect(() => {
    loadReports();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    loadReport();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [reportId]);

  useEffect(() => {
    if (!selectedItem) {
      setReviewDraft({ reviewed: false, review_comment: "" });
      return;
    }
    setReviewDraft({
      reviewed: Boolean(selectedItem.reviewed),
      review_comment: selectedItem.review_comment || "",
    });
  }, [selectedItem]);

  const summary = useMemo(() => {
    const map = { added: 0, removed: 0, modified: 0, renamed: 0 };
    for (const item of items) {
      map[item.change_type] = (map[item.change_type] || 0) + 1;
    }
    return map;
  }, [items]);

  const reviewStats = useMemo(() => {
    const reviewed = items.filter((item) => item.reviewed).length;
    return {
      reviewed,
      total: items.length,
      unreviewed: items.length - reviewed,
      percent: items.length > 0 ? Math.round((reviewed / items.length) * 100) : 0,
    };
  }, [items]);

  const selectedReportRow = useMemo(
    () => reports.find((row) => String(row.id) === String(reportId)) || null,
    [reports, reportId]
  );

  const filteredItems = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    return items.filter((item) => {
      if (typeFilter !== "all" && item.change_type !== typeFilter) {
        return false;
      }
      if (reviewFilter === "reviewed" && !item.reviewed) {
        return false;
      }
      if (reviewFilter === "unreviewed" && item.reviewed) {
        return false;
      }
      if (!query) {
        return true;
      }
      const haystack = [
        item.safeguard_old || "",
        item.safeguard_new || "",
        item.old_text || "",
        item.new_text || "",
        item.review_comment || "",
      ]
        .join(" ")
        .toLowerCase();
      return haystack.includes(query);
    });
  }, [items, reviewFilter, typeFilter, searchQuery]);

  useEffect(() => {
    if (!selectedItem && filteredItems.length > 0) {
      setSelectedItem(filteredItems[0]);
      return;
    }
    if (selectedItem && !filteredItems.some((item) => item.id === selectedItem.id)) {
      setSelectedItem(filteredItems[0] || null);
    }
  }, [filteredItems, selectedItem]);

  const selectedIndex = useMemo(
    () => filteredItems.findIndex((item) => item.id === selectedItem?.id),
    [filteredItems, selectedItem]
  );

  const selectNextInFilter = () => {
    if (filteredItems.length === 0) {
      return;
    }
    if (selectedIndex < 0 || selectedIndex >= filteredItems.length - 1) {
      setSelectedItem(filteredItems[0]);
      return;
    }
    setSelectedItem(filteredItems[selectedIndex + 1]);
  };

  const selectPreviousInFilter = () => {
    if (filteredItems.length === 0) {
      return;
    }
    if (selectedIndex <= 0) {
      setSelectedItem(filteredItems[filteredItems.length - 1]);
      return;
    }
    setSelectedItem(filteredItems[selectedIndex - 1]);
  };

  const goToNextUnreviewed = () => {
    const unreviewed = filteredItems.filter((item) => !item.reviewed);
    if (unreviewed.length === 0) {
      setMessage("No unreviewed items in current filter.");
      return;
    }
    if (!selectedItem) {
      setSelectedItem(unreviewed[0]);
      return;
    }
    const currentIndex = unreviewed.findIndex((item) => item.id === selectedItem.id);
    if (currentIndex === -1 || currentIndex === unreviewed.length - 1) {
      setSelectedItem(unreviewed[0]);
      return;
    }
    setSelectedItem(unreviewed[currentIndex + 1]);
  };

  const saveAndNextUnreviewed = async () => {
    if (!selectedItem) {
      return;
    }
    setSavingReview(true);
    const updated = await saveItemReview(selectedItem.id, {
      reviewed: reviewDraft.reviewed,
      review_comment: reviewDraft.review_comment,
    });
    setSavingReview(false);
    if (!updated) {
      return;
    }
    setTimeout(() => {
      goToNextUnreviewed();
    }, 0);
  };

  const versionView = useMemo(() => {
    const a = versionLabels.a || "";
    const b = versionLabels.b || "";
    if (!a && !b) {
      return { oldLabel: "Old", newLabel: "New", swapSides: false };
    }

    const compare = compareVersionLabels(a, b);
    if (compare <= 0) {
      return { oldLabel: a || b, newLabel: b || a, swapSides: false };
    }
    return { oldLabel: b || a, newLabel: a || b, swapSides: true };
  }, [versionLabels]);

  const downloadUrl = (format) => {
    if (!report?.id) {
      return "#";
    }
    return `${apiBase}/reports/${report.id}/download/${format}`;
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h6">Diff Review Workspace</Typography>

        <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
          <TextField
            label="Report ID"
            value={reportId}
            onChange={(event) => onReportIdChange(event.target.value)}
            sx={{ maxWidth: 220 }}
          />
          <Button variant="contained" onClick={() => loadReport()}>
            Open Report
          </Button>
          <Button variant="outlined" onClick={loadReports}>
            Refresh
          </Button>
        </Stack>

        {report && (
          <Paper variant="outlined" sx={{ p: 2 }}>
            <Stack spacing={1.5}>
              <Stack direction={{ xs: "column", md: "row" }} spacing={1} useFlexGap flexWrap="wrap" alignItems="center">
                {reportName && <Chip label={reportName} color="primary" />}
                <Chip label={`Status: ${report.status}`} />
                <Chip label={`Reviewed: ${reviewStats.reviewed}/${reviewStats.total}`} color={reviewStats.unreviewed === 0 ? "success" : "default"} />
                <Chip label={`${reviewStats.percent}% complete`} />
                <Chip label={`Added ${summary.added || 0}`} />
                <Chip label={`Removed ${summary.removed || 0}`} />
                <Chip label={`Modified ${summary.modified || 0}`} />
                <Chip label={`Renamed ${summary.renamed || 0}`} />
              </Stack>
              <LinearProgress variant="determinate" value={reviewStats.percent} sx={{ height: 10, borderRadius: 5 }} />
              <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                <Button size="small" variant="outlined" href={downloadUrl("json")}>
                  Download JSON
                </Button>
                <Button size="small" variant="outlined" href={downloadUrl("xlsx")}>
                  Download Excel
                </Button>
                <Button size="small" variant="outlined" href={downloadUrl("html")}>
                  Download HTML
                </Button>
              </Stack>
            </Stack>
          </Paper>
        )}

        {message && <Alert severity="success">{message}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}

        <Box sx={{ display: "grid", gap: 2, gridTemplateColumns: { xs: "1fr", lg: "390px minmax(760px, 1fr)" }, alignItems: "start" }}>
          <Stack spacing={2}>
            <Paper variant="outlined" sx={{ p: 2 }}>
              <Typography variant="subtitle1" sx={{ mb: 1 }}>
                Reports
              </Typography>
              <Stack spacing={1}>
                <FormControl size="small" fullWidth>
                  <InputLabel id="report-selector-label">Report</InputLabel>
                  <Select
                    labelId="report-selector-label"
                    value={reportId ? String(reportId) : ""}
                    label="Report"
                    onChange={(event) => {
                      const value = String(event.target.value || "");
                      onReportIdChange(value);
                      if (value) {
                        loadReport(value);
                      }
                    }}
                  >
                    {reports.map((row) => (
                      <MenuItem key={row.id} value={String(row.id)}>
                        #{row.id} {row.report_name || `${row.framework} ${row.version_a} -> ${row.version_b}`}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() => {
                      if (reportId) {
                        loadReport(reportId);
                      }
                    }}
                    disabled={!reportId}
                  >
                    Open Selected
                  </Button>
                  <Button
                    size="small"
                    variant="outlined"
                    color="error"
                    startIcon={<DeleteOutlineIcon fontSize="small" />}
                    onClick={() => deleteReport(reportId)}
                    disabled={!reportId}
                  >
                    Delete Selected
                  </Button>
                </Stack>
                {selectedReportRow && (
                  <Typography variant="caption" color="text.secondary">
                    Status: {selectedReportRow.status} | Items: {selectedReportRow.item_count}
                  </Typography>
                )}
              </Stack>
            </Paper>

            <Paper variant="outlined" sx={{ p: 2 }}>
              <Typography variant="subtitle1" sx={{ mb: 1 }}>
                Review Queue ({filteredItems.length}/{items.length})
              </Typography>
              <Stack spacing={1.25}>
                <FormControl size="small" fullWidth>
                  <InputLabel id="type-filter-label">Change Type</InputLabel>
                  <Select
                    labelId="type-filter-label"
                    value={typeFilter}
                    label="Change Type"
                    onChange={(event) => setTypeFilter(event.target.value)}
                  >
                    <MenuItem value="all">All Types</MenuItem>
                    {CHANGE_TYPES.map((changeType) => (
                      <MenuItem key={changeType} value={changeType}>
                        {changeType}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <FormControl size="small" fullWidth>
                  <InputLabel id="review-filter-label">Review State</InputLabel>
                  <Select
                    labelId="review-filter-label"
                    value={reviewFilter}
                    label="Review State"
                    onChange={(event) => setReviewFilter(event.target.value)}
                  >
                    <MenuItem value="all">All</MenuItem>
                    <MenuItem value="reviewed">Reviewed</MenuItem>
                    <MenuItem value="unreviewed">Unreviewed</MenuItem>
                  </Select>
                </FormControl>
                <TextField
                  size="small"
                  label="Search by safeguard or text"
                  value={searchQuery}
                  onChange={(event) => setSearchQuery(event.target.value)}
                />
                <Button size="small" variant="text" onClick={goToNextUnreviewed}>
                  Jump to Next Unreviewed
                </Button>
              </Stack>
              <List sx={{ mt: 1.5, maxHeight: 560, overflow: "auto", border: "1px solid #ddd", borderRadius: 1 }}>
                {filteredItems.map((item) => (
                  <ListItem key={item.id} disablePadding>
                    <ListItemButton onClick={() => setSelectedItem(item)} selected={selectedItem?.id === item.id} alignItems="flex-start">
                      <Checkbox
                        size="small"
                        checked={Boolean(item.reviewed)}
                        onClick={(event) => event.stopPropagation()}
                        onChange={(event) => toggleReviewedInline(event, item)}
                      />
                      <ListItemText
                        primary={
                          <Stack direction="row" spacing={1} alignItems="center">
                            <Chip size="small" label={item.change_type} />
                            <Typography variant="body2" sx={{ fontWeight: 600 }}>
                              {item.safeguard_new || item.safeguard_old}
                            </Typography>
                          </Stack>
                        }
                        secondary={
                          <Stack spacing={0.25} sx={{ mt: 0.5 }}>
                            <Typography variant="caption" color="text.secondary">
                              {compactText(item.new_text || item.old_text || "No text")}
                            </Typography>
                            {item.review_comment && (
                              <Typography variant="caption" color="info.main">
                                Comment saved
                              </Typography>
                            )}
                          </Stack>
                        }
                      />
                    </ListItemButton>
                  </ListItem>
                ))}
                {report && filteredItems.length === 0 && (
                  <Typography variant="body2" sx={{ p: 2 }}>
                    No items match the current filter.
                  </Typography>
                )}
              </List>
            </Paper>
          </Stack>

          <Stack spacing={2}>
            <Paper variant="outlined" sx={{ p: 2, minHeight: 760 }}>
              <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1 }}>
                <Typography variant="subtitle1">Selected Change</Typography>
                <Stack direction="row" spacing={1}>
                  <Button size="small" variant="outlined" onClick={selectPreviousInFilter} disabled={filteredItems.length === 0}>
                    Previous
                  </Button>
                  <Button size="small" variant="outlined" onClick={selectNextInFilter} disabled={filteredItems.length === 0}>
                    Next
                  </Button>
                </Stack>
              </Stack>
              <DiffPane
                item={selectedItem}
                oldVersionLabel={versionView.oldLabel}
                newVersionLabel={versionView.newLabel}
                swapSides={versionView.swapSides}
              />
              {selectedItem && (
                <Box sx={{ mt: 2, pt: 2, borderTop: "1px solid #e0e0e0" }}>
                  <Stack spacing={2}>
                    <Typography variant="subtitle1">Review Decision</Typography>
                    <FormControlLabel
                      control={
                        <Checkbox
                          checked={reviewDraft.reviewed}
                          onChange={(event) =>
                            setReviewDraft((previous) => ({ ...previous, reviewed: event.target.checked }))
                          }
                        />
                      }
                      label="Mark this item as reviewed"
                    />
                    <TextField
                      label="Reviewer comments"
                      value={reviewDraft.review_comment}
                      onChange={(event) =>
                        setReviewDraft((previous) => ({ ...previous, review_comment: event.target.value }))
                      }
                      fullWidth
                      multiline
                      minRows={4}
                      placeholder="Capture notes, rationale, and follow-up actions."
                    />
                    <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                      <Button variant="contained" onClick={saveSelectedReview} disabled={savingReview}>
                        {savingReview ? "Saving..." : "Save"}
                      </Button>
                      <Button variant="outlined" onClick={saveAndNextUnreviewed} disabled={savingReview}>
                        Save & Next Unreviewed
                      </Button>
                      <Button
                        variant="text"
                        onClick={() =>
                          setReviewDraft({
                            reviewed: Boolean(selectedItem.reviewed),
                            review_comment: selectedItem.review_comment || "",
                          })
                        }
                      >
                        Reset
                      </Button>
                    </Stack>
                    {selectedItem.reviewed_at && (
                      <Typography variant="caption" color="text.secondary">
                        Reviewed at: {new Date(selectedItem.reviewed_at).toLocaleString()}
                      </Typography>
                    )}
                  </Stack>
                </Box>
              )}
            </Paper>
          </Stack>
        </Box>
      </Stack>
    </Paper>
  );
}
