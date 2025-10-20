#!/usr/bin/env python3
"""
详细的AST解析器功能测试
测试Go和Python代码的AST解析功能
"""

import grpc
import json
import sys
import os

# 添加proto生成的模块路径
sys.path.append(os.path.join(os.path.dirname(__file__), 'src'))

from code_audit_mcp.proto import ast_parser_pb2
from code_audit_mcp.proto import ast_parser_pb2_grpc

def test_go_code_parsing():
    """测试Go代码解析"""
    print("🔍 测试Go代码解析...")
    
    go_code = '''package main

import (
    "fmt"
    "log"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func (u *User) GetInfo() string {
    return fmt.Sprintf("User: %s, Age: %d", u.Name, u.Age)
}

func main() {
    user := &User{
        ID:   1,
        Name: "Alice",
        Age:  30,
    }
    
    fmt.Println(user.GetInfo())
    log.Printf("User created: %+v", user)
}
'''
    
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = ast_parser_pb2_grpc.ASTParserStub(channel)
        
        request = ast_parser_pb2.ParseRequest(
            file_path="test.go",
            language="go",
            content=go_code.encode('utf-8'),
            include_metadata=True
        )
        
        try:
            response = stub.ParseFile(request)
            
            print(f"  📁 文件: {response.file_path}")
            print(f"  ✅ 解析成功: {response.success}")
            
            if response.success:
                # 解析AST数据
                if response.ast_data:
                    ast_json = json.loads(response.ast_data)
                    print(f"  🌳 AST根节点类型: {ast_json.get('type', 'unknown')}")
                    print(f"  📊 子节点数量: {len(ast_json.get('children', []))}")
                
                # 显示元数据
                if response.metadata:
                    print(f"  ⏱️  解析时间: {response.metadata.parse_time_ms}ms")
                    print(f"  📝 总行数: {response.metadata.total_lines}")
                    print(f"  🔧 函数数量: {response.metadata.total_functions}")
                    print(f"  📦 类/结构体数量: {response.metadata.total_classes}")
                    print(f"  🏷️  语言版本: {response.metadata.language_version}")
            
            if response.errors:
                print(f"  ⚠️  错误数量: {len(response.errors)}")
                for error in response.errors:
                    print(f"    - 行{error.line}列{error.column}: {error.message}")
            
            return response.success
            
        except grpc.RpcError as e:
            print(f"  ❌ gRPC错误: {e}")
            return False

def test_python_code_parsing():
    """测试Python代码解析"""
    print("\n🔍 测试Python代码解析...")
    
    python_code = '''#!/usr/bin/env python3
"""
示例Python模块
包含类、函数和变量定义
"""

import os
import sys
from typing import List, Dict, Optional

class DataProcessor:
    """数据处理器类"""
    
    def __init__(self, name: str):
        self.name = name
        self.data: List[Dict] = []
    
    def add_data(self, item: Dict) -> None:
        """添加数据项"""
        if self._validate_item(item):
            self.data.append(item)
    
    def _validate_item(self, item: Dict) -> bool:
        """验证数据项"""
        return isinstance(item, dict) and 'id' in item
    
    def process_data(self) -> List[Dict]:
        """处理数据"""
        processed = []
        for item in self.data:
            processed_item = self._transform_item(item)
            processed.append(processed_item)
        return processed
    
    def _transform_item(self, item: Dict) -> Dict:
        """转换数据项"""
        return {
            'id': item['id'],
            'processed': True,
            'timestamp': time.time()
        }

def main():
    """主函数"""
    processor = DataProcessor("test_processor")
    
    # 添加测试数据
    test_data = [
        {'id': 1, 'value': 'test1'},
        {'id': 2, 'value': 'test2'},
        {'id': 3, 'value': 'test3'}
    ]
    
    for item in test_data:
        processor.add_data(item)
    
    # 处理数据
    result = processor.process_data()
    print(f"处理了 {len(result)} 个数据项")

if __name__ == "__main__":
    main()
'''
    
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = ast_parser_pb2_grpc.ASTParserStub(channel)
        
        request = ast_parser_pb2.ParseRequest(
            file_path="test.py",
            language="python",
            content=python_code.encode('utf-8'),
            include_metadata=True
        )
        
        try:
            response = stub.ParseFile(request)
            
            print(f"  📁 文件: {response.file_path}")
            print(f"  ✅ 解析成功: {response.success}")
            
            if response.success:
                # 解析AST数据
                if response.ast_data:
                    ast_json = json.loads(response.ast_data)
                    print(f"  🌳 AST根节点类型: {ast_json.get('type', 'unknown')}")
                    print(f"  📊 子节点数量: {len(ast_json.get('children', []))}")
                
                # 显示元数据
                if response.metadata:
                    print(f"  ⏱️  解析时间: {response.metadata.parse_time_ms}ms")
                    print(f"  📝 总行数: {response.metadata.total_lines}")
                    print(f"  🔧 函数数量: {response.metadata.total_functions}")
                    print(f"  📦 类数量: {response.metadata.total_classes}")
                    print(f"  🏷️  语言版本: {response.metadata.language_version}")
            
            if response.errors:
                print(f"  ⚠️  错误数量: {len(response.errors)}")
                for error in response.errors:
                    print(f"    - 行{error.line}列{error.column}: {error.message}")
            
            return response.success
            
        except grpc.RpcError as e:
            print(f"  ❌ gRPC错误: {e}")
            return False

