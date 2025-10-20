import React, { useState } from 'react';
import Grid from '@mui/material/Grid';
import {
  Box,
  Typography,
  Paper,
  TextField,
  Button,
  Card,
  CardContent,
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
} from '@mui/material';
import {
  PlayArrow,
  ExpandMore,
  Timeline,
  Source,
  CallSplit,
  Info,
  Refresh,
} from '@mui/icons-material';

interface TaintPath {
  id: string;
  source: string;
  sink: string;
  path: PathSegment[];
  riskLevel: 'high' | 'medium' | 'low';
}

interface PathSegment {
  function: string;
  file: string;
  line: number;
  variable: string;
  operation: string;
}

interface SourceSink {
  name: string;
  type: 'source' | 'sink';
  file: string;
  line: number;
  description: string;
}

const TaintAnalysis: React.FC = () => {
  const [sourceFunction, setSourceFunction] = useState('');
  const [sinkFunction, setSinkFunction] = useState('');
  const [loading, setLoading] = useState(false);
  const [taintPaths, setTaintPaths] = useState<TaintPath[]>([]);
  const [sources, setSources] = useState<SourceSink[]>([]);
  const [sinks, setSinks] = useState<SourceSink[]>([]);
  const [sourcePattern, setSourcePattern] = useState('');
  const [sinkPattern, setSinkPattern] = useState('');

  // 模拟数据
  const mockSources: SourceSink[] = [
    {
      name: 'http.Request.FormValue',
      type: 'source',
      file: 'src/handlers/user.go',
      line: 25,
      description: 'HTTP表单输入，用户可控数据源',
    },
    {
      name: 'http.Request.URL.Query',
      type: 'source',
      file: 'src/handlers/search.go',
      line: 42,
      description: 'URL查询参数，用户可控数据源',
    },
    {
      name: 'json.Unmarshal',
      type: 'source',
      file: 'src/api/handler.go',
      line: 78,
      description: 'JSON反序列化，外部数据源',
    },
  ];

  const mockSinks: SourceSink[] = [
    {
      name: 'database.Query',
      type: 'sink',
      file: 'src/database/user.go',
      line: 156,
      description: 'SQL查询执行，潜在SQL注入点',
    },
    {
      name: 'os.Exec',
      type: 'sink',
      file: 'src/utils/system.go',
      line: 89,
      description: '系统命令执行，潜在命令注入点',
    },
    {
      name: 'template.Execute',
      type: 'sink',
      file: 'src/views/render.go',
      line: 34,
      description: '模板渲染，潜在XSS注入点',
    },
  ];

  const mockTaintPaths: TaintPath[] = [
    {
      id: '1',
      source: 'http.Request.FormValue',
      sink: 'database.Query',
      riskLevel: 'high',
      path: [
        {
          function: 'getUserInput',
          file: 'src/handlers/user.go',
          line: 25,
          variable: 'username',
          operation: 'assignment',
        },
        {
          function: 'validateUser',
          file: 'src/services/auth.go',
          line: 67,
          variable: 'userParam',
          operation: 'parameter_pass',
        },
        {
          function: 'queryUser',
          file: 'src/database/user.go',
          line: 156,
          variable: 'query',
          operation: 'string_concat',
        },
      ],
    },
    {
      id: '2',
      source: 'http.Request.URL.Query',
      sink: 'os.Exec',
      riskLevel: 'high',
      path: [
        {
          function: 'handleCommand',
          file: 'src/handlers/system.go',
          line: 42,
          variable: 'cmd',
          operation: 'assignment',
        },
        {
          function: 'executeCommand',
          file: 'src/utils/system.go',
          line: 89,
          variable: 'command',
          operation: 'parameter_pass',
        },
      ],
    },
  ];

  const handleTracePath = async () => {
    if (!sourceFunction || !sinkFunction) {
      return;
    }

    setLoading(true);
    // 模拟API调用
    setTimeout(() => {
      setTaintPaths(mockTaintPaths);
      setLoading(false);
    }, 2000);
  };

  const handleQuerySources = async () => {
    setLoading(true);
    setTimeout(() => {
      setSources(mockSources.filter(s => 
        !sourcePattern || s.name.toLowerCase().includes(sourcePattern.toLowerCase())
      ));
      setLoading(false);
    }, 1000);
  };

  const handleQuerySinks = async () => {
    setLoading(true);
    setTimeout(() => {
      setSinks(mockSinks.filter(s => 
        !sinkPattern || s.name.toLowerCase().includes(sinkPattern.toLowerCase())
      ));
      setLoading(false);
    }, 1000);
  };

  const getRiskColor = (risk: string) => {
    switch (risk) {
      case 'high': return 'error';
      case 'medium': return 'warning';
      case 'low': return 'success';
      default: return 'default';
    }
  };

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        污点分析
      </Typography>

      <Grid container spacing={3}>
        {/* 路径追踪 */}
        <Grid item xs={12} md={6}>
          <Paper elevation={2} sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              <Timeline sx={{ mr: 1, verticalAlign: 'middle' }} />
              路径追踪
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

            {taintPaths.length > 0 && (
              <Box>
                <Typography variant="subtitle1" gutterBottom>
                  发现的污点路径：
                </Typography>
                {taintPaths.map((path) => (
                  <Accordion key={path.id} sx={{ mb: 1 }}>
                    <AccordionSummary expandIcon={<ExpandMore />}>
                      <Box display="flex" alignItems="center" gap={1}>
                        <Chip
                          label={path.riskLevel}
                          color={getRiskColor(path.riskLevel) as any}
                          size="small"
                        />
                        <Typography>
                          {path.source} → {path.sink}
                        </Typography>
                      </Box>
                    </AccordionSummary>
                    <AccordionDetails>
                      <List dense>
                        {path.path.map((segment, index) => (
                          <ListItem key={index}>
                            <ListItemText
                              primary={`${segment.function} (${segment.file}:${segment.line})`}
                              secondary={`变量: ${segment.variable}, 操作: ${segment.operation}`}
                            />
                          </ListItem>
                        ))}
                      </List>
                    </AccordionDetails>
                  </Accordion>
                ))}
              </Box>
            )}
          </Paper>
        </Grid>

        {/* 源和汇聚查询 */}
        <Grid item xs={12} md={6}>
          <Grid container spacing={2}>
            {/* 数据源查询 */}
            <Grid item xs={12}>
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
                    placeholder="例如: http"
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
                            primary={source.name}
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

            {/* 汇聚点查询 */}
            <Grid item xs={12}>
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
                    placeholder="例如: database"
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
                            primary={sink.name}
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
        </Grid>

        {/* 使用说明 */}
        <Grid item xs={12}>
          <Alert severity="info" icon={<Info />}>
            <Typography variant="body2">
              <strong>使用说明：</strong>
              污点分析用于追踪数据从不可信源（如用户输入）到敏感操作（如数据库查询）的流动路径。
              选择源函数和汇聚函数，点击"开始追踪"来分析潜在的安全风险路径。
            </Typography>
          </Alert>
        </Grid>
      </Grid>
    </Box>
  );
};

export default TaintAnalysis;