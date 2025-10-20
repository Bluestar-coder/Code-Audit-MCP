"""MCP Server implementation for code audit"""

import asyncio
import logging
import json
from typing import Any

import mcp.types as types
from mcp.server import Server
from mcp.server.stdio import stdio_server

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Create MCP server application
app = Server("code-audit-mcp")


@app.list_tools()
async def list_tools() -> list[types.Tool]:
    """List available MCP tools"""
    return [
        types.Tool(
            name="scan_code",
            description="Scan codebase for security vulnerabilities",
            inputSchema={
                "type": "object",
                "properties": {
                    "path": {
                        "type": "string",
                        "description": "Path to code directory or file",
                    },
                    "language": {
                        "type": "string",
                        "description": "Programming language (auto-detect if not specified)",
                    },
                },
                "required": ["path"],
            },
        ),
        types.Tool(
            name="generate_poc",
            description="Generate POC (Proof of Concept) code for a vulnerability",
            inputSchema={
                "type": "object",
                "properties": {
                    "vulnerability_id": {
                        "type": "string",
                        "description": "Vulnerability ID to generate POC for",
                    },
                    "language": {
                        "type": "string",
                        "description": "Target language for POC",
                    },
                    "context": {
                        "type": "string",
                        "description": "Optional code context or snippet to guide PoC",
                    },
                },
                "required": ["vulnerability_id"],
            },
        ),
        types.Tool(
            name="analyze_call_graph",
            description="Analyze function call chains and dependencies",
            inputSchema={
                "type": "object",
                "properties": {
                    "path": {
                        "type": "string",
                        "description": "Path to code directory",
                    },
                    "entry_point": {
                        "type": "string",
                        "description": "Entry function to analyze from",
                    },
                },
                "required": ["path"],
            },
        ),
        types.Tool(
            name="trace_taint",
            description="Trace data flow from source to sink",
            inputSchema={
                "type": "object",
                "properties": {
                    "file_path": {
                        "type": "string",
                        "description": "Path to source file",
                    },
                    "source": {
                        "type": "string",
                        "description": "Source variable or function",
                    },
                    "sink": {
                        "type": "string",
                        "description": "Sink variable or function",
                    },
                },
                "required": ["file_path", "source", "sink"],
            },
        ),
        types.Tool(
            name="explain_code",
            description="AI-powered code explanation and analysis",
            inputSchema={
                "type": "object",
                "properties": {
                    "code": {
                        "type": "string",
                        "description": "Code snippet to explain",
                    },
                    "language": {
                        "type": "string",
                        "description": "Programming language",
                    },
                },
                "required": ["code"],
            },
        ),
        types.Tool(
            name="search_vulnerabilities",
            description="Search for known vulnerabilities in dependencies",
            inputSchema={
                "type": "object",
                "properties": {
                    "package_name": {
                        "type": "string",
                        "description": "Package name to search",
                    },
                    "version": {
                        "type": "string",
                        "description": "Package version (optional)",
                    },
                    "ecosystem": {
                        "type": "string",
                        "description": "Package ecosystem (e.g., Go, PyPI, npm)",
                    },
                },
                "required": ["package_name"],
            },
        ),
    ]


@app.call_tool()
async def call_tool(name: str, arguments: dict[str, Any]) -> list[types.TextContent]:
    """Handle tool calls"""
    logger.info(f"Tool called: {name} with arguments: {arguments}")

    if name == "scan_code":
        from .tools import CodeScanner
        path = arguments.get("path")
        language = arguments.get("language")
        async with CodeScanner() as tool:
            result = await tool.execute(path=path, language=language)
        return [types.TextContent(type="text", text=json.dumps(result, ensure_ascii=False, indent=2))]

    elif name == "generate_poc":
        from .tools import POCGenerator
        vuln_id = arguments.get("vulnerability_id")
        language = arguments.get("language", "python")
        context = arguments.get("context")
        async with POCGenerator() as tool:
            result = await tool.execute(vulnerability_id=vuln_id, language=language, context=context)
        return [types.TextContent(type="text", text=json.dumps(result, ensure_ascii=False, indent=2))]

    elif name == "analyze_call_graph":
        from .tools import CallGraphAnalyzer
        path = arguments.get("path")
        entry_point = arguments.get("entry_point")
        async with CallGraphAnalyzer() as tool:
            result = await tool.execute(path=path, entry_point=entry_point)
        return [types.TextContent(type="text", text=json.dumps(result, ensure_ascii=False, indent=2))]

    elif name == "trace_taint":
        from .tools import TaintTracer
        source = arguments.get("source")
        sink = arguments.get("sink")
        async with TaintTracer() as tool:
            result = await tool.execute(source=source, sink=sink)
        return [types.TextContent(type="text", text=json.dumps(result, ensure_ascii=False, indent=2))]

    elif name == "explain_code":
        from .tools import CodeExplainer
        code = arguments.get("code")
        language = arguments.get("language", "")
        async with CodeExplainer() as tool:
            result = await tool.execute(code=code, language=language)
        return [types.TextContent(type="text", text=json.dumps(result, ensure_ascii=False, indent=2))]

    elif name == "search_vulnerabilities":
        from .tools import VulnerabilitySearcher
        package = arguments.get("package_name")
        version = arguments.get("version", "latest")
        ecosystem = arguments.get("ecosystem")
        async with VulnerabilitySearcher() as tool:
            result = await tool.execute(package_name=package, version=version, ecosystem=ecosystem)
        return [types.TextContent(type="text", text=json.dumps(result, ensure_ascii=False, indent=2))]

    else:
        raise ValueError(f"Unknown tool: {name}")


async def main() -> None:
    """Main entry point"""
    logger.info("Starting Code Audit MCP Server...")
    async with stdio_server() as (read_stream, write_stream):
        await app.run(read_stream, write_stream)


if __name__ == "__main__":
    asyncio.run(main())
