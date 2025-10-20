from typing import Any, Dict

from .base import BaseTool


class CodeExplainer(BaseTool):
    async def execute(self, code: str, language: str = "") -> Dict[str, Any]:
        try:
            from code_audit_mcp.ai.provider import explain_code_with_ai, AIProvider
            lines = code.strip().splitlines()
            non_empty = [l for l in lines if l.strip()]
            metrics = {
                "lines": len(lines),
                "non_empty": len(non_empty),
                "avg_length": sum(len(l) for l in lines) / max(1, len(lines)),
            }
            # Try AI explanation; fallback to heuristic when AI not available
            provider = AIProvider()
            ai_available = await provider.available()
            summary = await explain_code_with_ai(code, language) if ai_available else (
                "基于启发式的解释：\n- 该代码大致由函数与类构成\n- 进一步解释需要启用 AI（设置 ANTHROPIC_API_KEY）"
            )
            return self.format_success({
                "language": language or "unknown",
                "summary": summary,
                "metrics": metrics,
                "ai_enabled": ai_available,
            })
        except Exception as e:
            return self.format_error(e)