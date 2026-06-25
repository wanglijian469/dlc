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
- `POST /api/admin/login`
- `GET/POST/PUT/DELETE /api/admin/*`
