# 云网关/CDN → Gin 生产部署基线

本项目生产环境不部署 Nginx。云网关/CDN 是唯一公网入口，Gin 是唯一应用源站，同时提供动态 HTML、公开与 CMS API、Sitemap、`robots.txt`、媒体和构建后的前端资源。

## 请求转发

- `/`、所有前台详情路径、`/admin/*`、`/api/*`、`/robots.txt`、`/sitemap.xml`、`/sitemaps/*`、`/uploads/*`、`/images/*` 和 `/assets/*` 全部回源到 Gin。
- 禁止使用“找不到文件则返回固定 `index.html`”的 CDN 回退。公开 HTML 必须由 Gin 判断 200、301、404 并注入 SEO。
- 网关传递原始 `Host`、`X-Forwarded-Proto: https`、`X-Forwarded-For`。`TRUSTED_PROXY_CIDRS` 只填写实际回源出口网段。
- 网关完成证书续期、HTTP→HTTPS、HTTP/2 或 HTTP/3、压缩、WAF 和基础限流。源站 8080 端口仅允许网关访问。

## 缓存策略

| 路径 | 建议策略 |
|---|---|
| `/assets/*` | 1 年，`immutable`；文件名包含内容哈希 |
| `/images/*`、`/uploads/*`、`/api/media/*` | 1 天，可 `stale-while-revalidate` |
| `/api/home`、`/api/site-meta`、`/api/layout-config`、`/api/menus` | 尊重源站 ETag，短缓存 |
| HTML、`robots.txt`、Sitemap、CMS 与其他 API | 不做长期缓存；尊重源站状态码和缓存头 |

不得缓存带 `Set-Cookie` 的响应，不得在不同用户间共享 `/api/admin/*` 与 `/api/auth/*`。

## 发布顺序

1. 备份 MySQL 与 `media_storage`。
2. 构建前端，并令 `PUBLIC_DIR` 指向构建产物。
3. 在维护窗口显式执行版本化迁移：`go run ./cmd/initdb`（发布包使用对应 initdb 可执行程序）。
4. 在 CMS 的 `site.meta` 配置正式 `https://` 域名；生产启动会拒绝空值、示例域名和非 HTTPS 地址。
5. 以 `RUN_MIGRATIONS=false` 启动 Gin，检查 `/api/health`。
6. 从公网入口验证：首页 HTML 含页面标题和 `data-server-rendered="true"` 正文；无效 slug 为 404；旧数字地址为 301。
7. 验证 `/sitemap.xml` 是 Sitemap Index 后，再提交搜索引擎。

## 公网验收示例

```bash
curl -I https://www.example.cn/vendors/1
curl -I https://www.example.cn/vendors/not-exist
curl -s https://www.example.cn/ | grep data-server-rendered
curl -s https://www.example.cn/sitemap.xml
```

若公网首页始终返回同一份静态 `index.html`，说明云网关/CDN 绕过了 Gin，动态 SEO、真实 404 和 301 均不会生效，应立即修正回源规则。
