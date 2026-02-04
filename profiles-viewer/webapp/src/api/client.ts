import { AppsResponse, BucketsResponse, FlameGraphNode } from './types';

const API_BASE = '/api';

export async function fetchApps(): Promise<AppsResponse> {
  const response = await fetch(`${API_BASE}/apps`);
  if (!response.ok) {
    throw new Error(`Failed to fetch apps: ${response.statusText}`);
  }
  return response.json();
}

export async function fetchBuckets(
  namespace: string,
  kind: string,
  name: string
): Promise<BucketsResponse> {
  const response = await fetch(
    `${API_BASE}/apps/${encodeURIComponent(namespace)}/${encodeURIComponent(kind)}/${encodeURIComponent(name)}/buckets`
  );
  if (!response.ok) {
    throw new Error(`Failed to fetch buckets: ${response.statusText}`);
  }
  return response.json();
}

export async function fetchFlameGraph(
  namespace: string,
  kind: string,
  name: string,
  start?: string,
  end?: string
): Promise<FlameGraphNode> {
  const params = new URLSearchParams();
  if (start) params.append('start', start);
  if (end) params.append('end', end);

  const queryString = params.toString();
  const url = `${API_BASE}/apps/${encodeURIComponent(namespace)}/${encodeURIComponent(kind)}/${encodeURIComponent(name)}/flamegraph${queryString ? `?${queryString}` : ''}`;

  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Failed to fetch flame graph: ${response.statusText}`);
  }
  return response.json();
}

export async function fetchStats(): Promise<Record<string, unknown>> {
  const response = await fetch(`${API_BASE}/stats`);
  if (!response.ok) {
    throw new Error(`Failed to fetch stats: ${response.statusText}`);
  }
  return response.json();
}
