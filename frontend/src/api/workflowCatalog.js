import axios from "axios";

export async function fetchWorkflowCatalog(apiBase) {
  const response = await axios.get(`${apiBase}/api/workflow/catalog`);
  return response.data || {};
}

