import React, { useState } from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Container,
  Box,
  Tabs,
  Tab,
  Paper,
  ThemeProvider,
  createTheme,
  CssBaseline,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  BugReport,
  Timeline,
  Code,
} from '@mui/icons-material';

// 导入组件
import Dashboard from './components/Dashboard';
import VulnerabilityList from './components/VulnerabilityList';
import TaintAnalysis from './components/TaintAnalysis';
import CodeAnalysis from './components/CodeAnalysis';

const theme = createTheme({
  palette: {
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});

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

function App() {
  const [value, setValue] = useState(0);

  const handleChange = (event: React.SyntheticEvent, newValue: number) => {
    setValue(newValue);
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box sx={{ flexGrow: 1 }}>
        <AppBar position="static">
          <Toolbar>
            <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
              代码安全审计平台
            </Typography>
          </Toolbar>
        </AppBar>
        
        <Container maxWidth="xl" sx={{ mt: 2 }}>
          <Paper elevation={1}>
            <Tabs
              value={value}
              onChange={handleChange}
              aria-label="audit tabs"
              variant="fullWidth"
            >
              <Tab
                icon={<DashboardIcon />}
                label="仪表板"
                id="tab-0"
                aria-controls="tabpanel-0"
              />
              <Tab
                icon={<BugReport />}
                label="漏洞列表"
                id="tab-1"
                aria-controls="tabpanel-1"
              />
              <Tab
                icon={<Timeline />}
                label="污点分析"
                id="tab-2"
                aria-controls="tabpanel-2"
              />
              <Tab
                icon={<Code />}
                label="代码分析"
                id="tab-3"
                aria-controls="tabpanel-3"
              />
            </Tabs>
          </Paper>

          <TabPanel value={value} index={0}>
            <Dashboard />
          </TabPanel>
          
          <TabPanel value={value} index={1}>
            <VulnerabilityList />
          </TabPanel>
          
          <TabPanel value={value} index={2}>
            <TaintAnalysis />
          </TabPanel>
          
          <TabPanel value={value} index={3}>
            <CodeAnalysis />
          </TabPanel>
        </Container>
      </Box>
    </ThemeProvider>
  );
}

export default App;
