package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	pb "code-audit-mcp/proto"
)

// Service 规则引擎服务
type Service struct {
	pb.UnimplementedVulnerabilityDetectorServer
	engine *RuleEngine
}

// NewService 创建新的规则引擎服务
func NewService() (*Service, error) {
	engine := NewRuleEngine()
	
	// 加载内置规则
	if err := engine.LoadBuiltinRules(); err != nil {
		return nil, fmt.Errorf("failed to load builtin rules: %w", err)
	}
	
	return &Service{
		engine: engine,
	}, nil
}

// ScanFile 扫描单个文件
func (s *Service) ScanFile(ctx context.Context, req *pb.ScanFileRequest) (*pb.ScanFileResponse, error) {
	// 检测语言
	language := detectLanguageFromPath(req.FilePath)
	if req.Language != "" {
		language = req.Language
	}

	// 扫描代码
	findings, err := s.engine.ScanCode(req.Content, req.FilePath, language)
	if err != nil {
		return &pb.ScanFileResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 转换为protobuf格式
	pbFindings := make([]*pb.VulnerabilityFinding, len(findings))
	for i, finding := range findings {
		metadata, _ := json.Marshal(finding.Metadata)
		pbFindings[i] = &pb.VulnerabilityFinding{
			RuleId:      finding.RuleID,
			RuleName:    finding.RuleName,
			Category:    finding.Category,
			Severity:    finding.Severity,
			Message:     finding.Message,
			FilePath:    finding.FilePath,
			Line:        int32(finding.Line),
			Column:      int32(finding.Column),
			Code:        finding.Code,
			Language:    finding.Language,
			Metadata:    string(metadata),
		}
	}

	// 生成统计信息
	stats := s.engine.GetStatistics(findings)
	statsJson, _ := json.Marshal(stats)

	return &pb.ScanFileResponse{
		Success:   true,
		FilePath:  req.FilePath,
		Language:  language,
		Findings:  pbFindings,
		Statistics: string(statsJson),
	}, nil
}

// ScanBatch 批量扫描文件
func (s *Service) ScanBatch(ctx context.Context, req *pb.ScanBatchRequest) (*pb.ScanBatchResponse, error) {
	var allFindings []Finding
	responses := make([]*pb.ScanFileResponse, len(req.Files))

	for i, file := range req.Files {
		// 检测语言
		language := detectLanguageFromPath(file.FilePath)
		if file.Language != "" {
			language = file.Language
		}

		// 扫描代码
		findings, err := s.engine.ScanCode(file.Content, file.FilePath, language)
		if err != nil {
			responses[i] = &pb.ScanFileResponse{
				Success:  false,
				FilePath: file.FilePath,
				Error:    err.Error(),
			}
			continue
		}

		// 转换为protobuf格式
		pbFindings := make([]*pb.VulnerabilityFinding, len(findings))
		for j, finding := range findings {
			metadata, _ := json.Marshal(finding.Metadata)
			pbFindings[j] = &pb.VulnerabilityFinding{
				RuleId:      finding.RuleID,
				RuleName:    finding.RuleName,
				Category:    finding.Category,
				Severity:    finding.Severity,
				Message:     finding.Message,
				FilePath:    finding.FilePath,
				Line:        int32(finding.Line),
				Column:      int32(finding.Column),
				Code:        finding.Code,
				Language:    finding.Language,
				Metadata:    string(metadata),
			}
		}

		// 生成统计信息
		stats := s.engine.GetStatistics(findings)
		statsJson, _ := json.Marshal(stats)

		responses[i] = &pb.ScanFileResponse{
			Success:    true,
			FilePath:   file.FilePath,
			Language:   language,
			Findings:   pbFindings,
			Statistics: string(statsJson),
		}

		allFindings = append(allFindings, findings...)
	}

	// 生成总体统计信息
	overallStats := s.engine.GetStatistics(allFindings)
	overallStatsJson, _ := json.Marshal(overallStats)

	return &pb.ScanBatchResponse{
		Results:           responses,
		OverallStatistics: string(overallStatsJson),
	}, nil
}

// GetRules 获取所有规则
func (s *Service) GetRules(ctx context.Context, req *pb.GetRulesRequest) (*pb.GetRulesResponse, error) {
	rules := s.engine.GetRules()
	
	// 按语言过滤
	if req.Language != "" {
		rules = s.engine.GetRulesByLanguage(req.Language)
	}

	// 转换为protobuf格式
	pbRules := make([]*pb.VulnerabilityRule, len(rules))
	for i, rule := range rules {
		// 转换模式
		patterns := make([]*pb.RulePattern, len(rule.Patterns))
		for j, pattern := range rule.Patterns {
			patterns[j] = &pb.RulePattern{
				Pattern:  pattern.Pattern,
				Message:  pattern.Message,
				Severity: pattern.Severity,
				Language: pattern.Language,
			}
		}

		// 转换安全模式
		safePatterns := make([]*pb.RulePattern, len(rule.SafePatterns))
		for j, pattern := range rule.SafePatterns {
			safePatterns[j] = &pb.RulePattern{
				Pattern:  pattern.Pattern,
				Message:  pattern.Message,
				Severity: pattern.Severity,
				Language: pattern.Language,
			}
		}

		pbRules[i] = &pb.VulnerabilityRule{
			Name:         rule.Name,
			Id:           rule.ID,
			Category:     rule.Category,
			Severity:     rule.Severity,
			Description:  rule.Description,
			Language:     rule.Language,
			Patterns:     patterns,
			SafePatterns: safePatterns,
		}
	}

	return &pb.GetRulesResponse{
		Rules: pbRules,
	}, nil
}

// GetRuleById 根据ID获取规则
func (s *Service) GetRuleById(ctx context.Context, req *pb.GetRuleByIdRequest) (*pb.GetRuleByIdResponse, error) {
	rule := s.engine.GetRuleByID(req.RuleId)
	if rule == nil {
		return &pb.GetRuleByIdResponse{
			Found: false,
		}, nil
	}

	// 转换模式
	patterns := make([]*pb.RulePattern, len(rule.Patterns))
	for i, pattern := range rule.Patterns {
		patterns[i] = &pb.RulePattern{
			Pattern:  pattern.Pattern,
			Message:  pattern.Message,
			Severity: pattern.Severity,
			Language: pattern.Language,
		}
	}

	// 转换安全模式
	safePatterns := make([]*pb.RulePattern, len(rule.SafePatterns))
	for i, pattern := range rule.SafePatterns {
		safePatterns[i] = &pb.RulePattern{
			Pattern:  pattern.Pattern,
			Message:  pattern.Message,
			Severity: pattern.Severity,
			Language: pattern.Language,
		}
	}

	pbRule := &pb.VulnerabilityRule{
		Name:         rule.Name,
		Id:           rule.ID,
		Category:     rule.Category,
		Severity:     rule.Severity,
		Description:  rule.Description,
		Language:     rule.Language,
		Patterns:     patterns,
		SafePatterns: safePatterns,
	}

	return &pb.GetRuleByIdResponse{
		Found: true,
		Rule:  pbRule,
	}, nil
}

// detectLanguageFromPath 从文件路径检测编程语言
func detectLanguageFromPath(filePath string) string {
    ext := filepath.Ext(filePath)
    switch ext {
    case ".js":
        return "javascript"
    case ".mjs", ".cjs":
        return "nodejs"
    case ".ejs":
        return "nodejs"
    case ".ts":
        return "typescript"
    case ".py":
        return "python"
    case ".go":
        return "go"
    case ".java":
        return "java"
    case ".xml":
        return "java"
    case ".cpp", ".cc", ".cxx":
        return "cpp"
    case ".c":
        return "c"
    case ".cs":
        return "csharp"
    case ".php":
        return "php"
    case ".rb":
        return "ruby"
    case ".rs":
        return "rust"
    case ".swift":
        return "swift"
    case ".kt":
        return "kotlin"
    case ".scala":
        return "scala"
    default:
        return "unknown"
    }
}