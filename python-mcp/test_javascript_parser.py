#!/usr/bin/env python3
"""
JavaScript AST解析器功能测试脚本
测试Go gRPC服务器的JavaScript代码解析功能
"""

import sys
import os

# 添加proto模块路径
sys.path.append(os.path.join(os.path.dirname(__file__), 'src'))

import grpc
from code_audit_mcp.proto import ast_parser_pb2
from code_audit_mcp.proto import ast_parser_pb2_grpc

def test_javascript_code_parsing():
    """测试JavaScript代码解析"""
    print("🧪 测试JavaScript代码解析...")
    
    # JavaScript代码示例
    js_code = '''
// 示例JavaScript代码
import React from 'react';
import { useState, useEffect } from 'react';

const MyComponent = () => {
    const [count, setCount] = useState(0);
    const [data, setData] = useState(null);
    
    useEffect(() => {
        fetchData();
    }, []);
    
    const fetchData = async () => {
        try {
            const response = await fetch('/api/data');
            const result = await response.json();
            setData(result);
        } catch (error) {
            console.error('Error fetching data:', error);
        }
    };
    
    const handleClick = () => {
        setCount(count + 1);
    };
    
    return (
        <div>
            <h1>Count: {count}</h1>
            <button onClick={handleClick}>Increment</button>
            {data && <pre>{JSON.stringify(data, null, 2)}</pre>}
        </div>
    );
};

export default MyComponent;

class DataService {
    constructor(baseUrl) {
        this.baseUrl = baseUrl;
    }
    
    async get(endpoint) {
        const url = `${this.baseUrl}${endpoint}`;
        const response = await fetch(url);
        return response.json();
    }
    
    async post(endpoint, data) {
        const url = `${this.baseUrl}${endpoint}`;
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });
        return response.json();
    }
}

export { DataService };
'''
    
    # 连接到gRPC服务器
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = ast_parser_pb2_grpc.ASTParserStub(channel)
        
        # 创建解析请求
        request = ast_parser_pb2.ParseRequest(
            file_path="test.js",
            content=js_code.encode('utf-8'),
            language="javascript"
        )
        
        try:
            # 发送解析请求
            response = stub.ParseFile(request)
            
            print(f"✅ 解析成功: {response.success}")
            print(f"📁 文件路径: {response.file_path}")
            
            if response.ast_data:
                print(f"🌳 AST数据长度: {len(response.ast_data)} 字节")
                # 显示AST数据的前200个字符
                ast_preview = response.ast_data[:200].decode('utf-8', errors='ignore')
                print(f"🔍 AST数据预览: {ast_preview}...")
            
            if response.errors:
                print(f"⚠️  解析错误 ({len(response.errors)}个):")
                for error in response.errors:
                    print(f"   - 行{error.line}列{error.column}: {error.message}")
            else:
                print("✅ 无解析错误")
            
            # 显示元数据
            if response.metadata:
                metadata = response.metadata
                print(f"📊 解析元数据:")
                print(f"   - 解析时间: {metadata.parse_time_ms}ms")
                print(f"   - 总行数: {metadata.total_lines}")
                print(f"   - 总函数数: {metadata.total_functions}")
                print(f"   - 总类数: {metadata.total_classes}")
                print(f"   - 语言版本: {metadata.language_version}")
            
            return True
            
        except grpc.RpcError as e:
            print(f"❌ gRPC错误: {e.code()} - {e.details()}")
            return False

def test_javascript_batch_parsing():
    """测试JavaScript批量解析"""
    print("\n🧪 测试JavaScript批量解析...")
    
    # 多个JavaScript文件示例
    js_files = [
        {
            "path": "utils.js",
            "content": '''
// 工具函数
export const formatDate = (date) => {
    return date.toISOString().split('T')[0];
};

export const debounce = (func, wait) => {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
};

export default { formatDate, debounce };
'''
        },
        {
            "path": "api.js", 
            "content": '''
// API客户端
class ApiClient {
    constructor(config) {
        this.baseURL = config.baseURL;
        this.timeout = config.timeout || 5000;
    }
    
    async request(method, url, data = null) {
        const config = {
            method,
            headers: {
                'Content-Type': 'application/json',
            },
        };
        
        if (data) {
            config.body = JSON.stringify(data);
        }
        
        const response = await fetch(`${this.baseURL}${url}`, config);
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return response.json();
    }
    
    get(url) {
        return this.request('GET', url);
    }
    
    post(url, data) {
        return this.request('POST', url, data);
    }
}

export default ApiClient;
'''
        }
    ]
    
    # 连接到gRPC服务器
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = ast_parser_pb2_grpc.ASTParserStub(channel)
        
        # 创建批量解析请求
        requests = []
        for file_info in js_files:
            request = ast_parser_pb2.ParseRequest(
                file_path=file_info["path"],
                content=file_info["content"].encode('utf-8'),
                language="javascript"
            )
            requests.append(request)
        
        batch_request = ast_parser_pb2.BatchParseRequest(
            requests=requests,
            max_concurrent=2
        )
        
        try:
            # 发送批量解析请求（返回流）
            response_stream = stub.ParseBatch(batch_request)
            
            results = []
            for response in response_stream:
                results.append(response)
            
            print(f"✅ 批量解析完成，处理了 {len(results)} 个文件")
            
            for i, result in enumerate(results):
                print(f"\n📄 文件 {i+1}: {result.file_path}")
                print(f"   ✅ 解析成功: {result.success}")
                
                if result.metadata:
                    metadata = result.metadata
                    print(f"   📊 元数据: {metadata.total_lines}行, {metadata.total_functions}函数, {metadata.total_classes}类")
                    print(f"   ⏱️  解析时间: {metadata.parse_time_ms}ms")
                
                if result.errors:
                    print(f"   ⚠️  错误数: {len(result.errors)}")
                else:
                    print(f"   ✅ 无错误")
            
            return True
            
        except grpc.RpcError as e:
            print(f"❌ gRPC错误: {e.code()} - {e.details()}")
            return False

def main():
    """主测试函数"""
    print("🚀 开始JavaScript AST解析器功能测试")
    print("=" * 50)
    
    # 测试单个文件解析
    success1 = test_javascript_code_parsing()
    
    # 测试批量解析
    success2 = test_javascript_batch_parsing()
    
    print("\n" + "=" * 50)
    if success1 and success2:
        print("🎉 所有JavaScript解析器测试通过！")
    else:
        print("❌ 部分测试失败")
        sys.exit(1)

if __name__ == "__main__":
    main()