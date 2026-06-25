# 大陆农机配件 MVP 设计

## 目标

实现一个可运行的“大陆农机配件”网站 MVP：前台首页尽量还原已确认效果图，数据全部来自后端 API；后台 CMS 支持基础内容录入和维护；后端使用 Gin、GORM、MySQL 5.7。

## 范围

第一版包含：

- PC 首页：顶部导航、左侧可滚动菜单、Banner 搜索、推荐厂商、更多厂商、入驻引导、底部统计。
- 移动端首页：顶部栏、搜索框、分类快捷入口、推荐厂商、更多厂商、固定底部导航。
- 后台 CMS：管理员登录、菜单、厂商、标签、分类、产品、Banner、平台配置的基础增删改查。
- 数据能力：MySQL 表结构自动迁移、初始化种子数据、前台所有展示数据通过接口读取。
- 跳转能力：厂商配置官网时“进入厂商”新窗口打开官网；没有官网时跳转站内详情。

第一版不包含：

- 多管理员角色与细粒度权限。
- 审计日志、操作审批、复杂工作流。
- 本地文件上传与图片裁剪。CMS 先录入图片 URL，前端使用 URL 或内置占位图展示。
- 真实采购/求购信息流。首页保留顶部“采购信息”入口，但主体不展示“最新求购”。

## 架构

项目采用同仓库双目录：

```text
backend/
frontend/
docs/
```

后端是一个 Gin REST API 服务，负责数据库连接、迁移、种子数据、公共 API 和后台 API。前端是 Vite React 应用，使用 React Router 管理前台和后台路由，使用 Axios 调用后端接口。

本地默认数据库连接：

```text
host: 127.0.0.1
port: 13306
user: root
password: root
database: dl_nongji_parts
```

后端启动时会读取环境变量覆盖默认配置，并在数据库存在时执行 AutoMigrate。初始化数据通过启动种子逻辑保证首页和 CMS 首次运行可用。

## 后端设计

### 模块

- `config`：读取端口、数据库 DSN、管理员账号等配置。
- `database`：建立 GORM 连接、迁移表结构、写入种子数据。
- `model`：定义 Menu、Vendor、Tag、VendorTag、Category、Product、Banner、SiteConfig、AdminUser。
- `handler`：实现公共 API 和后台 API。
- `middleware`：实现简单 token 鉴权和 CORS。
- `router`：注册路由。

### 数据模型

按需求文档为基础建表，并为 MVP 增加两个必要模型：

- `banners`：保存首页 Banner 标题、副标题、背景图、搜索提示、热门搜索词、启用状态和排序。
- `admin_users`：保存后台管理员账号、密码哈希和启用状态。

厂商标签使用 `vendor_tags` 关联表。产品关联分类和厂商。平台统计、入驻文案、保障文案使用 `site_configs` 存 key-value。

### 公共 API

- `GET /api/health`：健康检查。
- `GET /api/home`：聚合首页数据，包括顶部菜单、侧边菜单、辅助链接、移动端快捷入口、Banner、推荐厂商、更多厂商、统计和入驻配置。
- `GET /api/menus?type=sidebar`：按类型返回菜单树。
- `GET /api/vendors`：厂商列表，支持分页、地区、分类/关键词、服务标签和排序。
- `GET /api/vendors/recommended`：推荐厂商。
- `GET /api/vendors/:id`：厂商详情。
- `GET /api/products`：产品列表。
- `GET /api/search?keyword=xxx`：搜索厂商、产品、分类。

### 后台 API

- `POST /api/admin/login`：登录，返回 bearer token。
- `GET /api/admin/profile`：验证当前登录状态。
- `GET/POST/PUT/DELETE /api/admin/menus`
- `GET/POST/PUT/DELETE /api/admin/vendors`
- `GET/POST/PUT/DELETE /api/admin/tags`
- `GET/POST/PUT/DELETE /api/admin/categories`
- `GET/POST/PUT/DELETE /api/admin/products`
- `GET/POST/PUT/DELETE /api/admin/banners`
- `GET /api/admin/configs`
- `PUT /api/admin/configs/:key`

MVP 使用单管理员登录。默认账号可由环境变量配置，开发默认值为 `admin / admin123`。密码存储为 bcrypt 哈希。

## 前端设计

### 路由

前台：

- `/`
- `/vendors`
- `/vendors/:id`
- `/products`
- `/products/:id`
- `/search`
- `/join`
- `/about`

后台：

