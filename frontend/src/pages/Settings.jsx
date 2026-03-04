import React, { useEffect, useMemo, useState } from "react";
import axios from "axios";
import {
  Alert,
  Button,
  Checkbox,
  FormControl,
  FormControlLabel,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";

function emptyBranding() {
  return {
    org_name: "",
    logo_url: "",
    primary_color: "",
    secondary_color: "",
    support_email: "",
  };
}

export default function Settings({ apiBase }) {
  const [branding, setBranding] = useState(emptyBranding());
  const [roles, setRoles] = useState([]);
  const [users, setUsers] = useState([]);
  const [newRole, setNewRole] = useState({ name: "", description: "" });
  const [newUser, setNewUser] = useState({ email: "", display_name: "", role_id: "", is_active: true });
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const loadAll = async () => {
    setError("");
    try {
      const [brandingRes, rolesRes, usersRes] = await Promise.all([
        axios.get(`${apiBase}/settings/branding`),
        axios.get(`${apiBase}/settings/roles`),
        axios.get(`${apiBase}/settings/users`),
      ]);

      setBranding(brandingRes.data || emptyBranding());
      setRoles((rolesRes.data || []).map((role) => ({ ...role, _dirty: false })));
      setUsers(
        (usersRes.data || []).map((user) => ({
          ...user,
          _role_id_value: user.role_id ?? "",
          _dirty: false,
        }))
      );
    } catch (loadError) {
      setError(loadError?.response?.data?.error || "Failed to load settings.");
    }
  };

  useEffect(() => {
    loadAll();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiBase]);

  const roleOptions = useMemo(
    () => roles.map((role) => ({ id: role.id, label: role.name })),
    [roles]
  );

  const saveBranding = async () => {
    setError("");
    setMessage("");
    try {
      const response = await axios.put(`${apiBase}/settings/branding`, branding);
      setBranding(response.data || emptyBranding());
      setMessage("Branding settings saved.");
    } catch (saveError) {
      setError(saveError?.response?.data?.error || "Failed to save branding settings.");
    }
  };

  const createRole = async () => {
    if (!newRole.name.trim()) {
      setError("Role name is required.");
      return;
    }
    setError("");
    setMessage("");
    try {
      await axios.post(`${apiBase}/settings/roles`, {
        name: newRole.name,
        description: newRole.description,
      });
      setNewRole({ name: "", description: "" });
      setMessage("Role created.");
      loadAll();
    } catch (createError) {
      setError(createError?.response?.data?.error || "Failed to create role.");
    }
  };

  const saveRole = async (role) => {
    setError("");
    setMessage("");
    try {
      await axios.put(`${apiBase}/settings/roles/${role.id}`, {
        name: role.name,
        description: role.description,
      });
      setMessage(`Role "${role.name}" saved.`);
      loadAll();
    } catch (saveError) {
      setError(saveError?.response?.data?.error || "Failed to save role.");
    }
  };

  const deleteRole = async (role) => {
    if (!window.confirm(`Delete role "${role.name}"?`)) {
      return;
    }
    setError("");
    setMessage("");
    try {
      await axios.delete(`${apiBase}/settings/roles/${role.id}`);
      setMessage(`Role "${role.name}" deleted.`);
      loadAll();
    } catch (deleteError) {
      setError(deleteError?.response?.data?.error || "Failed to delete role.");
    }
  };

  const createUser = async () => {
    if (!newUser.email.trim()) {
      setError("User email is required.");
      return;
    }
    setError("");
    setMessage("");
    try {
      await axios.post(`${apiBase}/settings/users`, {
        email: newUser.email,
        display_name: newUser.display_name,
        role_id: newUser.role_id === "" ? null : Number(newUser.role_id),
        is_active: newUser.is_active,
      });
      setNewUser({ email: "", display_name: "", role_id: "", is_active: true });
      setMessage("User created.");
      loadAll();
    } catch (createError) {
      setError(createError?.response?.data?.error || "Failed to create user.");
    }
  };

  const saveUser = async (user) => {
    setError("");
    setMessage("");
    try {
      await axios.put(`${apiBase}/settings/users/${user.id}`, {
        email: user.email,
        display_name: user.display_name,
        role_id: user._role_id_value === "" ? null : Number(user._role_id_value),
        clear_role: user._role_id_value === "",
        is_active: user.is_active,
      });
      setMessage(`User "${user.email}" saved.`);
      loadAll();
    } catch (saveError) {
      setError(saveError?.response?.data?.error || "Failed to save user.");
    }
  };

  const deleteUser = async (user) => {
    if (!window.confirm(`Delete user "${user.email}"?`)) {
      return;
    }
    setError("");
    setMessage("");
    try {
      await axios.delete(`${apiBase}/settings/users/${user.id}`);
      setMessage(`User "${user.email}" deleted.`);
      loadAll();
    } catch (deleteError) {
      setError(deleteError?.response?.data?.error || "Failed to delete user.");
    }
  };

  return (
    <Stack spacing={2}>
      <Typography variant="h6">Settings</Typography>
      {message && <Alert severity="success">{message}</Alert>}
      {error && <Alert severity="error">{error}</Alert>}

      <Paper sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Typography variant="subtitle1">Org Branding</Typography>
          <TextField
            label="Organization Name"
            value={branding.org_name || ""}
            onChange={(event) => setBranding((prev) => ({ ...prev, org_name: event.target.value }))}
            fullWidth
          />
          <TextField
            label="Logo URL"
            value={branding.logo_url || ""}
            onChange={(event) => setBranding((prev) => ({ ...prev, logo_url: event.target.value }))}
            fullWidth
          />
          <Stack direction={{ xs: "column", md: "row" }} spacing={2}>
            <TextField
              label="Primary Color"
              value={branding.primary_color || ""}
              onChange={(event) => setBranding((prev) => ({ ...prev, primary_color: event.target.value }))}
              fullWidth
            />
            <TextField
              label="Secondary Color"
              value={branding.secondary_color || ""}
              onChange={(event) => setBranding((prev) => ({ ...prev, secondary_color: event.target.value }))}
              fullWidth
            />
          </Stack>
          <TextField
            label="Support Email"
            value={branding.support_email || ""}
            onChange={(event) => setBranding((prev) => ({ ...prev, support_email: event.target.value }))}
            fullWidth
          />
          <Stack direction="row" spacing={1}>
            <Button variant="contained" onClick={saveBranding}>
              Save Branding
            </Button>
            <Button variant="outlined" onClick={loadAll}>
              Refresh Settings
            </Button>
          </Stack>
        </Stack>
      </Paper>

      <Paper sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Typography variant="subtitle1">Roles</Typography>
          <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
            <TextField
              label="New Role"
              value={newRole.name}
              onChange={(event) => setNewRole((prev) => ({ ...prev, name: event.target.value }))}
              fullWidth
            />
            <TextField
              label="Description"
              value={newRole.description}
              onChange={(event) => setNewRole((prev) => ({ ...prev, description: event.target.value }))}
              fullWidth
            />
            <Button variant="contained" onClick={createRole}>
              Add Role
            </Button>
          </Stack>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Description</TableCell>
                <TableCell>System</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {roles.map((role) => (
                <TableRow key={role.id}>
                  <TableCell sx={{ minWidth: 180 }}>
                    <TextField
                      size="small"
                      value={role.name}
                      onChange={(event) =>
                        setRoles((prev) =>
                          prev.map((item) =>
                            item.id === role.id ? { ...item, name: event.target.value, _dirty: true } : item
                          )
                        )
                      }
                      disabled={role.is_system}
                      fullWidth
                    />
                  </TableCell>
                  <TableCell sx={{ minWidth: 280 }}>
                    <TextField
                      size="small"
                      value={role.description || ""}
                      onChange={(event) =>
                        setRoles((prev) =>
                          prev.map((item) =>
                            item.id === role.id ? { ...item, description: event.target.value, _dirty: true } : item
                          )
                        )
                      }
                      fullWidth
                    />
                  </TableCell>
                  <TableCell>{role.is_system ? "Yes" : "No"}</TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={1}>
                      <Button size="small" variant="outlined" onClick={() => saveRole(role)}>
                        Save
                      </Button>
                      <Button
                        size="small"
                        color="error"
                        variant="outlined"
                        onClick={() => deleteRole(role)}
                        disabled={role.is_system}
                      >
                        Delete
                      </Button>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Stack>
      </Paper>

      <Paper sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Typography variant="subtitle1">Users</Typography>
          <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
            <TextField
              label="Email"
              value={newUser.email}
              onChange={(event) => setNewUser((prev) => ({ ...prev, email: event.target.value }))}
              fullWidth
            />
            <TextField
              label="Display Name"
              value={newUser.display_name}
              onChange={(event) => setNewUser((prev) => ({ ...prev, display_name: event.target.value }))}
              fullWidth
            />
            <FormControl size="small" fullWidth>
              <InputLabel id="new-user-role-label">Role</InputLabel>
              <Select
                labelId="new-user-role-label"
                label="Role"
                value={newUser.role_id}
                onChange={(event) => setNewUser((prev) => ({ ...prev, role_id: event.target.value }))}
              >
                <MenuItem value="">No role</MenuItem>
                {roleOptions.map((role) => (
                  <MenuItem key={role.id} value={String(role.id)}>
                    {role.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <FormControlLabel
              control={
                <Checkbox
                  checked={newUser.is_active}
                  onChange={(event) => setNewUser((prev) => ({ ...prev, is_active: event.target.checked }))}
                />
              }
              label="Active"
            />
            <Button variant="contained" onClick={createUser}>
              Add User
            </Button>
          </Stack>

          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Email</TableCell>
                <TableCell>Display Name</TableCell>
                <TableCell>Role</TableCell>
                <TableCell>Active</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {users.map((user) => (
                <TableRow key={user.id}>
                  <TableCell sx={{ minWidth: 220 }}>
                    <TextField
                      size="small"
                      value={user.email}
                      onChange={(event) =>
                        setUsers((prev) =>
                          prev.map((item) =>
                            item.id === user.id ? { ...item, email: event.target.value, _dirty: true } : item
                          )
                        )
                      }
                      fullWidth
                    />
                  </TableCell>
                  <TableCell sx={{ minWidth: 200 }}>
                    <TextField
                      size="small"
                      value={user.display_name || ""}
                      onChange={(event) =>
                        setUsers((prev) =>
                          prev.map((item) =>
                            item.id === user.id ? { ...item, display_name: event.target.value, _dirty: true } : item
                          )
                        )
                      }
                      fullWidth
                    />
                  </TableCell>
                  <TableCell sx={{ minWidth: 180 }}>
                    <FormControl size="small" fullWidth>
                      <Select
                        value={String(user._role_id_value ?? "")}
                        onChange={(event) =>
                          setUsers((prev) =>
                            prev.map((item) =>
                              item.id === user.id ? { ...item, _role_id_value: event.target.value, _dirty: true } : item
                            )
                          )
                        }
                      >
                        <MenuItem value="">No role</MenuItem>
                        {roleOptions.map((role) => (
                          <MenuItem key={role.id} value={String(role.id)}>
                            {role.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                  </TableCell>
                  <TableCell>
                    <Checkbox
                      checked={Boolean(user.is_active)}
                      onChange={(event) =>
                        setUsers((prev) =>
                          prev.map((item) =>
                            item.id === user.id ? { ...item, is_active: event.target.checked, _dirty: true } : item
                          )
                        )
                      }
                    />
                  </TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={1}>
                      <Button size="small" variant="outlined" onClick={() => saveUser(user)}>
                        Save
                      </Button>
                      <Button size="small" color="error" variant="outlined" onClick={() => deleteUser(user)}>
                        Delete
                      </Button>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Stack>
      </Paper>
    </Stack>
  );
}
