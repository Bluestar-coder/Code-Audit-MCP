import asyncio
import json
from code_audit_mcp.tools.taint_tracer import TaintTracer

async def main():
    async with TaintTracer() as tool:
        res = await tool.execute(source=r"user_input", sink=r"os/exec", max_paths=2)
        print(json.dumps(res, ensure_ascii=False))

if __name__ == "__main__":
    asyncio.run(main())