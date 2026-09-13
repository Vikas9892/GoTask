// API Client using standard Fetch API
const BASE_URL = '';

export async function fetchHealth() {
  const res = await fetch(`${BASE_URL}/ready`);
  if (!res.ok) throw new Error(`Health check failed (${res.status})`);
  return res.json();
}

export async function fetchJobs(limit = 50, offset = 0) {
  const res = await fetch(`${BASE_URL}/api/jobs?limit=${limit}&offset=${offset}`);
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to fetch jobs (${res.status})`);
  }
  return res.json();
}

export async function fetchJobDetails(id) {
  const res = await fetch(`${BASE_URL}/api/jobs/${id}`);
  if (!res.ok) throw new Error(`Failed to fetch job details (${res.status})`);
  return res.json();
}

export async function fetchJobAttempts(id) {
  const res = await fetch(`${BASE_URL}/api/jobs/${id}/attempts`);
  if (!res.ok) throw new Error(`Failed to fetch job attempts (${res.status})`);
  return res.json();
}

export async function submitJob(type, payload) {
  const res = await fetch(`${BASE_URL}/api/jobs`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ type, payload }),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error || `Job submission failed (${res.status})`);
  }
  return res.json();
}

export async function deleteJob(id) {
  const res = await fetch(`${BASE_URL}/api/jobs/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error(`Failed to delete job (${res.status})`);
}
