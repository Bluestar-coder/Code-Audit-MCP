from typing import Any, Dict

from .base import BaseTool
from ..proto import taint_analysis_pb2, taint_analysis_pb2_grpc


class TaintTracer(BaseTool):
    async def execute(self, source: str, sink: str, max_paths: int = 3) -> Dict[str, Any]:
        try:
            stub = taint_analysis_pb2_grpc.TaintAnalyzerStub(self.grpc_client.channel)
            req = taint_analysis_pb2.TracePathRequest(
                source_function=source,
                sink_function=sink,
                max_paths=max_paths,
            )
            # Streaming response
            paths = []
            async for segment in stub.TracePath(req):
                nodes = [
                    {
                        "node_id": n.node_id,
                        "function_name": n.function_name,
                        "file_path": n.file_path,
                        "line_number": n.line_number,
                        "operation": n.operation,
                        "variable_name": n.variable_name,
                        "data_flow": n.data_flow,
                    }
                    for n in segment.nodes
                ]
                paths.append({
                    "path_index": segment.path_index,
                    "nodes": nodes,
                    "has_sanitizer": segment.has_sanitizer,
                })
            return self.format_success({
                "paths": paths,
                "total_paths": len(paths),
            })
        except Exception as e:
            return self.format_error(e)