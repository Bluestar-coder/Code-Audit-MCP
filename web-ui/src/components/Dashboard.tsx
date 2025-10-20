import React, { useState, useEffect } from 'react';
import Grid from '@mui/material/Grid';
import {
  Card,
  CardContent,
  Typography,
  Box,
  Paper,
  Chip,
  LinearProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Alert,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  Security,
  BugReport,
  Warning,
  CheckCircle,
  Code,
  Timeline,
  Refresh,
  History,
  Language,
  Folder,
  Schedule,
  TrendingUp,
  TrendingDown,
} from '@mui/icons-material';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, Legend, ResponsiveContainer, LineChart, Line } from 'recharts';
import { getDashboardData, getProjectStats, getVulnerabilityStats, getScanHistory, type DashboardData, type ProjectStats, type VulnerabilityStats, type ScanHistory } from '../api/client';

const Dashboard: React.FC = () => {
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());

  const loadDashboardData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // 尝试获取完整的仪表板数据
      try {
        const data = await getDashboardData();
        setDashboardData(data);
      } catch (err) {
        // 如果完整接口不可用，分别获取各部分数据
        console.warn('Dashboard API not available, falling back to individual APIs');
        const [projectStats, vulnStats, scanHistory] = await Promise.all([
          getProjectStats().catch(() => ({
            total_files: 0,
            total_lines: 0,
            total_functions: 0,
            total_classes: 0,
            languages: {},
          })),
          getVulnerabilityStats().catch(() => ({
            total: 0,
            critical: 0,
            high: 0,
            medium: 0,
            low: 0,
            fixed: 0,
            by_category: {},
          })),
          getScanHistory().catch(() => []),
        ]);

        setDashboardData({
          project_stats: projectStats,
          vulnerability_stats: vulnStats,
          scan_history: scanHistory,
          trend_data: [],
        });
      }
      
      setLastRefresh(new Date());
    } catch (err) {
      console.error('Failed to load dashboard data:', err);
      setError(err instanceof Error ? err.message : '加载数据失败');
      
      // 设置模拟数据作为后备
      setDashboardData({
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
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDashboardData();
  }, []);

  const handleRefresh = () => {
    loadDashboardData();
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}min`;
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString('zh-CN');
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'success';
      case 'failed': return 'error';
      case 'running': return 'warning';
      default: return 'default';
    }
  };

  const StatCard: React.FC<{
    title: string;
    value: number | string;
    icon: React.ReactNode;
    color: string;
    subtitle?: string;
    trend?: 'up' | 'down' | 'stable';
  }> = ({ title, value, icon, color, subtitle, trend }) => (
    <Card elevation={2} sx={{ height: '100%' }}>
      <CardContent>
        <Box display="flex" alignItems="center" justifyContent="space-between">
          <Box>
            <Typography color="textSecondary" gutterBottom variant="body2">
              {title}
            </Typography>
            <Box display="flex" alignItems="center" gap={1}>
              <Typography variant="h4" component="div" color={color}>
                {value}
              </Typography>
              {trend && (
                <Box sx={{ color: trend === 'up' ? '#f44336' : trend === 'down' ? '#4caf50' : '#757575' }}>
                  {trend === 'up' && <TrendingUp fontSize="small" />}
                  {trend === 'down' && <TrendingDown fontSize="small" />}
                </Box>
              )}
            </Box>
            {subtitle && (
              <Typography variant="body2" color="textSecondary">
                {subtitle}
              </Typography>
            )}
          </Box>
          <Box sx={{ color: color, fontSize: 40 }}>
            {icon}
          </Box>
        </Box>
      </CardContent>
    </Card>
  );

  if (loading) {
    return (
      <Box sx={{ width: '100%', mt: 2 }}>
        <LinearProgress />
        <Typography variant="h6" sx={{ mt: 2, textAlign: 'center' }}>
          正在加载仪表板数据...
        </Typography>
      </Box>
    );
  }

  if (!dashboardData) {
    return (
      <Box sx={{ mt: 2 }}>
        <Alert severity="error">
          无法加载仪表板数据
        </Alert>
      </Box>
    );
  }

  const { project_stats, vulnerability_stats, scan_history, trend_data } = dashboardData;

  const vulnerabilityData = [
    { name: '严重', value: vulnerability_stats.critical, color: '#f44336' },
    { name: '高危', value: vulnerability_stats.high, color: '#ff9800' },
    { name: '中危', value: vulnerability_stats.medium, color: '#ffeb3b' },
    { name: '低危', value: vulnerability_stats.low, color: '#4caf50' },
  ];

  const categoryData = Object.entries(vulnerability_stats.by_category).map(([name, value]) => ({
    name,
    value,
  }));

  const languageData = Object.entries(project_stats.languages).map(([name, value]) => ({
    name,
    value,
    percentage: ((value / project_stats.total_files) * 100).toFixed(1),
  }));

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" gutterBottom>
          安全分析仪表板
        </Typography>
        <Box display="flex" alignItems="center" gap={2}>
          <Typography variant="body2" color="textSecondary">
            最后更新: {formatTimestamp(lastRefresh.toISOString())}
          </Typography>
          <Tooltip title="刷新数据">
            <IconButton onClick={handleRefresh} disabled={loading}>
              <Refresh />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      {error && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          {error} - 显示模拟数据
        </Alert>
      )}
      
      {/* 统计卡片 */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard
            title="总漏洞数"
            value={vulnerability_stats.total}
            icon={<BugReport />}
            color="#f44336"
            subtitle={`${vulnerability_stats.fixed} 个已修复`}
            trend={vulnerability_stats.total > 40 ? 'up' : 'down'}
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard
            title="严重漏洞"
            value={vulnerability_stats.critical}
            icon={<Warning />}
            color="#d32f2f"
            trend={vulnerability_stats.critical > 0 ? 'up' : 'down'}
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard
            title="已分析文件"
            value={project_stats.total_files}
            icon={<Code />}
            color="#1976d2"
            subtitle={`${project_stats.total_functions} 个函数`}
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard
            title="代码行数"
            value={project_stats.total_lines.toLocaleString()}
            icon={<Timeline />}
            color="#388e3c"
            subtitle={`${project_stats.total_classes} 个类`}
          />
        </Grid>
      </Grid>

      {/* 图表区域 */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Paper elevation={2} sx={{ p: 2, height: 400 }}>
            <Typography variant="h6" gutterBottom>
              漏洞严重程度分布
            </Typography>
            <ResponsiveContainer width="100%" height="85%">
              <PieChart>
                <Pie
                  data={vulnerabilityData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={(entry: any) => `${entry.name} ${entry.value}`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {vulnerabilityData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <RechartsTooltip />
              </PieChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <Paper elevation={2} sx={{ p: 2, height: 400 }}>
            <Typography variant="h6" gutterBottom>
              漏洞类别分布
            </Typography>
            <ResponsiveContainer width="100%" height="85%">
              <BarChart data={categoryData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <RechartsTooltip />
                <Bar dataKey="value" fill="#8884d8" />
              </BarChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>
      </Grid>

      {/* 趋势图和语言分布 */}
      {trend_data.length > 0 && (
        <Grid container spacing={3} sx={{ mb: 3 }}>
          <Grid size={{ xs: 12, md: 8 }}>
            <Paper elevation={2} sx={{ p: 2, height: 350 }}>
              <Typography variant="h6" gutterBottom>
                漏洞趋势
              </Typography>
              <ResponsiveContainer width="100%" height="85%">
                <LineChart data={trend_data}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" />
                  <YAxis />
                  <RechartsTooltip />
                  <Legend />
                  <Line type="monotone" dataKey="vulnerabilities" stroke="#8884d8" name="漏洞数" />
                  <Line type="monotone" dataKey="fixed" stroke="#82ca9d" name="已修复" />
                </LineChart>
              </ResponsiveContainer>
            </Paper>
          </Grid>
          <Grid size={{ xs: 12, md: 4 }}>
            <Paper elevation={2} sx={{ p: 2, height: 350 }}>
              <Typography variant="h6" gutterBottom>
                编程语言分布
              </Typography>
              <List dense>
                {languageData.map((lang, index) => (
                  <ListItem key={index}>
                    <ListItemIcon>
                      <Language />
                    </ListItemIcon>
                    <ListItemText
                      primary={lang.name}
                      secondary={`${lang.value} 文件 (${lang.percentage}%)`}
                    />
                  </ListItem>
                ))}
              </List>
            </Paper>
          </Grid>
        </Grid>
      )}

      {/* 扫描历史 */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid size={12}>
          <Paper elevation={2} sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom display="flex" alignItems="center" gap={1}>
              <History />
              扫描历史
            </Typography>
            <Divider sx={{ mb: 2 }} />
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>时间</TableCell>
                    <TableCell>项目路径</TableCell>
                    <TableCell align="right">扫描文件</TableCell>
                    <TableCell align="right">发现漏洞</TableCell>
                    <TableCell align="right">耗时</TableCell>
                    <TableCell>状态</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {scan_history.slice(0, 5).map((scan) => (
                    <TableRow key={scan.id}>
                      <TableCell>
                        <Typography variant="body2">
                          {formatTimestamp(scan.timestamp)}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Box display="flex" alignItems="center" gap={1}>
                          <Folder fontSize="small" />
                          <Typography variant="body2" noWrap>
                            {scan.project_path}
                          </Typography>
                        </Box>
                      </TableCell>
                      <TableCell align="right">{scan.files_scanned}</TableCell>
                      <TableCell align="right">
                        <Typography
                          variant="body2"
                          color={scan.vulnerabilities_found > 0 ? 'error' : 'success'}
                        >
                          {scan.vulnerabilities_found}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Box display="flex" alignItems="center" gap={1} justifyContent="flex-end">
                          <Schedule fontSize="small" />
                          {formatDuration(scan.duration_ms)}
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={scan.status}
                          color={getStatusColor(scan.status) as any}
                          size="small"
                        />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </Paper>
        </Grid>
      </Grid>

      {/* 项目状态概览 */}
      <Grid container spacing={3}>
        <Grid size={12}>
          <Paper elevation={2} sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom>
              项目状态概览
            </Typography>
            <Grid container spacing={2}>
              <Grid size={{ xs: 12, md: 3 }}>
                <Chip
                  icon={<CheckCircle />}
                  label={`${project_stats.total_functions} 个函数`}
                  color="primary"
                  variant="outlined"
                  sx={{ width: '100%' }}
                />
              </Grid>
              <Grid size={{ xs: 12, md: 3 }}>
                <Chip
                  icon={<Code />}
                  label={`${project_stats.total_classes} 个类`}
                  color="secondary"
                  variant="outlined"
                  sx={{ width: '100%' }}
                />
              </Grid>
              <Grid size={{ xs: 12, md: 3 }}>
                <Chip
                  icon={<Security />}
                  label={`${vulnerability_stats.fixed}/${vulnerability_stats.total} 漏洞已修复`}
                  color={vulnerability_stats.fixed > vulnerability_stats.total / 2 ? 'success' : 'warning'}
                  variant="outlined"
                  sx={{ width: '100%' }}
                />
              </Grid>
              <Grid size={{ xs: 12, md: 3 }}>
                <Chip
                  icon={<Timeline />}
                  label={`${Object.keys(project_stats.languages).length} 种语言`}
                  color="info"
                  variant="outlined"
                  sx={{ width: '100%' }}
                />
              </Grid>
            </Grid>
            {project_stats.last_scan_time && (
              <Box mt={2}>
                <Typography variant="body2" color="textSecondary">
                  最后扫描时间: {formatTimestamp(project_stats.last_scan_time)}
                </Typography>
              </Box>
            )}
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard;