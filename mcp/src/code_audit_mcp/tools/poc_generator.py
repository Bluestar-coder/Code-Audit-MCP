from typing import Any, Dict

from .base import BaseTool


class POCGenerator(BaseTool):
    async def execute(self, vulnerability_id: str, language: str = "python", context: str | None = None) -> Dict[str, Any]:
        try:
            # 生成宿主 LLM 使用的提示词，不在服务器侧直接调用 SDK
            ctx = context or ""
            system = "你是资深安全研究员，擅长生成最小化可复现PoC。"
            user_prompt = (
                f"根据漏洞ID {vulnerability_id} 生成一个最小可复现的{language} PoC。\n"
                f"如果有上下文，考虑如下代码片段：\n{ctx}\n"
                "要求：\n"
                "- 用最少代码复现漏洞触发路径\n"
                "- 注明运行步骤与依赖\n"
                "- 不要包含破坏性操作，仅展示漏洞触发\n"
                "- 若涉及外部输入/输出，标注源、汇与净化点\n"
            )
            return self.format_success({
                "vulnerability_id": vulnerability_id,
                "language": language,
                "llm_prompt": {
                    "system": system,
                    "user": user_prompt,
                },
                "ai_mode": "host",
                "note": "此工具不直接调用LLM，请在宿主中使用提示词生成结果。",
            })
        except Exception as e:
            return self.format_error(e)