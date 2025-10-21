# CodeAuditMcp 发布包使用说明

本发布包包含：
- 后端可执行文件：`server.exe`（gRPC: `50051`, HTTP: `8080`）
- 前端静态资源：`frontend/`（React 生产构建）
- 规则文件：`rules/common/*.yaml`

## 快速启动

1) 启动后端服务

在 `release/` 目录执行：

```
./server.exe
```

启动成功后，HTTP 健康检查地址：

```
http://localhost:8080/api/health
```

2) 启动前端静态资源（任选其一）

- 方式 A：使用 Node 的 `serve`（需要安装 Node.js）
```
npm install -g serve
serve -s frontend
```
然后打开输出的地址（默认 `http://localhost:3000`）。

- 方式 B：使用任意静态服务器（如 Nginx、IIS、`python -m http.server` 等）
将根目录指向 `release/frontend` 即可。

## 配置说明

- 规则文件位于 `release/rules/common/`，可根据需要新增或修改 YAML 规则。
- 前端默认按 `/` 路径构建。如需部署到子路径，请在 `package.json` 的 `homepage` 字段设置部署路径后重新构建前端。
- 后端默认监听 `:8080` 和 `:50051`，如需更改端口，请使用系统环境变量或在代码中调整后重新构建。

## 目录结构
```
release/
  ├─ server.exe
  ├─ frontend/
  └─ rules/
      └─ common/
          ├─ xss.yaml
          ├─ sql_injection.yaml
          └─ path_traversal.yaml
```

## 常见问题
- 页面无法请求到后端：请确认后端已启动，且前端的请求地址指向 `http://localhost:8080`（如有反向代理，请检查代理配置）。
- 部署到子路径后资源404：为 React 应用，请确保静态服务器对前端路由进行回退（fallback）到 `index.html`。

## 版权与安全
- 请参考仓库中的 `SECURITY.md` 与 `CONTRIBUTING.md` 以了解安全与贡献规范。
- 发布前建议检查与清理多余资源，避免包含开发缓存文件。