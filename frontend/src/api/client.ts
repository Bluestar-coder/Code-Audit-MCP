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
// Standalone 开关：REACT_APP_STANDALONE=true/1 时使用本地 mock
const STANDALONE = String(process.env.REACT_APP_STANDALONE || '').toLowerCase() === 'true' || process.env.REACT_APP_STANDALONE === '1';

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
// ===== 仪表盘 mock 数据（Standalone 模式） =====
const MOCK_DASHBOARD: DashboardData = {
  project_stats: {
    total_files: 156,
    total_lines: 45230,
    total_functions: 892,
    total_classes: 234,
    languages: { 'JavaScript': 45, 'TypeScript': 35, 'Python': 15, 'Go': 5 },
    last_scan_time: new Date().toISOString(),
  },
  vulnerability_stats: {
    total: 47,
    critical: 3,
    high: 8,
    medium: 15,
    low: 21,
    fixed: 12,
    by_category: { 'SQL注入': 8, 'XSS': 12, '命令注入': 5, '路径遍历': 7, '其他': 15 },
  },
  scan_history: [
    {
      id: '1',
      timestamp: new Date(Date.now() - 3600000).toISOString(),
      project_path: '/current/project',
      files_scanned: 156,
      vulnerabilities_found: 47,
      duration_ms: 12500,
      status: 'completed',
    },
    {
      id: '2',
      timestamp: new Date(Date.now() - 7200000).toISOString(),
      project_path: '/current/project',
      files_scanned: 150,
      vulnerabilities_found: 52,
      duration_ms: 11800,
      status: 'completed',
    },
  ],
  trend_data: [
    { date: '1月', vulnerabilities: 52, fixed: 8 },
    { date: '2月', vulnerabilities: 48, fixed: 12 },
    { date: '3月', vulnerabilities: 45, fixed: 15 },
    { date: '4月', vulnerabilities: 47, fixed: 12 },
  ],
};

export async function getDashboardData(): Promise<DashboardData> {
  if (STANDALONE) {
    return MOCK_DASHBOARD;
  }
  const response = await fetch(`${API_BASE}/dashboard`);
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function getProjectStats(): Promise<ProjectStats> {
  if (STANDALONE) {
    return MOCK_DASHBOARD.project_stats;
  }
  const response = await fetch(`${API_BASE}/stats/project`);
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function getVulnerabilityStats(): Promise<VulnerabilityStats> {
  if (STANDALONE) {
    return MOCK_DASHBOARD.vulnerability_stats;
  }
  const response = await fetch(`${API_BASE}/stats/vulnerabilities`);
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function getScanHistory(limit: number = 10): Promise<ScanHistory[]> {
  if (STANDALONE) {
    return MOCK_DASHBOARD.scan_history.slice(0, limit);
  }
  const response = await fetch(`${API_BASE}/scans/history?limit=${limit}`);
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

// ===== 污点分析 mock 数据（Standalone 模式） =====
const MOCK_SOURCES: SourceInfo[] = [
  { id: 'source_1', name: 'http.Request.FormValue', type: 'HTTP Input', keywords: ['form', 'input'], description: 'HTTP表单输入，用户可控数据源' },
  { id: 'source_2', name: 'http.Request.URL.Query', type: 'HTTP Input', keywords: ['query', 'url'], description: 'URL查询参数，用户可控数据源' },
  { id: 'source_3', name: 'json.Unmarshal', type: 'Deserialization', keywords: ['json', 'unmarshal'], description: 'JSON反序列化，外部数据源' },
  { id: 'source_4', name: 'os.Getenv', type: 'Environment', keywords: ['env', 'config'], description: '环境变量读取' },
];

const MOCK_SINKS: SinkInfo[] = [
  { id: 'sink_1', name: 'database.Query', type: 'Database', keywords: ['sql', 'query'], vulnerability_type: 'SQL注入', description: 'SQL查询执行，潜在SQL注入点' },
  { id: 'sink_2', name: 'os.Exec', type: 'Command Execution', keywords: ['exec', 'command'], vulnerability_type: '命令注入', description: '系统命令执行，潜在命令注入点' },
];

const MOCK_TRACE_RESPONSE: TracePathResponse = {
  paths: [
    {
      path_index: 0,
      has_sanitizer: false,
      nodes: [
        { node_id: 'n1', function_name: 'parseJSON', file_path: 'src/api/handler.go', line_number: 78, operation: 'assignment', variable_name: 'data', data_flow: 'external_input' },
        { node_id: 'n2', function_name: 'renderTemplate', file_path: 'src/views/render.go', line_number: 34, operation: 'template_render', variable_name: 'content', data_flow: 'output_operation' },
      ],
    },
  ],
};

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
  if (STANDALONE) {
    const filtered = pattern ? MOCK_SOURCES.filter(s => s.name.toLowerCase().includes(pattern.toLowerCase())) : MOCK_SOURCES;
    return { sources: filtered, total_count: filtered.length };
  }
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
  if (STANDALONE) {
    const filtered = pattern ? MOCK_SINKS.filter(s => s.name.toLowerCase().includes(pattern.toLowerCase())) : MOCK_SINKS;
    return { sinks: filtered, total_count: filtered.length };
  }
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
  if (STANDALONE) {
    return MOCK_TRACE_RESPONSE;
  }
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