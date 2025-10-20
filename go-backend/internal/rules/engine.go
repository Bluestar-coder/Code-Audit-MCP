package rules

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// VulnerabilityRule 漏洞检测规则
type VulnerabilityRule struct {
	Name        string            `json:"name"`
	ID          string            `json:"id"`
	Category    string            `json:"category"`
	Severity    string            `json:"severity"`
	Description string            `json:"description"`
	Language    []string          `json:"language"`
	Patterns    []Pattern         `json:"patterns"`
	SafePatterns []Pattern        `json:"safe_patterns"`
	Examples    Examples          `json:"examples"`
}

// Pattern 检测模式
type Pattern struct {
	Pattern  string   `json:"pattern"`
	Message  string   `json:"message"`
	Severity string   `json:"severity"`
	Language []string `json:"language"`
}

// Examples 示例代码
type Examples struct {
	Vulnerable []string `json:"vulnerable"`
	Safe       []string `json:"safe"`
}

// Finding 检测发现的问题
type Finding struct {
	RuleID      string            `json:"rule_id"`
	RuleName    string            `json:"rule_name"`
	Category    string            `json:"category"`
	Severity    string            `json:"severity"`
	Message     string            `json:"message"`
	FilePath    string            `json:"file_path"`
	Line        int               `json:"line"`
	Column      int               `json:"column"`
	Code        string            `json:"code"`
	Language    string            `json:"language"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RuleEngine 规则引擎
type RuleEngine struct {
	rules       []VulnerabilityRule
	rulesByLang map[string][]VulnerabilityRule
}

// NewRuleEngine 创建新的规则引擎
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		rules:       make([]VulnerabilityRule, 0),
		rulesByLang: make(map[string][]VulnerabilityRule),
	}
}

// LoadBuiltinRules 加载内置规则
func (re *RuleEngine) LoadBuiltinRules() error {
	// SQL注入规则
	sqlInjectionRule := VulnerabilityRule{
		Name:        "SQL Injection Detection",
		ID:          "sql_injection",
		Category:    "injection",
		Severity:    "high",
		Description: "Detects potential SQL injection vulnerabilities",
		Language:    []string{"javascript", "typescript", "python", "go"},
		Patterns: []Pattern{
			{
				Pattern:  `["\'].*\+.*["\']`,
				Message:  "Potential SQL injection: string concatenation in SQL query",
				Severity: "high",
			},
			{
				Pattern:  `\$\{.*\}`,
				Message:  "Potential SQL injection: template string in SQL query",
				Severity: "high",
			},
			{
				Pattern:  `(?i)(query|execute|exec)\s*\(\s*.*%.*%`,
				Message:  "Potential SQL injection: string formatting in SQL query",
				Severity: "high",
			},
		},
		SafePatterns: []Pattern{
			{
				Pattern: `(?i)(prepare|prepared|parameterized)`,
				Message: "Safe: using prepared statements",
			},
		},
	}

	// XSS规则
	xssRule := VulnerabilityRule{
		Name:        "Cross-Site Scripting (XSS) Detection",
		ID:          "xss",
		Category:    "injection",
		Severity:    "medium",
		Description: "Detects potential XSS vulnerabilities",
		Language:    []string{"javascript", "typescript", "python"},
		Patterns: []Pattern{
			{
				Pattern:  `innerHTML\s*=`,
				Message:  "Potential XSS: direct assignment to innerHTML",
				Severity: "medium",
			},
			{
				Pattern:  `document\.write\s*\(`,
				Message:  "Potential XSS: document.write usage",
				Severity: "medium",
			},
			{
				Pattern:  `\$\(.*\)\.html\s*\(`,
				Message:  "Potential XSS: jQuery html() usage",
				Severity: "medium",
			},
		},
		SafePatterns: []Pattern{
			{
				Pattern: `textContent|innerText`,
				Message: "Safe: using textContent or innerText",
			},
		},
	}

	// 路径遍历规则
	pathTraversalRule := VulnerabilityRule{
		Name:        "Path Traversal Detection",
		ID:          "path_traversal",
		Category:    "path_traversal",
		Severity:    "high",
		Description: "Detects potential path traversal vulnerabilities",
		Language:    []string{"javascript", "typescript", "python", "go"},
		Patterns: []Pattern{
			{
				Pattern:  `\.\.\/|\.\.\\`,
				Message:  "Potential path traversal: relative path detected",
				Severity: "high",
			},
			{
				Pattern:  `["'][^"']*["']\s*\+\s*\w+`,
				Message:  "Potential path traversal: string concatenation with user input",
				Severity: "high",
			},
			{
				Pattern:  `(readFile|writeFile|open|readFileSync|writeFileSync|createReadStream|createWriteStream)\s*\(\s*.*\+`,
				Message:  "Potential path traversal: file operation with concatenated path",
				Severity: "high",
			},
			{
				Pattern:  `["']\.\./.*["']`,
				Message:  "Potential path traversal: relative path in string literal",
				Severity: "medium",
			},
		},
		SafePatterns: []Pattern{
			{
				Pattern: `path\.resolve|os\.path\.abspath|filepath\.Clean`,
				Message: "Safe: using path resolution functions",
			},
		},
	}

	// 添加规则到引擎
	re.rules = append(re.rules, sqlInjectionRule, xssRule, pathTraversalRule)
	
	// 构建按语言索引的规则映射
	re.buildLanguageIndex()
	
	fmt.Printf("✅ Loaded %d built-in vulnerability detection rules\n", len(re.rules))
	return nil
}

