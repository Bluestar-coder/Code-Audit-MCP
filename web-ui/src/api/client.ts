export interface ScanFileRequest {
  file_path: string;
  language?: string;
  content?: string;
  rule_ids?: string[];
}

export interface VulnerabilityFinding {
  rule_id: string;
  rule_name: string;
  category: string;
  severity: string;
  message: string;
  file_path: string;
  line: number;
  column?: number;
  code: string;
  language: string;
  metadata: string; // JSON string
}

export interface ScanFileResponse {
  success: boolean;
  file_path: string;
  language: string;
  findings: VulnerabilityFinding[];
  statistics: string; // JSON string
  error?: string;
}

export interface GetRulesResponse {
  rules: Array<{
    id: string;
    name: string;
    category: string;
    severity: string;
    description: string;
    language: string;
  }>;
}

const API_BASE = process.env.REACT_APP_API_BASE || 'http://localhost:8080/api';

export async function scanFile(params: ScanFileRequest): Promise<ScanFileResponse> {
  const res = await fetch(`${API_BASE}/scan`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }
  return res.json();
}

export async function getRules(language?: string): Promise<GetRulesResponse> {
  const url = new URL(`${API_BASE}/rules`);
  if (language) url.searchParams.set('language', language);
  const res = await fetch(url.toString());
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }
  return res.json();
}

// 新增接口定义
export interface ProjectStats {
  total_files: number;
  total_lines: number;
  total_functions: number;
  total_classes: number;
  languages: { [key: string]: number };
  last_scan_time?: string;
}

export interface VulnerabilityStats {
  total: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  fixed: number;
  by_category: { [key: string]: number };
}

export interface ScanHistory {
  id: string;
  timestamp: string;
  project_path: string;
  files_scanned: number;
  vulnerabilities_found: number;
  duration_ms: number;
  status: 'completed' | 'failed' | 'running';
}

export interface DashboardData {
  project_stats: ProjectStats;
  vulnerability_stats: VulnerabilityStats;
  scan_history: ScanHistory[];
  trend_data: Array<{
    date: string;
    vulnerabilities: number;
    fixed: number;
  }>;
}

// 新增 API 函数
export async function getDashboardData(): Promise<DashboardData> {
  const response = await fetch(`${API_BASE}/dashboard`);
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function getProjectStats(): Promise<ProjectStats> {
  const response = await fetch(`${API_BASE}/stats/project`);
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function getVulnerabilityStats(): Promise<VulnerabilityStats> {
  const response = await fetch(`${API_BASE}/stats/vulnerabilities`);
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function getScanHistory(limit: number = 10): Promise<ScanHistory[]> {
  const response = await fetch(`${API_BASE}/scans/history?limit=${limit}`);
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

// ===== 污点分析 API =====
export interface SourceInfo {
  id: string;
  name: string;
  type: string;
  keywords: string[];
  description: string;
}

export interface QuerySourcesResponse {
  sources: SourceInfo[];
  total_count: number;
}

export interface SinkInfo {
  id: string;
  name: string;
  type: string;
  keywords: string[];
  vulnerability_type: string;
  description: string;
}

export interface QuerySinksResponse {
  sinks: SinkInfo[];
  total_count: number;
}

export interface TracePathRequest {
  source_function: string;
  sink_function: string;
  max_paths?: number;
}

export interface TracePathNode {
  node_id: string;
  function_name: string;
  file_path: string;
  line_number: number;
  operation: string;
  variable_name: string;
  data_flow: string;
}

export interface TracePathSegment {
  path_index: number;
  nodes: TracePathNode[];
  has_sanitizer: boolean;
}

export interface TracePathResponse {
  paths: TracePathSegment[];
}

export async function queryTaintSources(pattern: string = '', language: string = ''): Promise<QuerySourcesResponse> {
  const url = new URL(`${API_BASE}/taint/sources`);
  if (pattern) url.searchParams.set('pattern', pattern);
  if (language) url.searchParams.set('language', language);
  const res = await fetch(url.toString());
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }
  return res.json();
}

export async function queryTaintSinks(pattern: string = '', language: string = ''): Promise<QuerySinksResponse> {
  const url = new URL(`${API_BASE}/taint/sinks`);
  if (pattern) url.searchParams.set('pattern', pattern);
  if (language) url.searchParams.set('language', language);
  const res = await fetch(url.toString());
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }
  return res.json();
}

export async function traceTaintPaths(req: TracePathRequest): Promise<TracePathResponse> {
  const res = await fetch(`${API_BASE}/taint/trace`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }
  return res.json();
}