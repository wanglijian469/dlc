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

可用环境变量覆盖：`DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASSWORD`、`DB_NAME`、`HTTP_ADDR`、`PUBLIC_DIR`、`STATIC_PAGES_ENABLED`、`STATIC_PAGE_DIR`。

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

平台管理员可在“平台配置 → 访问与采集防护”维护 AI 爬虫名单、IP/CIDR 白名单、慢速采集阈值、封禁记录和图片水印任务。默认采用“仅记录”模式；生产上线观察 7 天并确认搜索引擎访问正常后，再关闭“仅记录”以启用自动封禁。

Linux 入口限流和图片防盗链配置位于 `deploy/linux/nginx/dalu-parts.conf.template`。联系方式默认登录后可见，厂商提交公开范围变更后须经管理员审核。

## 验证

```powershell
cd backend
go test ./...

cd ..\frontend
npm run test
npm run build
```

## Monorepo 移动端

- `backend`：Gin + GORM + MySQL 公共后端，保留既有 `/api` 契约并新增 `/api/v1`。
- `frontend`：React PC 与同 URL 响应式移动网页（需求中的 `web`）。
- `mobile`：Android 首发、iOS 构建兼容的 Flutter App。
- `shared`：OpenAPI、JSON Schema 与脱敏契约样例，不存放 UI 或业务状态。

Flutter 首次准备与验证：

```powershell
cd mobile
flutter pub get
flutter analyze
flutter test
flutter build apk --debug --dart-define=API_BASE_URL=http://10.0.2.2:8080
```

仓库通过 `mobile/.fvmrc` 和 CI 固定 Flutter `3.44.9`。本地 Android 构建需先安装
Android SDK；CI 会同时生成 debug/release APK，并上传 release 构建产物。

供求信息通过 `/purchase` 在 PC 与移动网页共享同一 URL。采购商只能发布求购，
厂商只能发布供应；联系方式不在公开列表和详情响应中返回，必须登录后单独获取。
