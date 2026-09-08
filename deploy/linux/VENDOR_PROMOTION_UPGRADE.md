# 厂商推广工作台升级（Schema 19 → 20）

## 本次交付

- 厂商登录后进入工作台：待办、审核通知、企业资料完善建议、推广摘要。
- 本厂产品三步录入、图片上传进度/失败重试、私有自动保存草稿、恢复和复制新产品。既有六项预设与历史自定义值保留。
- 企业展厅优先展示本厂产品，支持型号搜索、分类、置顶排序；本厂产品地址为 `/v/:slug/products/:supplierId`，公共产品目录继续支持多供应商。
- 公开页面分享链接、二维码与本地生成海报；继续使用公开水印图片，不输出私有采集原图或隐藏联系方式。
- 7/30/90 天推广统计；厂商/供货关系由服务器核验归属。平台公共产品流量单列；历史数据不补造归属。联系操作不等于接通、添加微信或成交。
- 本厂产品 SEO、站点地图和静态生成；审核、供货信息变动会失效关联静态页。

## 数据升级边界

新表 `vendor_work_drafts` 保存账号私有产品/企业草稿及版本、审核提交标识。

`product_suppliers` 仅新增 `showroom_featured` 与 `showroom_order`；
`analytics_events` 仅新增 `vendor_id`、`supplier_id`、`source` 及归属索引。

19 → 20 独立迁移不运行旧数据清理、价格回填或账号清理；原企业、产品、媒体、审核记录均不批量覆盖。重复执行不会重复建表或产生第二条版本记录。旧静态构建标记为失效，动态页面仍然可用。

升级脚本调用新版本 `initdb --migrate-only`，不额外执行默认资料初始化或旧 SEO 回填。不要用旧包内的 initdb 替换新二进制。

## Linux 升级执行

适用于已有标准安装：`/opt/dalu-parts`，服务 `dalu-parts.service`，配置 `/etc/dalu-parts/dlc.env`。先在预发布环境演练，确认当前数据库版本为 19。更早版本应先按原升级说明升至 19。

本次发布包为 `dlc-deploy-linux-x86_64-v2026.09.08-linux.1.tar.gz`，适用于 Linux x86_64。将包上传并解压；以下命令在**新包解压目录**执行。生成发布包不代表已经部署到服务器：

```bash
# 先确认当前目录是新包根目录
test -f public/index.html && test -f server && test -f initdb
chmod +x server initdb scripts/*.sh

# 推荐：备份数据库/媒体和程序，停机，安装新程序，增量迁移并检查服务
sudo bash scripts/upgrade.sh

sudo systemctl status dalu-parts.service --no-pager
sudo bash scripts/health-check.sh
sudo journalctl -u dalu-parts.service -n 80 --no-pager
```

仅在需要单独执行数据库升级时使用以下命令；不要重复替代完整程序升级：

```bash
sudo bash scripts/database-upgrade.sh
```

该脚本先备份再停服务，执行迁移后恢复原运行状态；失败时保持停机并输出备份位置。若只升级了数据库，随后仍须安装新前后端。

使用自己的数据库管理工具确认版本（不要把密码放进命令历史）：

```sql
SELECT version, name, applied_at FROM schema_migrations ORDER BY version DESC LIMIT 1;
-- 预期：20 / vendor-showroom-workspace-v1
SHOW COLUMNS FROM product_suppliers LIKE 'showroom_%';
SHOW TABLES LIKE 'vendor_work_drafts';
```

升级后在后台“平台配置 → 静态化与缓存”执行全量静态生成。存在外部 CDN/反向代理 HTML 缓存时，也需清理旧 HTML，不能仅刷新浏览器。

## 回滚与保护

脚本会报告程序快照和数据库/媒体备份位置。迁移失败或上线异常，先停服务，按备份恢复流程恢复程序；涉及数据库恢复时先保护升级后产生的新草稿、审核记录和统计事件，再恢复同一时间点备份。不要直接删新表、新列或覆盖生产数据库。

原有腾讯云配置和密钥不变，本次不增加短信、支付或自动发布。

## 验证与试用

自动化覆盖：草稿串行保存/失败重试、历史值0、幂等提交、跨厂商权限、审核前保持原公开版本、审核通知、私有图片限制、统计归属/自身预览排除、私有草稿图片保留、静态失效、19→20 重复执行。

浏览器验收使用明确标注的测试资料及 Mock API，覆盖 360/390/430/1440px；不代表真实生产数据或真实厂商试用。截图和结果输出在源码的 `outputs/vendor-promotion/`。

运行开发验证：

```bash
cd frontend
npm ci
npm test
npm run build
npm run budget
cd ../backend
go test ./...
# 可选真实 MySQL 验证：仅创建并清理 dlc_promotion_test_<时间戳> 隔离库
RUN_PROMOTION_MYSQL_TESTS=1 go test ./internal/api -run TestVendorPromotionMySQL -count=1 -v
```

浏览器脚本：前端服务启动后，环境中提供 Playwright，运行 `node frontend/scripts/verify-vendor-promotion.cjs`。可设置 `PROMOTION_BASE_URL` 和 `PROMOTION_BROWSER`；默认使用本机 Edge。

上线前仍需 10 家厂商真实试用：记录3–5分钟提交、30秒更新供货的达成率、草稿恢复成功率、审核等待时间、手机实机弱网/上传中断与分享后的联系操作。当前不承诺这些业务指标已达成。

依赖检查发现现有 Vitest 2 开发测试链的安全告警；本次未做跨主版本自动升级。不要把 Vite/Vitest 开发或测试 UI 暴露到公网；正式环境只部署生产构建及 Go 服务。
