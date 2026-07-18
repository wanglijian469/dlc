# 大陆农机配件

可运行 MVP：Gin + GORM + MySQL 后端，React + TypeScript + Vite 前台与后台 CMS。

## 环境要求

- Go 1.25+
- Node.js 24+
- MySQL 5.7，默认监听 `127.0.0.1:13306`

## 数据库

```sql
CREATE DATABASE IF NOT EXISTS dl_nongji_parts DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

默认连接：

```text
root:root@127.0.0.1:13306/dl_nongji_parts
```

可用环境变量覆盖：`DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASSWORD`、`DB_NAME`、`HTTP_ADDR`、`PUBLIC_DIR`。

## 本地开发

启动后端：

```powershell
cd backend
go run ./cmd/server
```

启动前端：

```powershell
cd frontend
npm install
npm run dev
```

开发模式下，Vite 会把 `/api` 代理到 `http://127.0.0.1:8080`。生产打包后，前端默认同源访问 `/api`，不会请求访问者本机的 `127.0.0.1:8080`。

## Windows 部署包

生成可执行部署包：

```powershell
powershell -ExecutionPolicy Bypass -File scripts\package-windows.ps1
```

输出文件位于 `release/dlc-mvp-<version>.zip`。部署包包含：

- `server.exe`
- `frontend/dist`
- `.env.example`
- `scripts/init-db.sql`
- `start-windows.ps1`
- `GATEWAY_GIN.md`（云网关/CDN 统一回源 Gin；不部署 Nginx）
- `DEPLOY.md`

解压后复制 `.env.example` 为 `.env`，按服务器数据库信息修改配置，然后执行：

```powershell
powershell -ExecutionPolicy Bypass -File .\start-windows.ps1
```

## 后台

访问：`/admin/login`

默认账号：

```text
admin / admin123
```

## 验证

```powershell
cd backend
go test ./...

cd ..\frontend
npm run test
npm run build
```
