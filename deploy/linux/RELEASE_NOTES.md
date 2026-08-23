# 2026.08.23 Linux 发布说明

本版本新增厂商纸质材料智能采集与审核录入能力：

- 管理员和厂商可按资料包连续拍摄名片、产品彩页并进行人工校对；
- 新增数据库持久化识别队列，服务重启后可恢复未完成任务；
- 新增低置信字段确认、重复厂商/产品匹配、产品候选裁图和审核提交；
- 原始材料和候选裁图按私有媒体保存，不会通过公开媒体接口访问；
- 新增七天有效、单次使用的厂商接管邀请链接；
- 腾讯云高精度 OCR 和名片识别继续使用 SecretId/SecretKey；
- 混元视觉已迁移到 TokenHub OpenAI 兼容接口，使用独立 TokenHub API Key；
- 管理后台“平台配置 → 智能采集云服务”支持加密保存密钥且不会回显明文。

数据库架构升级到版本 19，新增：

- `capture_packages`
- `capture_documents`
- `capture_results`
- `capture_product_crops`
- `vendor_invitations`

同时扩展媒体用途和厂商、产品审核记录的资料包来源字段。升级脚本执行前会自动备份
数据库、上传媒体和旧程序。请勿跳过 `sha256sum -c SHA256SUMS` 和健康检查。

完整升级：

```bash
sudo ./scripts/upgrade.sh
```

仅执行数据库升级：

```bash
sudo ./scripts/database-upgrade.sh
```

详细步骤和回滚方式请阅读 `docs/UPGRADE.md`。
