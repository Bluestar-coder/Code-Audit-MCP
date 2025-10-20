import asyncio
import json
from code_audit_mcp.tools.poc_generator import POCGenerator

async def main():
    async with POCGenerator() as tool:
        res = await tool.execute(vulnerability_id="CVE-2023-12345", language="go", context="func vulnerable(){ /* ... */ }")
        print(json.dumps(res, ensure_ascii=False))

if __name__ == "__main__":
    asyncio.run(main())