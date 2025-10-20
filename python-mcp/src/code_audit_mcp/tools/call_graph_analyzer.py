from typing import Any, Dict, List, Optional

from .base import BaseTool
from ..proto import call_chain_pb2, call_chain_pb2_grpc


class CallGraphAnalyzer(BaseTool):
    async def execute(self, path: str, entry_point: Optional[str] = None, max_depth: int = 10) -> Dict[str, Any]:
        try:
            stub = call_chain_pb2_grpc.CallChainAnalyzerStub(self.grpc_client.channel)
            req = call_chain_pb2.BuildCallGraphRequest(
                file_path=path,
                entry_points=[entry_point] if entry_point else [],
                include_external=False,
                max_depth=max_depth,
            )
            resp = await stub.BuildCallGraph(req)
            return self.format_success({
                "graph_id": resp.graph_id,
                "success": resp.success,
                "node_count": resp.node_count,
                "edge_count": resp.edge_count,
                "build_time_ms": resp.build_time_ms,
                "error": resp.error_message,
            })
        except Exception as e:
            return self.format_error(e)