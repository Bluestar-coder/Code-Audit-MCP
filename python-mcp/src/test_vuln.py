import asyncio
import json
from code_audit_mcp.tools.vulnerability_searcher import VulnerabilitySearcher

async def main():
    async with VulnerabilitySearcher() as tool:
        res = await tool.execute(package_name="github.com/gin-gonic/gin", version="v1.9.0", ecosystem="Go")
        print(json.dumps(res, ensure_ascii=False))

if __name__ == "__main__":
    asyncio.run(main())