# 大陆农机配件平台生产部署

生产架构为“云网关/CDN → Gin → MySQL”，不部署 Nginx。详细回源、缓存和公网 SEO 验收规则见 [GATEWAY_GIN.md](GATEWAY_GIN.md)。

## 部署包

- `server.exe`：Gin 应用，统一提供 HTML、API、Sitemap 和媒体访问。
- `public/`：Vite 构建产物，必须由 `PUBLIC_DIR` 指向。
- `media_storage/`：上传媒体，不放入公开静态目录。
- `.env.example`：生产配置模板。
- `scripts/`：启动、停止、健康检查和备份脚本。

## 首次部署

1. 解压部署包并将 `.env.example` 复制为 `.env`。
2. 替换数据库密码、管理员密码、`AUTH_SECRET`、正式域名和网关出口 CIDR。
3. 先创建数据库账号和数据库，再备份并显式执行版本化迁移：`.\initdb.exe`。
4. 在 CMS 的站点配置中填写正式 HTTPS `siteUrl`。生产进程会拒绝空值、示例域名或 HTTP 地址。
5. 设置 `RUN_MIGRATIONS=false` 后启动服务。
6. 云网关/CDN 将所有站点请求转发至 Gin，源站 8080 只允许网关访问。

Windows 启动与检查：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\start-windows.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\health-check.ps1
```

## 升级

升级前执行：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\backup-data.ps1
```

随后停止旧进程，替换 `server.exe` 与 `public/`，保留 `.env`、`media_storage/` 和 `backups/`。在维护窗口使用与新版本一致的迁移程序升级数据库，再以 `RUN_MIGRATIONS=false` 启动应用。迁移失败时不得启动新版本，也不得在未恢复备份前继续写入。

## 上线验收

- `/api/health` 返回 200。
- 公网首页响应正文含动态标题和 `data-server-rendered="true"`。
- 禁用 JavaScript 后仍能读取正文、面包屑和内部链接。
- 旧数字地址返回 301；不存在的 slug 返回 404。
- `/admin/*` 为 `noindex,nofollow`；搜索页为 `noindex,follow`。
- `/sitemap.xml` 为 Sitemap Index，未来定时内容不在分片中。
- 登录响应设置 `HttpOnly; Secure; SameSite=Lax` 会话 Cookie，前端不保存令牌。

完成这些检查后再向百度等搜索引擎提交 Sitemap。
