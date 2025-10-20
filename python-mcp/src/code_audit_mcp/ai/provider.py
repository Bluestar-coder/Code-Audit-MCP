import os
import asyncio
from typing import Optional, Dict, Any

# Optional imports guarded by environment variables
_ANTHROPIC_AVAILABLE = False
try:
    from anthropic import Anthropic
    _ANTHROPIC_AVAILABLE = True
except Exception:
    _ANTHROPIC_AVAILABLE = False

class AIProvider:
    """
    Simple AI provider that supports Anthropic (Claude) if configured via env `ANTHROPIC_API_KEY`.
    Falls back to heuristic output when not available.
    """

    def __init__(self, model: Optional[str] = None):
        self.model = model or os.getenv("CLAUDE_MODEL", "claude-3-5-sonnet-latest")
        self._client = None
        api_key = os.getenv("ANTHROPIC_API_KEY")
        if _ANTHROPIC_AVAILABLE and api_key:
            try:
                self._client = Anthropic(api_key=api_key)
            except Exception:
                self._client = None

    async def available(self) -> bool:
        return self._client is not None

    async def ask(self, system: str, user: str, max_tokens: int = 1024) -> str:
        # If client not initialized, return a heuristic message
        if self._client is None:
            return (
                "[AI未配置] 基于启发式的结果:\n" + user[:4000]
            )
        # Anthropic messages API (sync), wrap in thread to avoid blocking loop
        def _call() -> str:
            try:
                msg = self._client.messages.create(
                    model=self.model,
                    max_tokens=max_tokens,
                    system=system,
                    messages=[{"role": "user", "content": user}],
                )
                # Return plain text content
                parts = getattr(msg, "content", [])
                if parts:
                    # parts is a list of content blocks
                    texts = []
                    for p in parts:
                        t = getattr(p, "text", None)
                        if t:
                            texts.append(t)
                    return "\n".join(texts) if texts else str(msg)
                return str(msg)
            except Exception as e:
                return f"[AI调用失败] {e}"
        return await asyncio.to_thread(_call)


async def explain_code_with_ai(code: str, language: str) -> str:
    provider = AIProvider()
    prompt = (
        "请用简洁、结构化的方式解释以下" + language + "代码：\n\n" + code +
        "\n\n要求：\n- 概述模块职责与关键函数\n- 指出可能的安全风险\n- 给出改进建议\n- 如果存在外部交互，描述数据流\n"
    )
    system = "你是资深代码审计专家，擅长安全与可维护性分析。"
    return await provider.ask(system, prompt, max_tokens=1200)


async def generate_poc_with_ai(vuln_id: str, language: str, context: Optional[str]) -> str:
    provider = AIProvider()
    ctx = context or ""
    prompt = (
        f"根据漏洞ID {vuln_id} 生成一个最小可复现的{language} PoC。\n"
        f"如果有上下文，考虑如下代码片段：\n{ctx}\n"
        "要求：\n- 用最少代码复现漏洞触发路径\n- 注明运行步骤与依赖\n- 不要包含破坏性操作，仅展示漏洞触发\n"
    )
    system = "你是资深安全研究员，擅长生成最小化可复现PoC。"
    return await provider.ask(system, prompt, max_tokens=800)