# 供求信息功能移除

已移除网页供求列表、详情、发布、个人发布管理、后台审核，以及对应 API、DTO、模型、样式和 Flutter 页面。网页账号入口改为账号资料，手机导航移除发布和供求入口。采购竞价、厂商动态、采购商资料及共享图片上传继续保留。

数据库 Schema 23 使用可重复执行的迁移，删除 `market_post_media`、`market_contact_access_logs`、`market_posts`，清除旧功能菜单及其子菜单、purchase 占位页面，并使旧静态页面失效。新建数据库不再创建这些表。升级前备份数据库，然后执行 `go run ./cmd/initdb -migrate-only`。

本地已备份并完成升级，确认三张表不存在、供求菜单为零，重复迁移成功。备份位于 `.tmp/db-backups/before_supply_removal_20260911.sql`，不应提交到代码仓库。

验证：Go 全量测试、前端测试和生产构建、入口体积检查、Flutter analyze 和测试、OpenAPI 契约校验通过。本地 `/purchase`、公开/个人/后台供求 API 返回 404，首页、健康检查和站点地图返回 200，无供求导航。浏览器连接工具本次不可用，未新增浏览器截图。
