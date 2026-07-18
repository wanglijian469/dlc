大陆农机配件平台生产部署包

1. 阅读 DEPLOY.md 与 GATEWAY_GIN.md。
2. 复制 .env.example 为 .env，替换所有 CHANGE_ME 值和正式域名。
3. 备份数据库与媒体，并显式执行版本化迁移。
4. 设置 RUN_MIGRATIONS=false，启动 Gin。
5. 云网关/CDN 将全部请求回源 Gin；不部署 Nginx，不配置静态 index.html 回退。
