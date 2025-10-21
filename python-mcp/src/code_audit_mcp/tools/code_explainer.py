from typing import Any, Dict

from .base import BaseTool


class CodeExplainer(BaseTool):
    async def execute(self, code: str, language: str = "") -> Dict[str, Any]:
        try:
            # 生成宿主 LLM 使用的提示词，不在服务器侧直接调用 SDK
            lines = code.strip().splitlines()
            non_empty = [l for l in lines if l.strip()]
            metrics = {
                "lines": len(lines),
                "non_empty": len(non_empty),
                "avg_length": sum(len(l) for l in lines) / max(1, len(lines)),
            }
            system = "你是资深代码审计专家，擅长安全与可维护性分析。"
            user_prompt = (
                "请用简洁、结构化的方式解释以下" + (language or "代码") + "：\n\n" + code +
                "\n\n要求：\n"
                "- 概述模块职责、关键函数与数据结构\n"
                "- 指出可能的安全风险（输入校验、命令执行、SQL注入、路径穿越、XSS 等）\n"
                "- 给出改进建议（边界检查、净化、最小权限、日志与错误处理）\n"
                "- 如果存在外部交互，描述数据流（源、汇、净化点）\n"
            )
            return self.format_success({
                "language": language or "unknown",
                "llm_prompt": {
                    "system": system,
                    "user": user_prompt,
                },
                "metrics": metrics,
                "ai_mode": "host",
                "note": "此工具不直接调用LLM，请在宿主中使用提示词生成结果。",
            })
        except Exception as e:
            return self.format_error(e)