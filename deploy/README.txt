大陆农机配件平台 Windows 部署包

请先阅读 DEPLOY.md。
快速开始：
1. 复制 .env.example 为 .env，并修改数据库密码、管理员密码和 AUTH_SECRET。
2. 执行 scripts\init-db.sql（可选；程序首次启动也会自动建库和迁移）。
3. 执行：powershell -ExecutionPolicy Bypass -File .\scripts\start-windows.ps1
4. 执行：powershell -ExecutionPolicy Bypass -File .\scripts\health-check.ps1
