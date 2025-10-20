# 代码审计及POC/EXP生成MCP服务器

> 基于AI的自动化代码安全审计工具，支持MCP协议，采用Python+Go混合架构

## 📌 项目状态

**当前阶段**: Phase 0 - 基础设施搭建 (90% 完成)  
**开始日期**: 2025-10-19  

### 📚 核心文档

| 文档 | 状态 | 描述 |
|------|------|------|
| [requirements.md](requirements.md) | ✅ 完成 | 36个功能需求详细说明 |
| [design.md](design.md) | ✅ 完成 | 完整系统设计（19章） |
| [implementation_plan.md](implementation_plan.md) | ✅ 完成 | 11周实施计划（5阶段） |
| [FINAL_STATUS.md](FINAL_STATUS.md) | ✅ 完成 | 项目状态报告 |

---

## 🎯 项目简介

本项目是一个智能化的代码安全审计系统，通过深度代码分析、污点追踪、多模型验证等技术，准确识别安全漏洞，并自动生成POC和EXP代码。系统基于MCP（Model Context Protocol）协议实现，可与Claude Desktop等AI客户端无缝集成。

## ✨ 核心特性

- 🔍 **多语言支持**: Java、Python、JavaScript、PHP、Go等主流语言
- 🎨 **框架深度适配**: Spring、Django、Express、Laravel等主流框架
- 🧠 **AI驱动分析**: 多模型验证（Claude、GPT-4、Gemini）减少误报
- 🚀 **高性能引擎**: Go实现AST解析和图分析，性能优异
- 🔬 **函数级污点分析**: 准确追踪数据流，避免断链
- 🤖 **自动POC/EXP生成**: AI自动生成漏洞验证和利用代码
- 📊 **语义代码搜索**: 基于txtai的智能代码检索
- 🔄 **CVE关联**: 关联历史漏洞，提供修复参考

## 🏗️ 架构设计

### Python + Go 混合架构

```
┌─────────────────────────────────────────┐
│         MCP Client (Claude Desktop)      │
└──────────────────┬──────────────────────┘
                   │ MCP Protocol
┌──────────────────▼──────────────────────┐
│      Python MCP Server (主进程)          │
│  - MCP协议处理 (mcp SDK)                 │
│  - AI模型集成 (Claude/GPT-4/Gemini)      │
│  - 语义搜索 (txtai)                      │
│  - POC/EXP生成                          │
│  - CVE查询                              │
└──────────────────┬──────────────────────┘
                   │ gRPC
┌──────────────────▼──────────────────────┐
│        Go Backend Service (后端)         │
│  - AST解析 (go-tree-sitter)             │
│  - 代码索引管理                          │
│  - 调用链分析                            │
│  - 污点分析                             │
│  - 高性能计算                            │
└─────────────────────────────────────────┘
```

**为什么选择Python+Go？**

| 组件 | 语言选择 | 原因 |
|------|---------|------|
| MCP Server | Python | 官方MCP SDK支持最完善 |
| AI/ML服务 | Python | 丰富的AI/ML生态（txtai、transformers） |
| 性能引擎 | Go | 高性能、并发处理AST和图算法 |
| 通信层 | gRPC | 高性能跨语言RPC |

## 🚀 快速开始

### 前置要求

- Python 3.9+
- Go 1.21+
- protoc (Protocol Buffers编译器)

### 安装

```bash
# 1. 克隆项目
git clone https://github.com/your-org/code-audit-mcp.git
cd code-audit-mcp

# 2. 运行安装脚本
./scripts/install.sh

# 3. 配置环境变量
export ANTHROPIC_API_KEY="your-api-key"
export OPENAI_API_KEY="your-api-key"
```

### 配置Claude Desktop

在Claude Desktop配置文件中添加：

```json
{
  "mcpServers": {
    "code-audit": {
      "command": "python",
      "args": ["-m", "codeaudit.server"],
      "env": {
        "ANTHROPIC_API_KEY": "${ANTHROPIC_API_KEY}",
        "GO_SERVICE_BINARY": "./bin/code-audit-go"
      }
    }
  }
}
```

