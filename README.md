# 大陆农机配件

可运行 MVP：Gin + GORM + MySQL 后端，React + TypeScript + Vite 前台与后台 CMS。

## 环境要求

- Go 1.25+
- Node.js 24+
- MySQL 5.7，监听 `127.0.0.1:13306`

## 数据库

```sql
CREATE DATABASE IF NOT EXISTS dl_nongji_parts DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

默认连接：

```text
root:root@127.0.0.1:13306/dl_nongji_parts
```

可用环境变量覆盖：`DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASSWORD`、`DB_NAME`、`HTTP_ADDR`。

## 启动后端

```powershell
cd backend
go run ./cmd/server
```

后端启动后会自动迁移表结构并写入首页、菜单、厂商、产品、Banner、统计和默认管理员种子数据。

## 启动前端

```powershell
cd frontend
npm install
npm run dev
```

前端默认访问后端 `http://127.0.0.1:8080`。

## 后台

访问：

```text
/admin/login
```

默认账号：

```text
admin / admin123
```

## 验证

```powershell
cd backend
$env:GOTMPDIR="$PWD\.gotmp"
go test ./...

cd ..\frontend
npm run test
npm run build
```