// buildLanguageIndex 构建按语言索引的规则映射
func (re *RuleEngine) buildLanguageIndex() {
	re.rulesByLang = make(map[string][]VulnerabilityRule)
	
	for _, rule := range re.rules {
		for _, lang := range rule.Language {
			re.rulesByLang[lang] = append(re.rulesByLang[lang], rule)
		}
	}
}

// ScanCode 扫描代码查找漏洞
func (re *RuleEngine) ScanCode(code, filePath, language string) ([]Finding, error) {
	startTime := time.Now()
	var findings []Finding

	// 获取适用于该语言的规则
	applicableRules := re.rulesByLang[language]
	if len(applicableRules) == 0 {
		// 如果没有特定语言的规则，使用通用规则
		for _, rule := range re.rules {
			if len(rule.Language) == 0 || contains(rule.Language, language) {
				applicableRules = append(applicableRules, rule)
			}
		}
	}

	// 按行分割代码
	lines := strings.Split(code, "\n")

	// 对每个规则进行检测
	for _, rule := range applicableRules {
		ruleFindings := re.scanWithRule(rule, lines, filePath, language)
		findings = append(findings, ruleFindings...)
	}

	// 记录扫描时间
	scanTime := time.Since(startTime)
	fmt.Printf("🔍 Scanned %s with %d rules in %v, found %d issues\n", 
		filePath, len(applicableRules), scanTime, len(findings))

	return findings, nil
}

// scanWithRule 使用单个规则扫描代码
func (re *RuleEngine) scanWithRule(rule VulnerabilityRule, lines []string, filePath, language string) []Finding {
	var findings []Finding

	// 检查每个模式
	for _, pattern := range rule.Patterns {
		// 如果模式指定了语言，检查是否匹配
		if len(pattern.Language) > 0 && !contains(pattern.Language, language) {
			continue
		}

		// 编译正则表达式
		regex, err := regexp.Compile(pattern.Pattern)
		if err != nil {
			fmt.Printf("⚠️ Invalid regex pattern in rule %s: %s\n", rule.ID, pattern.Pattern)
			continue
		}

		// 在每行中查找匹配
		for lineNum, line := range lines {
			if matches := regex.FindAllStringIndex(line, -1); matches != nil {
				for _, match := range matches {
					// 检查是否有安全模式排除这个匹配
					if re.isSafePattern(rule, line, language) {
						continue
					}

					severity := pattern.Severity
					if severity == "" {
						severity = rule.Severity
					}

					finding := Finding{
						RuleID:   rule.ID,
						RuleName: rule.Name,
						Category: rule.Category,
						Severity: severity,
						Message:  pattern.Message,
						FilePath: filePath,
						Line:     lineNum + 1,
						Column:   match[0] + 1,
						Code:     strings.TrimSpace(line),
						Language: language,
						Metadata: map[string]interface{}{
							"pattern":     pattern.Pattern,
							"match_start": match[0],
							"match_end":   match[1],
						},
					}

					findings = append(findings, finding)
				}
			}
		}
	}

	return findings
}

// isSafePattern 检查是否匹配安全模式
func (re *RuleEngine) isSafePattern(rule VulnerabilityRule, line, language string) bool {
	for _, safePattern := range rule.SafePatterns {
		// 如果安全模式指定了语言，检查是否匹配
		if len(safePattern.Language) > 0 && !contains(safePattern.Language, language) {
			continue
		}

		regex, err := regexp.Compile(safePattern.Pattern)
		if err != nil {
			continue
		}

		if regex.MatchString(line) {
			return true
		}
	}
	return false
}

// GetRules 获取所有规则
func (re *RuleEngine) GetRules() []VulnerabilityRule {
	return re.rules
}

// GetRulesByLanguage 获取特定语言的规则
func (re *RuleEngine) GetRulesByLanguage(language string) []VulnerabilityRule {
	return re.rulesByLang[language]
}

// GetRuleByID 根据ID获取规则
func (re *RuleEngine) GetRuleByID(id string) *VulnerabilityRule {
	for _, rule := range re.rules {
		if rule.ID == id {
			return &rule
		}
	}
	return nil
}

// ExportFindings 导出检测结果为JSON
func (re *RuleEngine) ExportFindings(findings []Finding) ([]byte, error) {
	return json.MarshalIndent(findings, "", "  ")
}

// GetStatistics 获取扫描统计信息
func (re *RuleEngine) GetStatistics(findings []Finding) map[string]interface{} {
	stats := map[string]interface{}{
		"total_findings": len(findings),
		"by_severity":    make(map[string]int),
		"by_category":    make(map[string]int),
		"by_rule":        make(map[string]int),
	}

	severityCount := make(map[string]int)
	categoryCount := make(map[string]int)
	ruleCount := make(map[string]int)

	for _, finding := range findings {
		severityCount[finding.Severity]++
		categoryCount[finding.Category]++
		ruleCount[finding.RuleID]++
	}

	stats["by_severity"] = severityCount
	stats["by_category"] = categoryCount
	stats["by_rule"] = ruleCount

	return stats
}

// contains 检查字符串切片是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}