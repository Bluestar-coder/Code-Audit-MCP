import React, { useState, useEffect } from 'react';
import {
  Grid2 as Grid,
  Card,
  CardContent,
  Typography,
  Box,
  Paper,
  Chip,
  LinearProgress,
} from '@mui/material';
import {
  Security,
  BugReport,
  Warning,
  CheckCircle,
  Code,
  Timeline,
} from '@mui/icons-material';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

interface VulnerabilityStats {
  total: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  fixed: number;
}

interface AnalysisStats {
  filesAnalyzed: number;
  linesOfCode: number;
  functions: number;
  classes: number;
}

const Dashboard: React.FC = () => {
  const [vulnStats, setVulnStats] = useState<VulnerabilityStats>({
    total: 0,
    critical: 0,
    high: 0,
    medium: 0,
    low: 0,
    fixed: 0,
  });

  const [analysisStats, setAnalysisStats] = useState<AnalysisStats>({
    filesAnalyzed: 0,
    linesOfCode: 0,
    functions: 0,
    classes: 0,
  });

  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 模拟数据加载
    setTimeout(() => {
      setVulnStats({
        total: 47,
        critical: 3,
        high: 8,
        medium: 15,
        low: 21,
        fixed: 12,
      });
      setAnalysisStats({
        filesAnalyzed: 156,
        linesOfCode: 45230,
        functions: 892,
        classes: 234,
      });
      setLoading(false);
    }, 1000);
  }, []);

  const vulnerabilityData = [
    { name: '严重', value: vulnStats.critical, color: '#f44336' },
    { name: '高危', value: vulnStats.high, color: '#ff9800' },
    { name: '中危', value: vulnStats.medium, color: '#ffeb3b' },
    { name: '低危', value: vulnStats.low, color: '#4caf50' },
  ];

  const trendData = [
    { month: '1月', vulnerabilities: 52, fixed: 8 },
    { month: '2月', vulnerabilities: 48, fixed: 12 },
    { month: '3月', vulnerabilities: 45, fixed: 15 },
    { month: '4月', vulnerabilities: 47, fixed: 12 },
  ];

  const StatCard: React.FC<{
    title: string;
    value: number | string;
    icon: React.ReactNode;
    color: string;
    subtitle?: string;
  }> = ({ title, value, icon, color, subtitle }) => (
    <Card elevation={2} sx={{ height: '100%' }}>
      <CardContent>
        <Box display="flex" alignItems="center" justifyContent="space-between">
          <Box>
            <Typography color="textSecondary" gutterBottom variant="body2">
              {title}
            </Typography>
            <Typography variant="h4" component="div" color={color}>
              {value}
            </Typography>
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

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        安全分析仪表板
      </Typography>
      
      {/* 统计卡片 */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid xs={12} sm={6} md={3}>
          <StatCard
            title="总漏洞数"
            value={vulnStats.total}
            icon={<BugReport />}
            color="#f44336"
            subtitle={`${vulnStats.fixed} 个已修复`}
          />
        </Grid>
        <Grid xs={12} sm={6} md={3}>
          <StatCard
            title="严重漏洞"
            value={vulnStats.critical}
            icon={<Warning />}
            color="#d32f2f"
          />
        </Grid>
        <Grid xs={12} sm={6} md={3}>
          <StatCard
            title="已分析文件"
            value={analysisStats.filesAnalyzed}
            icon={<Code />}
            color="#1976d2"
          />
        </Grid>
        <Grid xs={12} sm={6} md={3}>
          <StatCard
            title="代码行数"
            value={analysisStats.linesOfCode.toLocaleString()}
            icon={<Timeline />}
            color="#388e3c"
          />
        </Grid>
      </Grid>

      {/* 图表区域 */}
      <Grid container spacing={3}>
        <Grid xs={12} md={6}>
          <Paper elevation={2} sx={{ p: 2, height: 400 }}>
            <Typography variant="h6" gutterBottom>
              漏洞分布
            </Typography>
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>
        <Grid xs={12} md={6}>
          <Paper elevation={2} sx={{ p: 2, height: 400 }}>
            <Typography variant="h6" gutterBottom>
              漏洞趋势
            </Typography>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={barData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="count" fill="#8884d8" />
              </BarChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>
        {/* 快速状态 */}
        <Grid xs={12}>
          <Paper elevation={2} sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom>
              项目状态
            </Typography>
            <Grid container spacing={2}>
              <Grid>
                <Chip
                  icon={<CheckCircle />}
                  label={`${analysisStats.functions} 个函数`}
                  color="primary"
                  variant="outlined"
                />
              </Grid>
              <Grid>
                <Chip
                  icon={<Code />}
                  label={`${analysisStats.classes} 个类`}
                  color="secondary"
                  variant="outlined"
                />
              </Grid>
              <Grid>
                <Chip
                  icon={<Security />}
                  label={`${vulnStats.fixed}/${vulnStats.total} 漏洞已修复`}
                  color={vulnStats.fixed > vulnStats.total / 2 ? 'success' : 'warning'}
                  variant="outlined"
                />
              </Grid>
            </Grid>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard;