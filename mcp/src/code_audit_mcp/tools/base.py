"""Base classes for MCP tools"""

import asyncio
import logging
from abc import ABC, abstractmethod
from typing import Any, Dict, List, Optional

import grpc
from ..proto import (
    ast_parser_pb2,
    ast_parser_pb2_grpc,
    indexer_pb2,
    indexer_pb2_grpc,
    taint_analysis_pb2,
    taint_analysis_pb2_grpc,
    call_chain_pb2,
    call_chain_pb2_grpc,
)

logger = logging.getLogger(__name__)


class GRPCClient:
    """Base gRPC client for connecting to Go backend"""
    
    def __init__(self, host: str = "localhost", port: int = 50051):
        self.host = host
        self.port = port
        self.channel = None
        
    async def connect(self):
        """Establish gRPC connection"""
        if self.channel is None:
            self.channel = grpc.aio.insecure_channel(f"{self.host}:{self.port}")
            
    async def disconnect(self):
        """Close gRPC connection"""
        if self.channel:
            await self.channel.close()
            self.channel = None
            
    def get_ast_parser_stub(self):
        """Get AST parser service stub"""
        return ast_parser_pb2_grpc.ASTParserStub(self.channel)
        
    def get_indexer_stub(self):
        """Get indexer service stub"""
        return indexer_pb2_grpc.IndexerStub(self.channel)
        
    def get_taint_analyzer_stub(self):
        """Get taint analyzer service stub"""
        return taint_analysis_pb2_grpc.TaintAnalyzerStub(self.channel)
        
    def get_call_chain_stub(self):
        """Get call chain analyzer service stub"""
        return call_chain_pb2_grpc.CallChainAnalyzerStub(self.channel)


class BaseTool(ABC):
    """Base class for all MCP tools"""
    
    def __init__(self, grpc_client: Optional[GRPCClient] = None):
        self.grpc_client = grpc_client or GRPCClient()
        self.logger = logging.getLogger(self.__class__.__name__)
        
    async def __aenter__(self):
        await self.grpc_client.connect()
        return self
        
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.grpc_client.disconnect()
        
    @abstractmethod
    async def execute(self, **kwargs) -> Dict[str, Any]:
        """Execute the tool with given arguments"""
        pass
        
    def format_error(self, error: Exception) -> Dict[str, Any]:
        """Format error response"""
        return {
            "success": False,
            "error": str(error),
            "type": type(error).__name__
        }
        
    def format_success(self, data: Any) -> Dict[str, Any]:
        """Format success response"""
        return {
            "success": True,
            "data": data
        }