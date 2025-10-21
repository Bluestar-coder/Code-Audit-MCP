import React, { useState, useEffect, useRef, useCallback } from 'react';
import Grid from '@mui/material/Grid';
import {
  Box,
  Typography,
  Paper,
  TextField,
  Button,
  List,
  ListItem,
  ListItemText,
  Divider,
  Chip,
  Alert,
  CircularProgress,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  IconButton,
  Tooltip,
  Card,
  CardContent,
  CardActions,
   Tab,
  Tabs,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Switch,
  FormControlLabel,
  LinearProgress,
} from '@mui/material';
import PlayArrow from '@mui/icons-material/PlayArrow';
import ExpandMore from '@mui/icons-material/ExpandMore';
import Timeline from '@mui/icons-material/Timeline';
import Source from '@mui/icons-material/Source';
import CallSplit from '@mui/icons-material/CallSplit';
import Info from '@mui/icons-material/Info';
import Refresh from '@mui/icons-material/Refresh';
import Visibility from '@mui/icons-material/Visibility';
import Download from '@mui/icons-material/Download';
import Settings from '@mui/icons-material/Settings';
import AccountTree from '@mui/icons-material/AccountTree';

import Security from '@mui/icons-material/Security';
import BugReport from '@mui/icons-material/BugReport';
import FilterList from '@mui/icons-material/FilterList';
import ZoomIn from '@mui/icons-material/ZoomIn';
import ZoomOut from '@mui/icons-material/ZoomOut';
import CenterFocusStrong from '@mui/icons-material/CenterFocusStrong';
import * as d3 from 'd3';
import { queryTaintSources, queryTaintSinks, traceTaintPaths } from '../api/client';

interface TaintPath {
  id: string;
  source: string;
  sink: string;
  path: PathSegment[];
  riskLevel: 'high' | 'medium' | 'low';
  confidence: number;
  vulnerability_type: string;
  description: string;
}

interface PathSegment {
  function: string;
  file: string;
  line: number;
  variable: string;
  operation: string;
  taint_type: string;
}

interface SourceSink {
  name: string;
  type: 'source' | 'sink';
  file: string;
  line: number;
  description: string;
  category: string;
  risk_level: string;
}

interface TaintFlowNode {
  id: string;
  label: string;
  type: 'source' | 'sink' | 'intermediate';
  file: string;
  line: number;
  risk: string;
}

interface TaintFlowEdge {
  source: string;
  target: string;
  label: string;
  type: string;
}



// 模拟数据（模块级常量，避免 Hook 依赖警告）
const mockSources: SourceSink[] = [
  {
    name: 'http.Request.FormValue',
    type: 'source',
    file: 'src/handlers/user.go',
    line: 25,
    description: 'HTTP表单输入，用户可控数据源',
    category: 'HTTP Input',
    risk_level: 'high',
  },
  {
    name: 'http.Request.URL.Query',
    type: 'source',
    file: 'src/handlers/search.go',
    line: 42,
    description: 'URL查询参数，用户可控数据源',
    category: 'HTTP Input',
    risk_level: 'high',
  },
  {
    name: 'json.Unmarshal',
    type: 'source',
    file: 'src/api/handler.go',
    line: 78,
    description: 'JSON反序列化，外部数据源',
    category: 'Deserialization',
    risk_level: 'medium',
  },
  {
    name: 'os.Getenv',
    type: 'source',
    file: 'src/config/env.go',
    line: 15,
    description: '环境变量读取',
    category: 'Environment',
    risk_level: 'low',
  },
];

