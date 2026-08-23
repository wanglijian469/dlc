# 已部署系统升级与回滚手册

本手册用于升级已按标准目录部署的系统。适用目录为：

- 程序与前台文件：`/opt/dalu-parts`
- 配置文件：`/etc/dalu-parts/dlc.env`
- 上传媒体文件：`/var/lib/dalu-parts/media_storage`

升级会保留现有配置和已上传文件。**不要**用安装包中的示例配置覆盖
`/etc/dalu-parts/dlc.env`。

## 一、升级前准备

1. 使用具备 `sudo` 权限的管理员登录 Linux 服务器。
2. 将新版本 `tar.gz` 部署包上传到服务器，例如上传到 `/tmp`。
3. 确认旧系统当前可访问，并预留足够磁盘空间保存数据库、媒体文件和旧程序快照。
4. 如有自行修改过 Nginx 配置，请先备份：

```bash
sudo cp /etc/nginx/conf.d/dalu-parts.conf \
  /etc/nginx/conf.d/dalu-parts.conf.before-upgrade
```

## 二、一键升级（推荐）

在 Linux 服务器执行以下命令。请把压缩包文件名替换为实际版本号：

```bash
cd /tmp
tar -xzf dlc-deploy-linux-x86_64-v2026.08.23-linux.1.tar.gz
cd dlc-deploy-linux-x86_64-v2026.08.23-linux.1
chmod +x server initdb scripts/*.sh
sha256sum -c SHA256SUMS
sudo ./scripts/upgrade.sh
```

其中：

- `sha256sum -c SHA256SUMS` 的每一项都应显示 `OK`；若失败，请停止升级并重新上传部署包。
- `chmod` 用于确保从 Windows 制作或经文件传输后的 Linux 可执行权限正确。
- `upgrade.sh` 仅适用于已有系统；全新服务器请使用《安装手册》。

升级脚本将按以下顺序执行：

1. 备份数据库和上传媒体文件到 `/var/backups/dalu-parts/database-media/`；
2. 快照旧版本的程序、迁移程序、前台文件、systemd 服务文件及环境配置到
   `/var/backups/dalu-parts/releases/`；
3. 停止服务，替换程序和前台文件，执行数据库迁移；
4. 启动服务，并检查 `/api/health` 健康接口。

生产环境仍应保持 `RUN_MIGRATIONS=false`。升级脚本会调用独立的 `initdb`
程序执行迁移，不会依赖服务启动时自动迁移。

脚本会检查旧版标准目录和环境配置是否存在；不满足条件时会停止，不会改动已有安装。
对于静态页目录，只有旧版默认值
`/opt/dalu-parts/public/static-pages` 会自动调整为
`/var/lib/dalu-parts/static-pages`；已设置的自定义目录不会被修改。

## 三、仅升级数据库

若程序文件已提前部署，或需要先单独验证数据库迁移，可执行：

```bash
cd /tmp/dlc-deploy-linux-x86_64-v2026.08.23-linux.1
sudo ./scripts/database-upgrade.sh
```

该脚本会备份数据库和媒体文件、停止正在运行的应用、使用升级包内的 `initdb`
执行迁移，成功后恢复服务并执行健康检查。版本 19 迁移可重复执行；如果数据库已是
当前版本，脚本会正常完成而不会重复创建表。

若迁移失败，应用会保持停止状态，脚本会输出备份目录。确认并恢复数据库后再重新启动，
不要在迁移失败后直接运行新版服务。

## 四、更新 Nginx 配置

本版本增加了按 IP 限流、超限返回 `429`、媒体防盗链等配置。升级程序后，请用
**当前正在使用的域名和证书路径**重新生成 Nginx 配置：

```bash
sudo ./scripts/render-nginx-config.sh \
  www.example.cn \
  /etc/pki/tls/certs/www.example.cn.fullchain.pem \
  /etc/pki/tls/private/www.example.cn.key
sudo nginx -t
sudo systemctl reload nginx
```

请将上述三项示例值替换为正式域名、证书完整链文件路径和私钥路径。`nginx -t`
必须通过后才执行 reload。

若旧系统未使用 Nginx，应用自身的访问行为识别仍会生效；但按请求速率的限流需要
部署并启用 Nginx 后才会生效。

## 五、升级后的检查

1. 访问 `https://你的域名/api/health`，应返回正常健康状态。
2. 使用平台管理员登录后台，检查“平台配置 → 访问与采集防护”和“智能采集云服务”。
   OCR 使用腾讯云 SecretId/SecretKey，视觉服务需要另行创建 TokenHub API Key。
   首次上线后的 7 天内，建议保持访问防护“仅记录，不自动封禁”。
3. 检查 `https://你的域名/robots.txt`，确认 AI 爬虫规则和搜索引擎规则符合预期。
4. 在后台执行历史图片水印生成；对状态为“需重新生成”的厂商页和产品页重新生成静态页面。
5. 通过 HTTPS 分别检查一个厂商页面、一个产品页面、一张图片和管理员登录。
6. 查看服务日志确认没有错误：

```bash
sudo journalctl -u dalu-parts -n 100 --no-pager
```

## 六、失败处理与回滚

如果替换程序、迁移或健康检查失败，`upgrade.sh` 会自动恢复旧版本的程序文件。
但是，**仅恢复程序不能安全撤销已经执行过的数据库变更**。

如数据库迁移已经改变数据，且必须完整回滚，请使用脚本输出的备份路径，按以下步骤操作：

```bash
sudo systemctl stop dalu-parts
# 从脚本提示的备份目录恢复 database.sql.gz 和 media_storage.tar.gz。
# 从 /var/backups/dalu-parts/releases/<时间戳>/ 恢复旧程序文件。
sudo systemctl start dalu-parts
sudo ./scripts/health-check.sh
```

恢复数据库前，请先确认备份文件的时间、数据库名称和恢复命令；建议由熟悉 MySQL 的
运维人员执行。MySQL 服务版本升级应遵循 MySQL 官方升级流程并先完成备份和检查，
不支持直接原地降级。

## 七、全新安装

没有旧系统、标准目录不一致，或希望在新服务器部署时，请阅读包内
`docs/INSTALL.md`（《Linux 安装手册》），不要执行 `upgrade.sh`。
