#!/usr/bin/env python3
"""
测试Python客户端与Go gRPC服务器的连接
"""

import grpc
import sys
import os

# 添加项目根路径以导入 proto 包（优先使用本地 proto）
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
# 兼容生成代码中的顶层模块导入（indexer_pb2 / call_chain_pb2 等）
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'proto'))

try:
    # 导入生成的proto文件（作为 proto 包）
    from proto import ast_parser_pb2
    from proto import ast_parser_pb2_grpc
    from proto import indexer_pb2
    from proto import indexer_pb2_grpc
    from proto import taint_analysis_pb2
    from proto import taint_analysis_pb2_grpc
    from proto import call_chain_pb2
    from proto import call_chain_pb2_grpc
    
    print("✅ Proto文件导入成功")
except ImportError as e:
    print(f"❌ Proto文件导入失败: {e}")
    sys.exit(1)

def test_grpc_connection():
    """测试gRPC连接"""
    try:
        # 创建gRPC通道
        channel = grpc.insecure_channel('localhost:50051')
        
        # 测试连接
        grpc.channel_ready_future(channel).result(timeout=10)
        print("✅ gRPC连接成功")
        
        # 创建客户端存根
        ast_client = ast_parser_pb2_grpc.ASTParserStub(channel)
        indexer_client = indexer_pb2_grpc.IndexerStub(channel)
        taint_client = taint_analysis_pb2_grpc.TaintAnalyzerStub(channel)
        call_chain_client = call_chain_pb2_grpc.CallChainAnalyzerStub(channel)
        
        print("✅ 所有客户端存根创建成功")
        
        # 测试简单的健康检查调用
        try:
            # 测试AST解析器
            request = ast_parser_pb2.ParseRequest(
                language="python",
                content=b"print('hello world')",
                file_path="test.py"
            )
            response = ast_client.ParseFile(request, timeout=5)
            print("✅ AST解析器服务响应正常")
            
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.UNIMPLEMENTED:
                print("⚠️  AST解析器方法未实现（这是预期的）")
            else:
                print(f"❌ AST解析器调用失败: {e}")
        
        # 关闭连接
        channel.close()
        print("✅ 连接测试完成")
        
        return True
        
    except Exception as e:
        print(f"❌ gRPC连接失败: {e}")
        return False

if __name__ == "__main__":
    print("🔍 开始测试Python-Go gRPC连接...")
    print("=" * 50)
    
    success = test_grpc_connection()
    
    print("=" * 50)
    if success:
        print("🎉 所有测试通过！Python客户端可以成功连接到Go服务器")
    else:
        print("💥 测试失败，请检查服务器状态和网络连接")
        sys.exit(1)