def test_batch_parsing():
    """测试批量解析"""
    print("\n🔍 测试批量解析...")
    
    files = [
        {
            'path': 'simple.go',
            'language': 'go',
            'content': 'package main\n\nfunc main() {\n    println("Hello, World!")\n}'
        },
        {
            'path': 'simple.py',
            'language': 'python',
            'content': '#!/usr/bin/env python3\n\ndef hello():\n    print("Hello, World!")\n\nif __name__ == "__main__":\n    hello()'
        }
    ]
    
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = ast_parser_pb2_grpc.ASTParserStub(channel)
        
        # 创建批量请求
        requests = []
        for file_info in files:
            request = ast_parser_pb2.ParseRequest(
                file_path=file_info['path'],
                language=file_info['language'],
                content=file_info['content'].encode('utf-8'),
                include_metadata=True
            )
            requests.append(request)
        
        batch_request = ast_parser_pb2.BatchParseRequest(requests=requests)
        
        try:
            response_stream = stub.ParseBatch(batch_request)
            
            success_count = 0
            total_count = 0
            
            for response in response_stream:
                total_count += 1
                print(f"  📁 文件 {total_count}: {response.file_path}")
                print(f"    ✅ 解析成功: {response.success}")
                
                if response.success:
                    success_count += 1
                    if response.metadata:
                        print(f"    ⏱️  解析时间: {response.metadata.parse_time_ms}ms")
                        print(f"    📝 总行数: {response.metadata.total_lines}")
                        print(f"    🔧 函数数量: {response.metadata.total_functions}")
                
                if response.errors:
                    print(f"    ⚠️  错误数量: {len(response.errors)}")
            
            print(f"\n  📊 批量解析结果: {success_count}/{total_count} 成功")
            return success_count == total_count
            
        except grpc.RpcError as e:
            print(f"  ❌ gRPC错误: {e}")
            return False

def main():
    """主测试函数"""
    print("🧪 开始AST解析器详细功能测试...")
    print("=" * 60)
    
    tests = [
        ("Go代码解析", test_go_code_parsing),
        ("Python代码解析", test_python_code_parsing),
        ("批量解析", test_batch_parsing)
    ]
    
    passed = 0
    total = len(tests)
    
    for test_name, test_func in tests:
        print(f"\n🔬 运行测试: {test_name}")
        try:
            if test_func():
                print(f"✅ {test_name} - 通过")
                passed += 1
            else:
                print(f"❌ {test_name} - 失败")
        except Exception as e:
            print(f"❌ {test_name} - 异常: {e}")
    
    print("\n" + "=" * 60)
    print(f"🎯 测试结果: {passed}/{total} 通过")
    
    if passed == total:
        print("🎉 所有AST解析器测试通过！")
        return True
    else:
        print("⚠️  部分测试失败，请检查实现")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)