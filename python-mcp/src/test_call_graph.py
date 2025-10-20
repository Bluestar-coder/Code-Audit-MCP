import asyncio
import json
from code_audit_mcp.tools.call_graph_analyzer import CallGraphAnalyzer

async def main():
    async with CallGraphAnalyzer() as tool:
        res = await tool.execute(path=r"E:\Code\CodeAuditMcp\go-backend\internal\indexer\service.go", max_depth=3)
        print(json.dumps(res, ensure_ascii=False))

if __name__ == "__main__":
    asyncio.run(main())