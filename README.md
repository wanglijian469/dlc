# 大陆农机配件

可运行 MVP：Gin + GORM + MySQL 后端，React + TypeScript + Vite 前台与后台 CMS。

## 厂商推广工作台

厂商工作台、三步产品录入、私有草稿、企业展厅、本厂详情与推广统计已接入。数据库需从 Schema 19 增量升级至 20；见 [升级与验收说明](deploy/linux/VENDOR_PROMOTION_UPGRADE.md)。本次改动不会自动发布厂商资料。

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

智能采集默认关闭。管理员可在“平台配置 → 智能采集云服务”中填写腾讯云 OCR SecretId/SecretKey 与独立的 TokenHub API Key，密钥使用 `AUTH_SECRET` 派生密钥进行 AES-GCM 加密且不会通过读取接口回显。视觉理解通过 TokenHub OpenAI 兼容接口调用；部署环境也可配置 `TENCENT_CLOUD_SECRET_ID`、`TENCENT_CLOUD_SECRET_KEY`、`TENCENT_CLOUD_REGION`、`TENCENT_TOKENHUB_API_KEY`、`TENCENT_TOKENHUB_BASE_URL`、`CAPTURE_AI_ENABLED` 和 `CAPTURE_VISION_MODEL`，环境变量优先于后台配置。原始名片和彩页保存在 `MEDIA_DIR` 私有目录中，不通过前台媒体接口公开。

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

## Linux x86_64 部署包

在 Windows PowerShell 构建真正的 Linux ELF64 x86_64 升级包：

```powershell
powershell -ExecutionPolicy Bypass -File scripts\package-linux.ps1 -Version v2026.08.24-linux.1
powershell -ExecutionPolicy Bypass -File scripts\verify-linux-package.ps1 -Archive release\dlc-deploy-linux-x86_64-v2026.08.24-linux.1.tar.gz
```

在已安装 Bash 的 Linux/macOS 构建机上也可执行：

```bash
./scripts/package-linux.sh v2026.08.24-linux.1
./scripts/verify-linux-package.sh release/dlc-deploy-linux-x86_64-v2026.08.24-linux.1.tar.gz
```

包内包含一键升级、数据库独立升级、备份、健康检查和失败回滚脚本，不包含生产
密钥、数据库、上传媒体或用户文档。服务器升级步骤见包内 `docs/UPGRADE.md`。

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
