# 大陆农机配件平台部署说明（Windows）

## 1. 部署包内容

- `server.exe`：后端服务，同时托管前端静态页面和 `/api` 接口。
- `public/`：已构建的前端页面及行业背景图片。
- `media_storage/`：用户上传的 Logo、厂商图集和产品图片存储目录。
- `.env.example`：生产环境配置模板。
- `scripts/start-windows.ps1`、`stop-windows.ps1`：后台启动、停止服务。
- `scripts/health-check.ps1`：健康检查。
- `scripts/backup-data.ps1`：数据库与媒体文件备份。
- `scripts/init-db.sql`：可选的 MySQL 初始化脚本。

## 2. 环境要求

- Windows Server 2019 或更高版本。
- MySQL 5.7+ 或兼容版本。
- 默认数据库端口为 `13306`；如使用 `3306`，请在 `.env` 修改 `DB_PORT`。
- 服务端口默认 `8080`，需在服务器防火墙放行。

## 3. 首次部署

1. 解压 `dlc-deploy-windows-*.zip` 到，例如 `D:\apps\dlc`。
2. 将 `.env.example` 复制为 `.env`。
3. 修改 `.env` 中至少以下内容：`DB_PASSWORD`、`ADMIN_PASSWORD`、`AUTH_SECRET`、`CORS_ALLOWED_ORIGINS`。
4. 如数据库账号尚未创建，可执行 `scripts\init-db.sql`；或使用已有业务账号填写 `.env`。
5. 启动服务：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\start-windows.ps1
```

6. 检查服务：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\health-check.ps1
```

7. 访问首页：`http://服务器IP:8080/`；管理端：`http://服务器IP:8080/admin/login`。

程序首次启动会自动执行数据库迁移和默认数据初始化。生产环境请勿使用模板中的占位密码。

## 4. 数据迁移与升级

部署包默认不打入现网用户媒体。升级前先在旧服务器运行：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\backup-data.ps1
```

将备份目录中的 `database.sql` 导入新服务器数据库，并将 `media_storage` 复制到新部署目录同名位置。若明确需要把当前工作区媒体一并打入包，可在构建时使用：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\package-windows.ps1 -IncludeMedia
```

升级时先运行 `stop-windows.ps1`，替换 `server.exe` 和 `public/`，保留 `.env`、`media_storage/`、`backups/`，再启动服务。

## 5. 运维命令

停止：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\stop-windows.ps1
```

前台方式启动（便于排查日志）：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\start-windows.ps1 -Foreground
```

后台日志：`logs\server.out.log`、`logs\server.err.log`。

## 6. 常见检查

- 页面无法访问：确认防火墙已放行 8080，随后执行健康检查。
- 启动失败：检查 `.env` 中数据库地址、账号、密码，以及 `logs\server.err.log`。
- 图片未显示：确认媒体迁移到了 `media_storage/`，并且 `.env` 的 `MEDIA_DIR=./media_storage`。
- 管理员无法登录：确认不是模板默认密码，并检查数据库连接是否正常。
