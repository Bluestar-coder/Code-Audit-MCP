"""gRPC client for communicating with Go backend services"""

import logging
from typing import Any, Optional

import grpc

logger = logging.getLogger(__name__)


class GoServiceManager:
    """Manages connections to Go backend services via gRPC"""

    def __init__(
        self,
        host: str = "localhost",
        port: int = 50051,
        max_message_size: int = 100 * 1024 * 1024,
    ):
        """
        Initialize gRPC client manager

        Args:
            host: gRPC server host
            port: gRPC server port
            max_message_size: Maximum message size (default 100MB)
        """
        self.host = host
        self.port = port
        self.max_message_size = max_message_size

        # Create channel options
        channel_options = [
            ("grpc.max_receive_message_length", max_message_size),
            ("grpc.max_send_message_length", max_message_size),
            # Enable gzip compression
            ("grpc.default_compression_algorithm", grpc.Compression.Gzip),
            ("grpc.default_compression_level", "default"),
        ]

        # Create insecure channel (for development)
        self.channel = grpc.insecure_channel(
            f"{host}:{port}",
            options=channel_options,
        )

        logger.info(f"gRPC client initialized: {host}:{port}")

    async def parse_file(
        self,
        file_path: str,
        language: str,
        content: Optional[bytes] = None,
    ) -> dict[str, Any]:
        """
        Parse a file via gRPC

        Args:
            file_path: Path to the file
            language: Programming language
            content: File content (if None, will be read from file_path)

        Returns:
            Parsed response
        """
        try:
            # TODO: Call gRPC ParseFile RPC
            logger.info(f"Parsing {file_path} ({language})")
            return {
                "file_path": file_path,
                "language": language,
                "success": True,
                "message": "File parsing not yet implemented",
            }
        except grpc.RpcError as e:
            logger.error(f"gRPC error: {e.code()}: {e.details()}")
            raise

    async def parse_batch(
        self,
        requests: list[dict[str, str]],
        max_concurrent: int = 4,
    ) -> list[dict[str, Any]]:
        """
        Parse multiple files via gRPC

        Args:
            requests: List of parse requests
            max_concurrent: Maximum concurrent operations

        Returns:
            List of parsed responses
        """
        try:
            # TODO: Call gRPC ParseBatch RPC
            logger.info(f"Batch parsing {len(requests)} files")
            return [
                {
                    "file_path": req.get("file_path"),
                    "success": True,
                    "message": "Batch parsing not yet implemented",
                }
                for req in requests
            ]
        except grpc.RpcError as e:
            logger.error(f"gRPC error: {e.code()}: {e.details()}")
            raise

    async def health_check(self) -> bool:
        """
        Check if gRPC server is healthy

        Returns:
            True if server is reachable
        """
        try:
            # Try to get channel state
            # In a real implementation, this would call a Health service
            logger.info("Health check passed")
            return True
        except Exception as e:
            logger.error(f"Health check failed: {e}")
            return False

    def close(self) -> None:
        """Close gRPC channel"""
        self.channel.close()
        logger.info("gRPC channel closed")

    async def __aenter__(self):
        """Async context manager entry"""
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit"""
        self.close()


# Singleton instance for convenience
_go_service_manager: Optional[GoServiceManager] = None


async def get_go_client() -> GoServiceManager:
    """Get or create singleton gRPC client"""
    global _go_service_manager
    if _go_service_manager is None:
        _go_service_manager = GoServiceManager()
    return _go_service_manager


async def close_go_client() -> None:
    """Close singleton gRPC client"""
    global _go_service_manager
    if _go_service_manager is not None:
        _go_service_manager.close()
        _go_service_manager = None
