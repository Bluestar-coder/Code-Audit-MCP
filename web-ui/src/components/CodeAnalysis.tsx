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
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Tab,
  Tabs,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';
import {
  PlayArrow,
  ExpandMore,
  Code,
  BugReport,
  Security,
  Speed,
  Folder,
  Description,
} from '@mui/icons-material';

interface AnalysisResult {
  id: string;
  file: string;
  language: string;
  issues: Issue[];
  metrics: CodeMetrics;
  timestamp: string;
}

interface Issue {
  type: 'vulnerability' | 'bug' | 'code_smell' | 'security_hotspot';
  severity: 'critical' | 'major' | 'minor' | 'info';
  rule: string;
  message: string;
  line: number;
  column?: number;
}

interface CodeMetrics {
  linesOfCode: number;
  complexity: number;
  duplicatedLines: number;
  coverage: number;
  maintainabilityIndex: number;
}

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ p: 3 }}>
          {children}
        </Box>
      )}
    </div>
  );
}

const CodeAnalysis: React.FC = () => {
  const [projectPath, setProjectPath] = useState('');
  const [language, setLanguage] = useState('go');
  const [analysisType, setAnalysisType] = useState('full');
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<AnalysisResult[]>([]);
  const [tabValue, setTabValue] = useState(0);

  // 模拟分析结果
  const mockResults: AnalysisResult[] = [
    {
      id: '1',
      file: 'src/handlers/user.go',
      language: 'go',
      timestamp: new Date().toISOString(),
      issues: [
        {
          type: 'vulnerability',
          severity: 'critical',
          rule: 'sql-injection',
          message: '潜在的SQL注入漏洞：直接拼接用户输入到SQL查询中',
          line: 45,
          column: 12,
        },
        {
          type: 'security_hotspot',
          severity: 'major',
          rule: 'hardcoded-credentials',
          message: '硬编码的敏感信息',
          line: 23,
        },
        {
          type: 'code_smell',
          severity: 'minor',
          rule: 'function-complexity',
          message: '函数复杂度过高，建议重构',
          line: 78,
        },
      ],
      metrics: {
        linesOfCode: 234,
        complexity: 15,
        duplicatedLines: 12,
        coverage: 78.5,
        maintainabilityIndex: 65,
      },
    },
    {
      id: '2',
      file: 'src/database/connection.go',
      language: 'go',
      timestamp: new Date().toISOString(),
      issues: [
        {
          type: 'bug',
          severity: 'major',
          rule: 'resource-leak',
          message: '数据库连接可能未正确关闭',
          line: 67,
        },
        {
          type: 'code_smell',
          severity: 'minor',
          rule: 'unused-variable',
          message: '未使用的变量',
          line: 34,
        },
      ],
      metrics: {
        linesOfCode: 156,
        complexity: 8,
        duplicatedLines: 0,
        coverage: 85.2,
        maintainabilityIndex: 72,
      },
    },
  ];

  const handleAnalyze = async () => {
    if (!projectPath) {
      return;
    }

    setLoading(true);
    // 模拟分析过程
    setTimeout(() => {
      setResults(mockResults);
      setLoading(false);
    }, 3000);
  };

  const getIssueTypeIcon = (type: string) => {
    switch (type) {
      case 'vulnerability': return <Security color="error" />;
      case 'bug': return <BugReport color="warning" />;
      case 'security_hotspot': return <Security color="warning" />;
      case 'code_smell': return <Code color="info" />;
      default: return <Code />;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'error';
      case 'major': return 'warning';
      case 'minor': return 'info';
      case 'info': return 'default';
      default: return 'default';
    }
  };

  const getMetricColor = (value: number, type: string) => {
    switch (type) {
      case 'complexity':
        return value > 10 ? 'error' : value > 5 ? 'warning' : 'success';
      case 'coverage':
        return value > 80 ? 'success' : value > 60 ? 'warning' : 'error';
      case 'maintainability':
        return value > 70 ? 'success' : value > 50 ? 'warning' : 'error';
      default:
        return 'default';
    }
  };

  const totalIssues = results.reduce((sum, result) => sum + result.issues.length, 0);
  const criticalIssues = results.reduce((sum, result) => 
    sum + result.issues.filter(issue => issue.severity === 'critical').length, 0);

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        代码分析
      </Typography>

      {/* 分析配置 */}
      <Paper elevation={2} sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          分析配置
        </Typography>
        
        <Grid container spacing={2} alignItems="center">
          <Grid size={{ xs: 12, md: 4 }}>
            <TextField
              fullWidth
              label="项目路径"
              value={projectPath}
              onChange={(e) => setProjectPath(e.target.value)}
              placeholder="/path/to/your/project"
            />
          </Grid>
          
          <Grid size={{ xs: 12, md: 2 }}>
            <FormControl fullWidth>
              <InputLabel>语言</InputLabel>
              <Select
                value={language}
                label="语言"
                onChange={(e) => setLanguage(e.target.value)}
              >
                <MenuItem value="go">Go</MenuItem>
                <MenuItem value="java">Java</MenuItem>
                <MenuItem value="python">Python</MenuItem>
                <MenuItem value="javascript">JavaScript</MenuItem>
                <MenuItem value="typescript">TypeScript</MenuItem>
              </Select>
            </FormControl>
          </Grid>

          <Grid size={{ xs: 12, md: 2 }}>
            <FormControl fullWidth>
              <InputLabel>分析类型</InputLabel>
              <Select
                value={analysisType}
                label="分析类型"
                onChange={(e) => setAnalysisType(e.target.value)}
              >
                <MenuItem value="full">完整分析</MenuItem>
                <MenuItem value="security">安全分析</MenuItem>
                <MenuItem value="quality">代码质量</MenuItem>
                <MenuItem value="performance">性能分析</MenuItem>
              </Select>
            </FormControl>
          </Grid>

          <Grid size={{ xs: 12, md: 2 }}>
            <Button
              variant="contained"
              startIcon={loading ? <CircularProgress size={20} /> : <PlayArrow />}
              onClick={handleAnalyze}
              disabled={loading || !projectPath}
              fullWidth
            >
              {loading ? '分析中...' : '开始分析'}
            </Button>
          </Grid>
        </Grid>
      </Paper>

      {/* 分析结果概览 */}
      {results.length > 0 && (
        <Grid container spacing={3} sx={{ mb: 3 }}>
          <Grid size={{ xs: 12, sm: 6, md: 3 }}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      总问题数
                    </Typography>
                    <Typography variant="h4">
                      {totalIssues}
                    </Typography>
                  </Box>
                  <BugReport color="primary" sx={{ fontSize: 40 }} />
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid size={{ xs: 12, sm: 6, md: 3 }}>
             <Card>
               <CardContent>
                 <Box display="flex" alignItems="center" justifyContent="space-between">
                   <Box>
                     <Typography color="textSecondary" gutterBottom>
                       严重问题
                     </Typography>
                     <Typography variant="h4" color="error">
                       {criticalIssues}
                     </Typography>
                   </Box>
                   <Security color="error" sx={{ fontSize: 40 }} />
                 </Box>
               </CardContent>
             </Card>
           </Grid>

           <Grid size={{ xs: 12, sm: 6, md: 3 }}>
             <Card>
               <CardContent>
                 <Box display="flex" alignItems="center" justifyContent="space-between">
                   <Box>
                     <Typography color="textSecondary" gutterBottom>
                       分析文件
                     </Typography>
                     <Typography variant="h4">
                       {results.length}
                     </Typography>
                   </Box>
                   <Description color="info" sx={{ fontSize: 40 }} />
                 </Box>
               </CardContent>
             </Card>
           </Grid>

           <Grid size={{ xs: 12, sm: 6, md: 3 }}>
             <Card>
               <CardContent>
                 <Box display="flex" alignItems="center" justifyContent="space-between">
                   <Box>
                     <Typography color="textSecondary" gutterBottom>
                       平均质量
                     </Typography>
                     <Typography variant="h4" color="success.main">
                       B+
                     </Typography>
                   </Box>
                   <Speed color="success" sx={{ fontSize: 40 }} />
                 </Box>
               </CardContent>
             </Card>
           </Grid>
        </Grid>
      )}

      {/* 详细结果 */}
      {results.length > 0 && (
        <Paper elevation={2}>
          <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
            <Tabs value={tabValue} onChange={(e, newValue) => setTabValue(newValue)}>
              <Tab label="问题列表" />
              <Tab label="代码指标" />
              <Tab label="文件详情" />
            </Tabs>
          </Box>

          <TabPanel value={tabValue} index={0}>
            {/* 问题列表 */}
            <List>
              {results.map((result) =>
                result.issues.map((issue, index) => (
                  <React.Fragment key={`${result.id}-${index}`}>
                    <ListItem>
                      <Box display="flex" alignItems="center" gap={2} width="100%">
                        {getIssueTypeIcon(issue.type)}
                        <Box flexGrow={1}>
                          <Typography variant="body1">
                            {issue.message}
                          </Typography>
                          <Typography variant="body2" color="textSecondary">
                            {result.file}:{issue.line}
                            {issue.column && `:${issue.column}`}
                          </Typography>
                        </Box>
                        <Box display="flex" gap={1}>
                          <Chip
                            label={issue.severity}
                            color={getSeverityColor(issue.severity) as any}
                            size="small"
                          />
                          <Chip
                            label={issue.rule}
                            variant="outlined"
                            size="small"
                          />
                        </Box>
                      </Box>
                    </ListItem>
                    <Divider />
                  </React.Fragment>
                ))
              )}
            </List>
          </TabPanel>

          <TabPanel value={tabValue} index={1}>
            {/* 代码指标 */}
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>文件</TableCell>
                    <TableCell>代码行数</TableCell>
                    <TableCell>复杂度</TableCell>
                    <TableCell>重复行</TableCell>
                    <TableCell>覆盖率</TableCell>
                    <TableCell>可维护性</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {results.map((result) => (
                    <TableRow key={result.id}>
                      <TableCell>{result.file}</TableCell>
                      <TableCell>{result.metrics.linesOfCode}</TableCell>
                      <TableCell>
                        <Chip
                          label={result.metrics.complexity}
                          color={getMetricColor(result.metrics.complexity, 'complexity') as any}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>{result.metrics.duplicatedLines}</TableCell>
                      <TableCell>
                        <Chip
                          label={`${result.metrics.coverage}%`}
                          color={getMetricColor(result.metrics.coverage, 'coverage') as any}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={result.metrics.maintainabilityIndex}
                          color={getMetricColor(result.metrics.maintainabilityIndex, 'maintainability') as any}
                          size="small"
                        />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </TabPanel>

          <TabPanel value={tabValue} index={2}>
            {/* 文件详情 */}
            {results.map((result) => (
              <Accordion key={result.id}>
                <AccordionSummary expandIcon={<ExpandMore />}>
                  <Box display="flex" alignItems="center" gap={2}>
                    <Folder />
                    <Typography>{result.file}</Typography>
                    <Chip
                      label={`${result.issues.length} 问题`}
                      color={result.issues.length > 0 ? 'warning' : 'success'}
                      size="small"
                    />
                  </Box>
                </AccordionSummary>
                <AccordionDetails>
                  <Grid container spacing={2}>
                    <Grid size={{ xs: 12, md: 6 }}>
                      <Typography variant="h6" gutterBottom>
                        问题详情
                      </Typography>
                      <List dense>
                        {result.issues.map((issue, index) => (
                          <ListItem key={index}>
                            <ListItemText
                              primary={issue.message}
                              secondary={`行 ${issue.line} - ${issue.rule}`}
                            />
                            <Chip
                              label={issue.severity}
                              color={getSeverityColor(issue.severity) as any}
                              size="small"
                            />
                          </ListItem>
                        ))}
                      </List>
                    </Grid>
                    <Grid size={{ xs: 12, md: 6 }}>
                      <Typography variant="h6" gutterBottom>
                        代码指标
                      </Typography>
                      <Box>
                        <Typography>代码行数: {result.metrics.linesOfCode}</Typography>
                        <Typography>复杂度: {result.metrics.complexity}</Typography>
                        <Typography>重复行: {result.metrics.duplicatedLines}</Typography>
                        <Typography>覆盖率: {result.metrics.coverage}%</Typography>
                        <Typography>可维护性: {result.metrics.maintainabilityIndex}</Typography>
                      </Box>
                    </Grid>
                  </Grid>
                </AccordionDetails>
              </Accordion>
            ))}
          </TabPanel>
        </Paper>
      )}

      {/* 使用说明 */}
      {results.length === 0 && !loading && (
        <Alert severity="info">
          <Typography variant="body2">
            <strong>使用说明：</strong>
            输入项目路径，选择编程语言和分析类型，然后点击"开始分析"来检测代码中的安全漏洞、
            代码质量问题和性能问题。分析结果将显示详细的问题列表、代码指标和修复建议。
          </Typography>
        </Alert>
      )}
    </Box>
  );
};

export default CodeAnalysis;