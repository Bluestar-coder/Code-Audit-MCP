package analyzer

import (
	"regexp"

	pb "code-audit-mcp/proto"
)

// TaintAnalyzer 污点分析器
type TaintAnalyzer struct {
	pb.UnimplementedTaintAnalyzerServer
	sources     map[string]*SourceInfo
	sinks       map[string]*SinkInfo
	sanitizers  map[string]*SanitizerInfo
}

// SourceInfo 污点源信息
type SourceInfo struct {
	ID          string
	Name        string
	Type        string
	Keywords    []string
	Pattern     *regexp.Regexp
	Description string
}

// SinkInfo 污点汇信息
type SinkInfo struct {
	ID               string
	Name             string
	Type             string
	Keywords         []string
	Pattern          *regexp.Regexp
	VulnerabilityType string
	Description      string
}

// SanitizerInfo 净化函数信息
type SanitizerInfo struct {
	ID          string
	Name        string
	Pattern     *regexp.Regexp
	Description string
}

// TaintPath 污点路径
type TaintPath struct {
	Source      *PathNode
	Sink        *PathNode
	Nodes       []*PathNode
	HasSanitizer bool
	Confidence  float64
}

// PathNode 路径节点
type PathNode struct {
	NodeID       string
	FunctionName string
	FilePath     string
	LineNumber   int32
	Operation    string
	VariableName string
	DataFlow     string
}

// NewTaintAnalyzer 创建新的污点分析器
func NewTaintAnalyzer() *TaintAnalyzer {
	ta := &TaintAnalyzer{
		sources:    make(map[string]*SourceInfo),
		sinks:      make(map[string]*SinkInfo),
		sanitizers: make(map[string]*SanitizerInfo),
	}
	
	ta.loadBuiltinSources()
	ta.loadBuiltinSinks()
	ta.loadBuiltinSanitizers()
	
	return ta
}

// loadBuiltinSources 加载内置污点源
func (ta *TaintAnalyzer) loadBuiltinSources() {
	sources := []*SourceInfo{
		{
			ID:          "user_input",
			Name:        "User Input",
			Type:        "USER_INPUT",
			Keywords:    []string{"req.body", "req.query", "req.params", "request.form", "input", "prompt"},
			Pattern:     regexp.MustCompile(`(req\.(body|query|params)|request\.form|prompt\(|input\()`),
			Description: "User input from HTTP requests or interactive prompts",
		},
		{
			ID:          "file_input",
			Name:        "File Input",
			Type:        "FILE_INPUT",
			Keywords:    []string{"readFile", "readFileSync", "fs.read", "open"},
			Pattern:     regexp.MustCompile(`(readFile|readFileSync|fs\.read|open)\s*\(`),
			Description: "Data read from files",
		},
		{
			ID:          "env_vars",
			Name:        "Environment Variables",
			Type:        "ENVIRONMENT_VARIABLE",
			Keywords:    []string{"process.env", "os.environ", "getenv"},
			Pattern:     regexp.MustCompile(`(process\.env|os\.environ|getenv)\s*[\[\(]`),
			Description: "Environment variables",
		},
	}
	
	for _, source := range sources {
		ta.sources[source.ID] = source
	}
}

// loadBuiltinSinks 加载内置污点汇
func (ta *TaintAnalyzer) loadBuiltinSinks() {
	sinks := []*SinkInfo{
		{
			ID:               "sql_query",
			Name:             "SQL Query",
			Type:             "SQL_QUERY",
			Keywords:         []string{"query", "execute", "exec", "prepare"},
			Pattern:          regexp.MustCompile(`\.(query|execute|exec|prepare)\s*\(`),
			VulnerabilityType: "sql_injection",
			Description:      "SQL database operations",
		},
		{
			ID:               "html_output",
			Name:             "HTML Output",
			Type:             "HTML_OUTPUT",
			Keywords:         []string{"innerHTML", "outerHTML", "write", "document.write"},
			Pattern:          regexp.MustCompile(`(innerHTML|outerHTML|document\.write)\s*=`),
			VulnerabilityType: "xss",
			Description:      "HTML content output",
		},
		{
			ID:               "command_exec",
			Name:             "Command Execution",
			Type:             "COMMAND_EXECUTION",
			Keywords:         []string{"exec", "spawn", "system", "eval"},
			Pattern:          regexp.MustCompile(`(exec|spawn|system|eval)\s*\(`),
			VulnerabilityType: "command_injection",
			Description:      "Command execution functions",
		},
		{
			ID:               "file_write",
			Name:             "File Write",
			Type:             "FILE_WRITE",
			Keywords:         []string{"writeFile", "writeFileSync", "createWriteStream"},
			Pattern:          regexp.MustCompile(`(writeFile|writeFileSync|createWriteStream)\s*\(`),
			VulnerabilityType: "path_traversal",
			Description:      "File write operations",
		},
	}
	
	for _, sink := range sinks {
		ta.sinks[sink.ID] = sink
	}
}

// loadBuiltinSanitizers 加载内置净化函数
func (ta *TaintAnalyzer) loadBuiltinSanitizers() {
	sanitizers := []*SanitizerInfo{
		{
			ID:          "escape_html",
			Name:        "HTML Escape",
			Pattern:     regexp.MustCompile(`(escapeHtml|htmlspecialchars|escape)\s*\(`),
			Description: "HTML escaping functions",
		},
		{
			ID:          "sql_escape",
			Name:        "SQL Escape",
			Pattern:     regexp.MustCompile(`(mysql_real_escape_string|addslashes|prepare)\s*\(`),
			Description: "SQL escaping functions",
		},
		{
			ID:          "path_resolve",
			Name:        "Path Resolution",
			Pattern:     regexp.MustCompile(`(path\.resolve|path\.normalize|filepath\.Clean)\s*\(`),
			Description: "Path resolution and normalization functions",
		},
	}
	
	for _, sanitizer := range sanitizers {
		ta.sanitizers[sanitizer.ID] = sanitizer
	}
}