const mockSinks: SourceSink[] = [
  {
    name: 'database.Query',
    type: 'sink',
    file: 'src/database/user.go',
    line: 156,
    description: 'SQL查询执行，潜在SQL注入点',
    category: 'Database',
    risk_level: 'high',
  },
  {
    name: 'os.Exec',
    type: 'sink',
    file: 'src/utils/system.go',
    line: 89,
    description: '系统命令执行，潜在命令注入点',
    category: 'Command Execution',
    risk_level: 'high',
  },
  {
    name: 'template.Execute',
    type: 'sink',
    file: 'src/views/render.go',
    line: 34,
    description: '模板渲染，潜在XSS注入点',
    category: 'Template',
    risk_level: 'medium',
  },
  {
    name: 'log.Printf',
    type: 'sink',
    file: 'src/utils/logger.go',
    line: 67,
    description: '日志输出，信息泄露风险',
    category: 'Logging',
    risk_level: 'low',
  },
];

const TaintAnalysis: React.FC = () => {
  const [sourceFunction, setSourceFunction] = useState('');
  const [sinkFunction, setSinkFunction] = useState('');
  const [loading, setLoading] = useState(false);
  const [taintPaths, setTaintPaths] = useState<TaintPath[]>([]);
  const [sources, setSources] = useState<SourceSink[]>([]);
  const [sinks, setSinks] = useState<SourceSink[]>([]);
  const [sourcePattern, setSourcePattern] = useState('');
  const [sinkPattern, setSinkPattern] = useState('');
  const [selectedPath, setSelectedPath] = useState<TaintPath | null>(null);
  const [tabValue, setTabValue] = useState(0);
  const [projectPath, setProjectPath] = useState('');
  const [maxDepth, setMaxDepth] = useState(10);
  const [includeSanitizers, setIncludeSanitizers] = useState(true);
  const [filterRisk, setFilterRisk] = useState<string>('all');
  const [statistics, setStatistics] = useState({
    total_paths: 0,
    high_risk_paths: 0,
    sources_found: 0,
    sinks_found: 0,
  });
  const [error, setError] = useState<string | null>(null);
  
  const svgRef = useRef<SVGSVGElement>(null);
  const [, setFlowData] = useState<{ nodes: TaintFlowNode[], edges: TaintFlowEdge[] }>({ nodes: [], edges: [] });

  // 模拟数据（mockSources/mockSinks 已移至模块级）

  const mockTaintPaths: TaintPath[] = [
    {
      id: '1',
      source: 'http.Request.FormValue',
      sink: 'database.Query',
      riskLevel: 'high',
      confidence: 0.95,
      vulnerability_type: 'SQL Injection',
      description: '用户输入直接拼接到SQL查询中，存在SQL注入风险',
      path: [
        {
          function: 'getUserInput',
          file: 'src/handlers/user.go',
          line: 25,
          variable: 'username',
          operation: 'assignment',
          taint_type: 'user_input',
        },
        {
          function: 'validateUser',
          file: 'src/services/auth.go',
          line: 67,
          variable: 'userParam',
          operation: 'parameter_pass',
          taint_type: 'propagation',
        },
        {
          function: 'queryUser',
          file: 'src/database/user.go',
          line: 156,
          variable: 'query',
          operation: 'string_concat',
          taint_type: 'dangerous_operation',
        },
      ],
    },
    {
      id: '2',
      source: 'http.Request.URL.Query',
      sink: 'os.Exec',
      riskLevel: 'high',
      confidence: 0.88,
      vulnerability_type: 'Command Injection',
      description: 'URL参数直接传递给系统命令执行，存在命令注入风险',
      path: [
        {
          function: 'handleCommand',
          file: 'src/handlers/system.go',
          line: 42,
          variable: 'cmd',
          operation: 'assignment',
          taint_type: 'user_input',
        },
        {
          function: 'executeCommand',
          file: 'src/utils/system.go',
          line: 89,
          variable: 'command',
          operation: 'parameter_pass',
          taint_type: 'dangerous_operation',
        },
      ],
    },
    {
      id: '3',
      source: 'json.Unmarshal',
      sink: 'template.Execute',
      riskLevel: 'medium',
      confidence: 0.72,
      vulnerability_type: 'XSS',
      description: 'JSON数据未经过滤直接渲染到模板，存在XSS风险',
      path: [
        {
          function: 'parseJSON',
          file: 'src/api/handler.go',
          line: 78,
          variable: 'data',
          operation: 'assignment',
          taint_type: 'external_input',
        },
        {
          function: 'renderTemplate',
          file: 'src/views/render.go',
          line: 34,
          variable: 'content',
          operation: 'template_render',
          taint_type: 'output_operation',
        },
      ],
    },
  ];





  const handleTracePath = async () => {
    if (!sourceFunction || !sinkFunction) {
      setError('请选择源函数和汇聚函数');
      return;
    }

    setLoading(true);
    setError(null);
    
    try {
      const resp = await traceTaintPaths({
        source_function: sourceFunction,
        sink_function: sinkFunction,
        max_paths: Math.max(1, maxDepth || 10),
      });

      const convertedPaths: TaintPath[] = (resp.paths || []).map((seg, idx) => {
        const nodes = seg.nodes || [];
        const pathSegments: PathSegment[] = nodes.map(n => ({
          function: n.function_name,
          file: n.file_path || 'unknown',
          line: n.line_number || 0,
          variable: n.variable_name || '',
          operation: n.operation || '',
          taint_type: n.data_flow || '',
        }));
        const sourceFn = nodes[0]?.function_name || sourceFunction;
        const sinkFn = nodes[nodes.length - 1]?.function_name || sinkFunction;
        return {
          id: `${seg.path_index ?? idx}`,
          source: sourceFn,
          sink: sinkFn,
          path: pathSegments,
          riskLevel: seg.has_sanitizer ? 'low' : 'medium',
          confidence: seg.has_sanitizer ? 0.6 : 0.8,
          vulnerability_type: 'TaintFlow',
          description: seg.has_sanitizer ? '路径包含净化处理' : '从源到汇的污点路径',
        };
      });
      
      setTaintPaths(convertedPaths);
      setStatistics({
        total_paths: convertedPaths.length,
        high_risk_paths: convertedPaths.filter(p => p.riskLevel === 'high').length,
        sources_found: sources.length,
        sinks_found: sinks.length,
      });
      
    } catch (err) {
      console.error('Taint analysis failed:', err);
      setError(err instanceof Error ? err.message : '污点分析失败');
      
      // 使用模拟数据作为后备
      setTaintPaths(mockTaintPaths);
      setStatistics({
        total_paths: mockTaintPaths.length,
        high_risk_paths: mockTaintPaths.filter(p => p.riskLevel === 'high').length,
        sources_found: mockSources.length,
        sinks_found: mockSinks.length,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleQuerySources = useCallback(async () => {
    try {
      const resp = await queryTaintSources(sourcePattern || '', '');
      const mapped: SourceSink[] = (resp.sources || []).map(s => ({
        name: s.name,
        type: 'source',
        file: 'unknown',
        line: 0,
        description: s.description || '',
        category: s.type || 'Unknown',
        risk_level: 'medium',
      }));
      setSources(mapped);
    } catch (err) {
      console.error('Failed to query sources:', err);
      setSources(mockSources);
    }
  }, [sourcePattern]);

  const handleQuerySinks = useCallback(async () => {
    try {
      const resp = await queryTaintSinks(sinkPattern || '', '');
      const mapped: SourceSink[] = (resp.sinks || []).map(s => ({
        name: s.name,
        type: 'sink',
        file: 'unknown',
        line: 0,
        description: s.description || '',
        category: s.type || 'Unknown',
        risk_level: 'medium',
      }));
      setSinks(mapped);
    } catch (err) {
      console.error('Failed to query sinks:', err);
      setSinks(mockSinks);
    }
  }, [sinkPattern]);




  const renderFlowDiagram = useCallback((nodes: TaintFlowNode[], edges: TaintFlowEdge[]) => {
    if (!svgRef.current) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove();

    const width = 800;
    const height = 400;
    const margin = { top: 20, right: 20, bottom: 20, left: 20 };

    svg.attr("width", width).attr("height", height);

    const g = svg.append("g")
      .attr("transform", `translate(${margin.left},${margin.top})`);

    // 创建力导向图
    const simulation = d3.forceSimulation(nodes as any)
      .force("link", d3.forceLink(edges).id((d: any) => d.id).distance(100))
      .force("charge", d3.forceManyBody().strength(-300))
      .force("center", d3.forceCenter((width - margin.left - margin.right) / 2, (height - margin.top - margin.bottom) / 2));

    // 绘制边
    const link = g.append("g")
      .selectAll("line")
      .data(edges)
      .enter().append("line")
      .attr("stroke", "#999")
      .attr("stroke-opacity", 0.6)
      .attr("stroke-width", 2)
      .attr("marker-end", "url(#arrowhead)");

    // 添加箭头标记
    svg.append("defs").append("marker")
      .attr("id", "arrowhead")
      .attr("viewBox", "0 -5 10 10")
      .attr("refX", 15)
      .attr("refY", 0)
      .attr("markerWidth", 6)
      .attr("markerHeight", 6)
      .attr("orient", "auto")
      .append("path")
      .attr("d", "M0,-5L10,0L0,5")
      .attr("fill", "#999");

    // 绘制节点
    const node = g.append("g")
      .selectAll("circle")
      .data(nodes)
      .enter().append("circle")
      .attr("r", 20)
      .attr("fill", (d: TaintFlowNode) => {
        switch (d.type) {
          case 'source': return '#4caf50';
          case 'sink': return '#f44336';
          default: return '#2196f3';
        }
      })
      .attr("stroke", "#fff")
      .attr("stroke-width", 2)
      .call(d3.drag<any, any>()
        .on("start", (event, d: any) => {
          if (!event.active) simulation.alphaTarget(0.3).restart();
          d.fx = d.x;
          d.fy = d.y;
        })
        .on("drag", (event, d: any) => {
          d.fx = event.x;
          d.fy = event.y;
        })
        .on("end", (event, d: any) => {
          if (!event.active) simulation.alphaTarget(0);
          d.fx = null;
          d.fy = null;
        }));

    // 添加节点标签
    const label = g.append("g")
      .selectAll("text")
      .data(nodes)
      .enter().append("text")
      .text((d: TaintFlowNode) => d.label)
      .attr("font-size", "12px")
      .attr("text-anchor", "middle")
      .attr("dy", 4);

    // 添加边标签
    const edgeLabel = g.append("g")
      .selectAll("text")
      .data(edges)
      .enter().append("text")
      .text((d: TaintFlowEdge) => d.label)
      .attr("font-size", "10px")
      .attr("text-anchor", "middle")
      .attr("fill", "#666");

    simulation.on("tick", () => {
      link
        .attr("x1", (d: any) => d.source.x)
        .attr("y1", (d: any) => d.source.y)
        .attr("x2", (d: any) => d.target.x)
        .attr("y2", (d: any) => d.target.y);

      node
        .attr("cx", (d: any) => d.x)
        .attr("cy", (d: any) => d.y);

      label
        .attr("x", (d: any) => d.x)
        .attr("y", (d: any) => d.y);

      edgeLabel
        .attr("x", (d: any) => (d.source.x + d.target.x) / 2)
        .attr("y", (d: any) => (d.source.y + d.target.y) / 2);
    });
  }, []);

  // 在渲染函数之后定义，避免 use-before-define
  const generateFlowDiagram = useCallback((path: TaintPath) => {
    const nodes: TaintFlowNode[] = [];
    const edges: TaintFlowEdge[] = [];

    // 创建节点
    path.path.forEach((segment, index) => {
      const nodeId = `node-${index}`;
      const nodeType = index === 0 ? 'source' : 
                      index === path.path.length - 1 ? 'sink' : 'intermediate';
      
      nodes.push({
        id: nodeId,
        label: segment.function,
        type: nodeType,
        file: segment.file,
        line: segment.line,
        risk: path.riskLevel,
      });

      // 创建边
      if (index > 0) {
        edges.push({
          source: `node-${index - 1}`,
          target: nodeId,
          label: segment.operation,
          type: segment.taint_type,
        });
      }
    });

    setFlowData({ nodes, edges });
    renderFlowDiagram(nodes, edges);
  }, [renderFlowDiagram]);

  // 初始化时加载数据源和汇聚点（函数已在上方定义，避免 use-before-define）
  useEffect(() => {
    handleQuerySources();
    handleQuerySinks();
  }, [handleQuerySources, handleQuerySinks]);

  // 当选择路径时，生成数据流图
  useEffect(() => {
    if (selectedPath) {
      generateFlowDiagram(selectedPath);
    }
  }, [selectedPath, generateFlowDiagram]);

  const getRiskColor = (risk: string) => {
    switch (risk) {
      case 'high': return 'error';
      case 'medium': return 'warning';
      case 'low': return 'success';
      default: return 'default';
    }
  };

  const getCategoryColor = (category: string) => {
    const colors: { [key: string]: string } = {
      'HTTP Input': '#f44336',
      'Database': '#ff9800',
      'Command Execution': '#d32f2f',
      'Template': '#9c27b0',
      'Deserialization': '#3f51b5',
      'Environment': '#4caf50',
      'Logging': '#607d8b',
    };
    return colors[category] || '#757575';
  };

  const filteredPaths = taintPaths.filter(path => 
    filterRisk === 'all' || path.riskLevel === filterRisk
  );

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" gutterBottom>
          污点分析
        </Typography>
        <Box display="flex" gap={2}>
          <Chip
            icon={<Security />}
            label={`${statistics.total_paths} 条路径`}
            color="primary"
            variant="outlined"
          />
          <Chip
            icon={<BugReport />}
            label={`${statistics.high_risk_paths} 高危`}
            color="error"
            variant="outlined"
          />
        </Box>
      </Box>

      {error && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          {error} - 显示模拟数据
        </Alert>
      )}

      <Tabs value={tabValue} onChange={(_, newValue) => setTabValue(newValue)} sx={{ mb: 3 }}>
        <Tab label="路径追踪" icon={<Timeline />} />
        <Tab label="数据流图" icon={<AccountTree />} />
        <Tab label="源汇查询" icon={<Source />} />
        <Tab label="分析设置" icon={<Settings />} />
      </Tabs>

      {/* 路径追踪标签页 */}
      {tabValue === 0 && (
        <Grid container spacing={3}>
          <Grid size={{ xs: 12, md: 4 }}>
            <Paper elevation={2} sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                <Timeline sx={{ mr: 1, verticalAlign: 'middle' }} />
                路径追踪配置
              </Typography>
              
              <Box sx={{ mb: 2 }}>
                <TextField
                  fullWidth
                  label="源函数"
                  value={sourceFunction}
                  onChange={(e) => setSourceFunction(e.target.value)}
                  placeholder="例如: http.Request.FormValue"
                  sx={{ mb: 2 }}
                />
                <TextField
                  fullWidth
                  label="汇聚函数"
                  value={sinkFunction}
                  onChange={(e) => setSinkFunction(e.target.value)}
                  placeholder="例如: database.Query"
                  sx={{ mb: 2 }}
                />
                <TextField
                  fullWidth
                  label="项目路径"
                  value={projectPath}
                  onChange={(e) => setProjectPath(e.target.value)}
                  placeholder="可选：指定分析路径"
                  sx={{ mb: 2 }}
                />
                <Button
                  variant="contained"
                  startIcon={loading ? <CircularProgress size={20} /> : <PlayArrow />}
                  onClick={handleTracePath}
                  disabled={loading || !sourceFunction || !sinkFunction}
                  fullWidth
                >
                  {loading ? '分析中...' : '开始追踪'}
                </Button>
              </Box>

              {/* 过滤器 */}
              <Box sx={{ mt: 3 }}>
                <Typography variant="subtitle2" gutterBottom>
                  <FilterList sx={{ mr: 1, verticalAlign: 'middle' }} />
                  风险过滤
                </Typography>
                <FormControl fullWidth size="small">
                  <InputLabel>风险等级</InputLabel>
                  <Select
                    value={filterRisk}
                    onChange={(e) => setFilterRisk(e.target.value)}
                    label="风险等级"
                  >
                    <MenuItem value="all">全部</MenuItem>
                    <MenuItem value="high">高危</MenuItem>
                    <MenuItem value="medium">中危</MenuItem>
                    <MenuItem value="low">低危</MenuItem>
                  </Select>
                </FormControl>
              </Box>
            </Paper>
          </Grid>

          <Grid size={{ xs: 12, md: 8 }}>
            <Paper elevation={2} sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                发现的污点路径 ({filteredPaths.length})
              </Typography>
              
              {loading && <LinearProgress sx={{ mb: 2 }} />}
              
              {filteredPaths.length > 0 ? (
                filteredPaths.map((path) => (
                  <Card key={path.id} sx={{ mb: 2 }}>
                    <CardContent>
                      <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                        <Box>
                          <Box display="flex" alignItems="center" gap={1} mb={1}>
                            <Chip
                              label={path.riskLevel}
                              color={getRiskColor(path.riskLevel) as any}
                              size="small"
                            />
                            <Chip
                              label={path.vulnerability_type}
                              variant="outlined"
                              size="small"
                            />
                            <Chip
                              label={`置信度: ${(path.confidence * 100).toFixed(0)}%`}
                              variant="outlined"
                              size="small"
                            />
                          </Box>
                          <Typography variant="h6">
                            {path.source} → {path.sink}
                          </Typography>
                          <Typography variant="body2" color="textSecondary">
                            {path.description}
                          </Typography>
                        </Box>
                      </Box>
                      
                      <Accordion>
                        <AccordionSummary expandIcon={<ExpandMore />}>
                          <Typography>查看路径详情 ({path.path.length} 步)</Typography>
                        </AccordionSummary>
                        <AccordionDetails>
                          <List dense>
                            {path.path.map((segment, index) => (
                              <ListItem key={index}>
                                <ListItemText
                                  primary={`${index + 1}. ${segment.function} (${segment.file}:${segment.line})`}
                                  secondary={
                                    <Box>
                                      <Typography variant="body2">
                                        变量: {segment.variable}, 操作: {segment.operation}
                                      </Typography>
                                      <Chip
                                        label={segment.taint_type}
                                        size="small"
                                        variant="outlined"
                                        sx={{ mt: 0.5 }}
                                      />
                                    </Box>
                                  }
                                />
                              </ListItem>
                            ))}
                          </List>
                        </AccordionDetails>
                      </Accordion>
                    </CardContent>
                    <CardActions>
                      <Button
                        size="small"
                        startIcon={<Visibility />}
                        onClick={() => {
                          setSelectedPath(path);
                          setTabValue(1);
                        }}
                      >
                        查看数据流图
                      </Button>
                      <Button
                        size="small"
                        startIcon={<Download />}
                      >
                        导出报告
                      </Button>
                    </CardActions>
                  </Card>
                ))
              ) : (
                !loading && (
                  <Alert severity="info">
                    暂无发现污点路径。请检查源函数和汇聚函数配置。
                  </Alert>
                )
              )}
            </Paper>
          </Grid>
        </Grid>
      )}

      {/* 数据流图标签页 */}
      {tabValue === 1 && (
        <Grid container spacing={3}>
          <Grid size={12}>
            <Paper elevation={2} sx={{ p: 3 }}>
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                <Typography variant="h6">
                  <AccountTree sx={{ mr: 1, verticalAlign: 'middle' }} />
                  数据流图
                </Typography>
                <Box>
                  <IconButton size="small"><ZoomIn /></IconButton>
                  <IconButton size="small"><ZoomOut /></IconButton>
                  <IconButton size="small"><CenterFocusStrong /></IconButton>
                </Box>
              </Box>
              
              {selectedPath ? (
                <Box>
                  <Alert severity="info" sx={{ mb: 2 }}>
                    当前显示路径: {selectedPath.source} → {selectedPath.sink}
                  </Alert>
                  <Box sx={{ border: '1px solid #ddd', borderRadius: 1, overflow: 'hidden' }}>
                    <svg ref={svgRef}></svg>
                  </Box>
                </Box>
              ) : (
                <Alert severity="info">
                  请先在"路径追踪"标签页中选择一个路径来查看数据流图。
                </Alert>
              )}
            </Paper>
          </Grid>
        </Grid>
      )}

      {/* 源汇查询标签页 */}
      {tabValue === 2 && (
        <Grid container spacing={3}>
          <Grid size={{ xs: 12, md: 6 }}>
            <Paper elevation={2} sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                <Source sx={{ mr: 1, verticalAlign: 'middle' }} />
                数据源查询
              </Typography>
              
              <Box display="flex" gap={1} sx={{ mb: 2 }}>
                <TextField
                  fullWidth
                  label="搜索模式"
                  value={sourcePattern}
                  onChange={(e) => setSourcePattern(e.target.value)}
                  placeholder="例如: http, json"
                  size="small"
                />
                <Button
                  variant="outlined"
                  onClick={handleQuerySources}
                  disabled={loading}
                >
                  <Refresh />
                </Button>
              </Box>

              {sources.length > 0 && (
                <List dense>
                  {sources.map((source, index) => (
                    <React.Fragment key={index}>
                      <ListItem>
                        <ListItemText
                          primary={
                            <Box display="flex" alignItems="center" gap={1}>
                              <Typography variant="body1">{source.name}</Typography>
                              <Chip
                                label={source.category}
                                size="small"
                                sx={{ backgroundColor: getCategoryColor(source.category), color: 'white' }}
                              />
                            </Box>
                          }
                          secondary={
                            <Box>
                              <Typography variant="body2">
                                {source.file}:{source.line}
                              </Typography>
                              <Typography variant="caption" color="textSecondary">
                                {source.description}
                              </Typography>
                            </Box>
                          }
                        />
                        <Tooltip title="选择为源函数">
                          <IconButton
                            size="small"
                            onClick={() => setSourceFunction(source.name)}
                          >
                            <CallSplit />
                          </IconButton>
                        </Tooltip>
                      </ListItem>
                      {index < sources.length - 1 && <Divider />}
                    </React.Fragment>
                  ))}
                </List>
              )}
            </Paper>
          </Grid>

          <Grid size={{ xs: 12, md: 6 }}>
            <Paper elevation={2} sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                <CallSplit sx={{ mr: 1, verticalAlign: 'middle' }} />
                汇聚点查询
              </Typography>
              
              <Box display="flex" gap={1} sx={{ mb: 2 }}>
                <TextField
                  fullWidth
                  label="搜索模式"
                  value={sinkPattern}
                  onChange={(e) => setSinkPattern(e.target.value)}
                  placeholder="例如: database, exec"
                  size="small"
                />
                <Button
                  variant="outlined"
                  onClick={handleQuerySinks}
                  disabled={loading}
                >
                  <Refresh />
                </Button>
              </Box>

              {sinks.length > 0 && (
                <List dense>
                  {sinks.map((sink, index) => (
                    <React.Fragment key={index}>
                      <ListItem>
                        <ListItemText
                          primary={
                            <Box display="flex" alignItems="center" gap={1}>
                              <Typography variant="body1">{sink.name}</Typography>
                              <Chip
                                label={sink.category}
                                size="small"
                                sx={{ backgroundColor: getCategoryColor(sink.category), color: 'white' }}
                              />
                            </Box>
                          }
                          secondary={
                            <Box>
                              <Typography variant="body2">
                                {sink.file}:{sink.line}
                              </Typography>
                              <Typography variant="caption" color="textSecondary">
                                {sink.description}
                              </Typography>
                            </Box>
                          }
                        />
                        <Tooltip title="选择为汇聚函数">
                          <IconButton
                            size="small"
                            onClick={() => setSinkFunction(sink.name)}
                          >
                            <CallSplit />
                          </IconButton>
                        </Tooltip>
                      </ListItem>
                      {index < sinks.length - 1 && <Divider />}
                    </React.Fragment>
                  ))}
                </List>
              )}
            </Paper>
          </Grid>
        </Grid>
      )}

      {/* 分析设置标签页 */}
      {tabValue === 3 && (
        <Grid container spacing={3}>
          <Grid size={{ xs: 12, md: 6 }}>
            <Paper elevation={2} sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                <Settings sx={{ mr: 1, verticalAlign: 'middle' }} />
                分析参数
              </Typography>
              
              <Box sx={{ mb: 2 }}>
                <Typography variant="body2" gutterBottom>
                  最大分析深度
                </Typography>
                <TextField
                  type="number"
                  fullWidth
                  value={maxDepth}
                  onChange={(e) => setMaxDepth(parseInt(e.target.value) || 10)}
                  inputProps={{ min: 1, max: 50 }}
                  size="small"
                />
              </Box>

              <FormControlLabel
                control={
                  <Switch
                    checked={includeSanitizers}
                    onChange={(e) => setIncludeSanitizers(e.target.checked)}
                  />
                }
                label="包含净化函数分析"
              />
            </Paper>
          </Grid>

          <Grid size={{ xs: 12, md: 6 }}>
            <Paper elevation={2} sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                分析统计
              </Typography>
              
              <Grid container spacing={2}>
                <Grid size={6}>
                  <Box textAlign="center">
                    <Typography variant="h4" color="primary">
                      {statistics.total_paths}
                    </Typography>
                    <Typography variant="body2">总路径数</Typography>
                  </Box>
                </Grid>
                <Grid size={6}>
                  <Box textAlign="center">
                    <Typography variant="h4" color="error">
                      {statistics.high_risk_paths}
                    </Typography>
                    <Typography variant="body2">高危路径</Typography>
                  </Box>
                </Grid>
                <Grid size={6}>
                  <Box textAlign="center">
                    <Typography variant="h4" color="success">
                      {statistics.sources_found}
                    </Typography>
                    <Typography variant="body2">数据源</Typography>
                  </Box>
                </Grid>
                <Grid size={6}>
                  <Box textAlign="center">
                    <Typography variant="h4" color="warning">
                      {statistics.sinks_found}
                    </Typography>
                    <Typography variant="body2">汇聚点</Typography>
                  </Box>
                </Grid>
              </Grid>
            </Paper>
          </Grid>
        </Grid>
      )}

      {/* 使用说明 */}
      <Grid container spacing={3} sx={{ mt: 2 }}>
        <Grid size={12}>
          <Alert severity="info" icon={<Info />}>
            <Typography variant="body2">
              <strong>使用说明：</strong>
              污点分析用于追踪数据从不可信源（如用户输入）到敏感操作（如数据库查询）的流动路径。
              在"路径追踪"中配置源函数和汇聚函数，在"数据流图"中可视化分析结果，
              在"源汇查询"中浏览可用的数据源和汇聚点，在"分析设置"中调整分析参数。
            </Typography>
          </Alert>
        </Grid>
      </Grid>
    </Box>
  );
};

export default TaintAnalysis;