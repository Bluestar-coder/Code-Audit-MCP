import os
import json
import asyncio
from typing import Any, Dict, List, Optional

from .base import BaseTool
from ..proto import indexer_pb2, indexer_pb2_grpc

LANG_EXT = {
    ".go": "go",
    ".js": "javascript",
    ".jsx": "javascript",
    ".ts": "typescript",
    ".tsx": "typescript",
    ".py": "python",
}


def detect_language(file_path: str) -> Optional[str]:
    _, ext = os.path.splitext(file_path.lower())
    return LANG_EXT.get(ext)


class CodeScanner(BaseTool):
    async def execute(self, path: str, language: Optional[str] = None, incremental: bool = True) -> Dict[str, Any]:
        try:
            # Collect files
            files: List[str] = []
            if os.path.isdir(path):
                for root, _, filenames in os.walk(path):
                    for name in filenames:
                        fp = os.path.join(root, name)
                        if detect_language(fp):
                            files.append(fp)
            else:
                files.append(path)
            
            if not files:
                return self.format_success({
                    "files_scanned": 0,
                    "message": "No source files found",
                })
            
            stub = indexer_pb2_grpc.IndexerStub(self.grpc_client.channel)
            
            async def scan_file(fp: str) -> Dict[str, Any]:
                lang = language or detect_language(fp) or "unknown"
                req = indexer_pb2.BuildIndexRequest(
                    file_path=fp,
                    language=lang,
                    ast_data=b"",
                    incremental=incremental,
                )
                resp = await stub.BuildIndex(req)
                return {
                    "file_path": fp,
                    "language": lang,
                    "indexed": resp.success,
                    "functions_indexed": resp.functions_indexed,
                    "classes_indexed": resp.classes_indexed,
                    "variables_indexed": resp.variables_indexed,
                    "index_id": resp.index_id,
                    "error": resp.error_message,
                }
            
            # Run with limited concurrency
            semaphore = asyncio.Semaphore(8)
            results: List[Dict[str, Any]] = []
            
            async def worker(fp: str):
                async with semaphore:
                    results.append(await scan_file(fp))
            
            await asyncio.gather(*(worker(fp) for fp in files))
            
            total_functions = sum(r.get("functions_indexed", 0) for r in results)
            total_classes = sum(r.get("classes_indexed", 0) for r in results)
            total_variables = sum(r.get("variables_indexed", 0) for r in results)
            successes = sum(1 for r in results if r.get("indexed"))
            
            return self.format_success({
                "files_scanned": len(files),
                "successes": successes,
                "totals": {
                    "functions": total_functions,
                    "classes": total_classes,
                    "variables": total_variables,
                },
                "results": results,
            })
        except Exception as e:
            return self.format_error(e)