#!/usr/bin/env python3
"""
TypeScript解析器功能测试脚本
"""

import grpc
import sys
import os

# 添加项目根目录到路径
project_root = os.path.dirname(__file__)
sys.path.insert(0, project_root)

from src.code_audit_mcp.proto import ast_parser_pb2
from src.code_audit_mcp.proto import ast_parser_pb2_grpc

def test_typescript_parser():
    """测试TypeScript解析器功能"""
    
    # 连接到gRPC服务器
    channel = grpc.insecure_channel('localhost:50051')
    stub = ast_parser_pb2_grpc.ASTParserStub(channel)
    
    # TypeScript测试代码
    typescript_code = """
// TypeScript接口定义
interface User {
    id: number;
    name: string;
    email?: string;
}

// 类型别名
type UserRole = 'admin' | 'user' | 'guest';

// 枚举
enum Status {
    Active = 1,
    Inactive = 0,
    Pending = 2
}

// 类定义
class UserService {
    private users: User[] = [];
    
    constructor(private apiUrl: string) {}
    
    async getUser(id: number): Promise<User | null> {
        const response = await fetch(`${this.apiUrl}/users/${id}`);
        return response.json();
    }
    
    addUser(user: User): void {
        this.users.push(user);
    }
}

// 函数定义
function validateUser(user: User): boolean {
    return user.id > 0 && user.name.length > 0;
}

// 箭头函数
const formatUser = (user: User): string => {
    return `${user.name} (${user.id})`;
};

// 命名空间
namespace Utils {
    export function log(message: string): void {
        console.log(`[LOG] ${message}`);
    }
}

// 导入导出
export { User, UserService };
export default UserService;
"""
    
    print("=== TypeScript解析器测试 ===")
    
    # 测试单个文件解析
    print("\n1. 测试单个TypeScript文件解析:")
    request = ast_parser_pb2.ParseRequest(
        content=typescript_code.encode('utf-8'),
        file_path="test.ts",
        language="typescript"
    )
    
    try:
        response = stub.ParseFile(request)
        print(f"解析成功: {response.success}")
        
        if response.success:
            print(f"📁 文件路径: {response.file_path}")
            
            # 打印元数据
            if response.metadata:
                print(f"📊 代码行数: {response.metadata.total_lines}")
                print(f"🔧 函数数量: {response.metadata.total_functions}")
                print(f"🏗️ 类数量: {response.metadata.total_classes}")
                print(f"📝 语言版本: {response.metadata.language_version}")
                print(f"⏱️ 解析时间: {response.metadata.parse_time_ms}ms")
            
            # 打印AST数据信息
            if response.ast_data:
                import json
                try:
                    ast_json = json.loads(response.ast_data.decode('utf-8'))
                    print(f"🌳 AST根节点类型: {ast_json.get('type', 'unknown')}")
                    children = ast_json.get('children', [])
                    print(f"📋 子节点数量: {len(children)}")
                    
                    # 打印前几个子节点
                    for i, child in enumerate(children[:5]):
                        print(f"  📄 子节点{i+1}: {child.get('type', 'unknown')}")
                        if child.get('attributes'):
                            for key, value in child['attributes'].items():
                                print(f"    {key}: {value}")
                except json.JSONDecodeError:
                    print("❌ AST数据解析失败")
        else:
            print("解析失败")
            for error in response.errors:
                print(f"错误: {error.message} (行 {error.line}, 列 {error.column})")
    
    except grpc.RpcError as e:
        print(f"gRPC错误: {e}")
        return
    
    # 测试批量解析
    print("\n2. 测试批量TypeScript文件解析:")
    
    # 准备多个TypeScript文件
    files = [
        ("interface.ts", """
interface Product {
    id: string;
    name: string;
    price: number;
}
"""),
        ("service.ts", """
class ProductService {
    getProduct(id: string): Product {
        return { id, name: "Test", price: 100 };
    }
}
"""),
        ("utils.ts", """
export function formatPrice(price: number): string {
    return `$${price.toFixed(2)}`;
}
""")
    ]
    
    batch_request = ast_parser_pb2.BatchParseRequest()
    for file_path, code in files:
        parse_request = ast_parser_pb2.ParseRequest(
            content=code.encode('utf-8'),
            file_path=file_path,
            language="typescript"
        )
        batch_request.requests.append(parse_request)
    
    try:
        results = []
        for response in stub.ParseBatch(batch_request):
            results.append(response)
        
        print(f"批量解析完成，处理了 {len(results)} 个文件:")
        for i, result in enumerate(results):
            print(f"  文件 {i+1}: {result.file_path} - {'成功' if result.success else '失败'}")
            if result.metadata:
                print(f"    函数数: {result.metadata.total_functions}, 类数: {result.metadata.total_classes}")
    
    except grpc.RpcError as e:
        print(f"批量解析gRPC错误: {e}")
    
    print("\n=== TypeScript解析器测试完成 ===")

if __name__ == "__main__":
    test_typescript_parser()