from typing import Any, Dict

from .base import BaseTool


class POCGenerator(BaseTool):
    async def execute(self, vulnerability_id: str, language: str = "python", context: str | None = None) -> Dict[str, Any]:
        try:
            from code_audit_mcp.ai.provider import generate_poc_with_ai, AIProvider
            provider = AIProvider()
            ai_available = await provider.available()
            if ai_available:
                poc = await generate_poc_with_ai(vulnerability_id, language, context)
            else:
                poc = f"""
# POC for vulnerability: {vulnerability_id}
# Language: {language}
# [AI未启用] 使用占位模板。设置 ANTHROPIC_API_KEY 后可生成更贴近漏洞的PoC。

print("POC placeholder for", {vulnerability_id!r})
"""
            return self.format_success({
                "vulnerability_id": vulnerability_id,
                "language": language,
                "poc_code": poc.strip(),
                "ai_enabled": ai_available,
            })
        except Exception as e:
            return self.format_error(e)