### 使用示例

在Claude Desktop中与AI对话：

```
你: 帮我扫描 /path/to/project 的安全漏洞

AI: 好的，我将使用code-audit工具扫描该项目...
[调用 scan_vulnerabilities 工具]

扫描完成！发现以下安全问题：

1. SQL注入漏洞 (严重程度: 高)
   - 位置: UserController.java:45
   - 描述: 用户输入直接拼接到SQL查询
   - POC已生成: [显示POC代码]
   - 修复建议: 使用参数化查询

2. XSS漏洞 (严重程度: 中)
   ...
```

## 📚 主要功能

### MCP工具列表

| 工具名称 | 功能描述 |
|---------|---------|
| `search_class` | 搜索类定义 |
| `search_method` | 搜索方法定义 |
| `scan_vulnerabilities` | 扫描代码漏洞 |
| `analyze_call_chain` | 分析函数调用链 |
| `trace_sink` | 追踪危险函数 |
| `generate_poc` | 生成POC代码 |
| `generate_exp` | 生成EXP代码 |
| `semantic_search` | 语义代码搜索 |
| `query_cve` | 查询历史漏洞 |
| `extract_chain_code` | 提取漏洞链路代码 |

### 支持的漏洞类型

- ✅ SQL注入（包括盲注、时间盲注）
- ✅ 跨站脚本（XSS）
- ✅ 命令注入和代码注入
- ✅ 文件包含（LFI/RFI）
- ✅ 路径遍历
- ✅ 不安全的反序列化
- ✅ SSRF（服务器端请求伪造）
- ✅ XXE（XML外部实体注入）
- ✅ 硬编码敏感信息
- ✅ 不安全的加密算法
- ✅ CSRF防护缺失

## 🛠️ 技术栈

### Python端
- **MCP SDK**: mcp Python库
- **AI模型**: anthropic、openai、google-generativeai
- **语义搜索**: txtai
- **ML框架**: transformers、scikit-learn
- **gRPC**: grpcio

### Go端
- **AST解析**: go-tree-sitter
- **gRPC**: google.golang.org/grpc
- **数据库**: gorm
- **并发**: goroutines

### 数据存储
- **SQLite/PostgreSQL**: 代码索引
- **Qdrant**: 向量存储
- **Redis**: 缓存（可选）

## 📖 文档

- [需求文档](requirements.md) - 详细功能需求
- [设计文档](design.md) - 系统架构和设计
- [API文档](docs/api.md) - API接口说明（待完成）

## 🗺️ 开发路线图

- [x] 需求分析和设计
- [ ] **Phase 1** (2-3个月): 核心基础 - AST解析、索引、基础MCP工具
- [ ] **Phase 2** (2-3个月): 漏洞检测 - 污点分析、调用链、多模型验证
- [ ] **Phase 3** (2-3个月): 高级功能 - 多语言、框架适配、POC生成
- [ ] **Phase 4** (1-2个月): 增强优化 - 性能优化、文档完善
- [ ] **Phase 5** (1个月): 生产就绪 - 安全加固、部署发布

预计总开发时间: **7-8个月**

## 🤝 贡献

欢迎贡献代码、报告bug或提出新功能建议！

1. Fork项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [MCP Protocol](https://modelcontextprotocol.io/) - Model Context Protocol规范
- [go-tree-sitter](https://github.com/smacker/go-tree-sitter) - Go语言的Tree-sitter绑定
- [txtai](https://github.com/neuml/txtai) - AI驱动的语义搜索
- [OWASP](https://owasp.org/) - 安全最佳实践和规则

## 📮 联系方式

- 问题反馈: [GitHub Issues](https://github.com/your-org/code-audit-mcp/issues)
- 邮件: security@example.com

---

**注意**: 本工具仅用于合法的安全研究和授权的安全测试。使用者需对使用本工具产生的一切后果负责。

