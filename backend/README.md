# Backend

Gin + GORM + MySQL API 服务。

```powershell
go run ./cmd/server
```

默认数据库：`root:root@127.0.0.1:13306/dl_nongji_parts`。

主要接口：

- `GET /api/health`
- `GET /api/home`
- `GET /api/vendors`
- `GET /api/products`
- `POST /api/auth/login`（仅厂商账号）
- `POST /api/auth/register`（仅厂商入驻）
- `POST /api/admin/login`（仅管理员、编辑、审核员）
- `GET/POST/PUT/DELETE /api/admin/*`