- `/admin/login`
- `/admin/dashboard`
- `/admin/menus`
- `/admin/vendors`
- `/admin/tags`
- `/admin/categories`
- `/admin/products`
- `/admin/banners`
- `/admin/configs`

### 前台组件

- `PublicHeader`：PC 顶部导航，移动端隐藏。
- `SidebarNav`：PC 左侧主导航和辅助链接，支持二级菜单展开、默认展开、高亮和滚动。
- `HeroSearch`：Banner、搜索框和热门搜索词。
- `VendorCard`：PC 网格厂商卡片和移动列表厂商卡片共用数据。
- `RecommendedVendors`：推荐厂商区，PC 一行 5 个。
- `MoreVendors`：更多厂商区，PC 至少两行，移动端信息流。
- `JoinBanner`：入驻引导条。
- `StatsFooter`：保障文案和统计数字。
- `MobileHeader`、`MobileCategoryGrid`、`MobileBottomNav`：移动端专用结构。

首页根据 CSS media query 在 PC 和移动端切换布局。PC 布局以 64px 顶部栏和 250px 左侧栏为基准；移动端隐藏 PC 导航并启用底部固定导航。

### CMS 组件

- `AdminLoginPage`：登录表单。
- `AdminLayout`：后台侧边栏、顶部栏、内容区。
- `AdminTablePage`：可复用列表页骨架，支持搜索、启用状态、编辑和删除。
- `MenuForm`、`VendorForm`、`TagForm`、`CategoryForm`、`ProductForm`、`BannerForm`、`ConfigForm`：各实体录入表单。

CMS 表单以文本字段、数字字段、选择框、开关为主。图片字段先使用 URL 输入框。

## 数据流

前台首页加载时调用 `/api/home`。后端按排序读取启用数据，构造菜单树、推荐厂商、更多厂商、Banner 和配置项后返回。前端只负责展示和交互，不内置业务数据。

CMS 登录后将 token 保存在 `localStorage`。后台请求通过 Axios 拦截器附带 `Authorization: Bearer <token>`。401 时跳回登录页。

厂商卡片点击规则：

- 有 `websiteUrl`：`window.open(websiteUrl, "_blank", "noopener,noreferrer")`。
- 无 `websiteUrl`：跳转 `/vendors/:id`。
- 查看详情始终跳转 `/vendors/:id`。

搜索规则：

- 首页搜索框有关键词时跳转 `/search?keyword=xxx`。
- 搜索页调用 `/api/search` 展示结果。
- 空关键词不跳转，保留在当前页。

## 视觉实现

使用需求文档给定色值作为 CSS 变量。整体控制为蓝白、清爽、工业信息平台风格。

PC 首页关键约束：

- 顶部栏约 64px。
- 左侧栏约 250px，主导航区域可滚动，辅助链接在下方独立分区。
- Banner 高度约 230px，左侧蓝色渐变，右侧使用农机/麦田视觉图或渐变占位。
- 推荐厂商一行 5 张卡片。
- 更多厂商至少 10 个展示位。
- 首页主体不出现“最新求购”等模块。

移动端关键约束：

- 顶部为菜单图标、Logo、搜索图标。
- 搜索框和分类宫格靠近首屏。
- 厂商卡片为横向图片 + 文本的纵向列表。
- 底部导航固定，内容区保留底部 padding。

## 错误处理

后端统一返回 JSON：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

错误时 `code` 非 0，`message` 给出可显示文案。后台表单保存失败时在页面顶部显示错误提示。公共首页接口失败时展示空状态和重试按钮。

## 测试与验收

后端测试：

- 菜单树构建：父子级、排序、默认展开字段正确。
- 厂商官网跳转数据：有无官网 URL 的响应字段正确。
- `/api/home` 聚合结构包含首页所需关键字段。
- 后台鉴权：未登录访问后台 API 返回 401。

前端测试：

- API 类型和数据适配函数测试。
- 首页搜索有关键词时生成正确 `/search?keyword=...`。
- 厂商进入按钮根据 `websiteUrl` 选择外链或站内链接。

人工验收：

- PC 首页接近效果图，推荐厂商一行 5 个，更多厂商至少 2 行。
- 左侧“传动配件”默认展开并显示 7 个二级菜单。
- 移动端布局接近手机效果图，底部导航固定且不遮挡内容。
- CMS 能登录并维护菜单、厂商、产品、分类、标签、Banner 和统计配置。
- 修改 CMS 数据后，前台刷新能看到变化。
