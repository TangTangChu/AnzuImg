## 认证

系统支持两类认证凭证，分别是 Web 会话和 API 令牌。请求到达后会按固定顺序读取凭证，先读取 Cookie 中的 `anzuimg_session`，再读取 `Authorization: Bearer <token>`，最后读取 `X-Session-Token`。

## 基础路径

公开媒体资源通过 `/i` 提供，管理接口统一位于 `/api/v1`。

如果你通过 Nginx 做了 API 前缀分流，例如外部使用 `/kotori/` 转发到后端根路径，那么对外接口应写成 `/kotori/api/v1/...`。媒体直链通常仍保持 `/i/...`，这样更方便外部引用。

---

## 错误响应

后端错误响应统一使用同一结构。

```json
{
  "code": "bad_request",
  "message": "invalid request",
  "request_id": "2f6e1b2c9a3d4e5f..."
}
```

`code` 用于稳定分支判断，`message` 用于用户可读提示，`request_id` 用于排查链路问题。响应头中也会返回 `X-Request-ID`。

---

## 1. 资源访问接口

这一组接口用于访问已经上传的媒体文件。媒体既包含图片，也包含视频。

### 获取原始媒体

`GET /i/:hash`

该接口按哈希返回原始文件内容。

### 获取缩略图

`GET /i/:hash/thumbnail`

该接口返回媒体缩略图。对于图片，返回图片缩略图。对于视频，返回上传后生成的视频封面图。

### 通过路由别名访问媒体

`GET /i/r/:route`

该接口通过预设路由别名访问对应媒体。

---

## 2. 管理接口

除初始化、登录等认证入口外，其余管理接口都需要身份验证。

### 2.1 系统健康

#### 健康检查

`GET /health`

该接口用于基础健康检查，成功时返回 `200 OK` 且无响应体。

#### 认证链路检查

`GET /api/v1/ping`

该接口用于校验认证链路是否可用，成功时返回 `200 OK` 且无响应体。

### 2.2 认证管理

认证管理接口基路径为 `/api/v1/auth`。需要特别说明的是，`/api/v1/auth` 下需要登录态的管理操作仅接受 Session，不接受 API Token。

#### 检查初始化状态

`GET /api/v1/auth/status`

该接口用于判断系统是否已完成初始化。

```json
{
  "initialized": true
}
```

#### 系统初始化

`POST /api/v1/auth/setup`

该接口用于首次设置管理员密码，仅在未初始化时可调用。如果服务端配置了 `ANZUIMG_SETUP_TOKEN`，请求体应携带 `setup_token`，并且兼容 `X-Setup-Token` 请求头。

```json
{
  "password": "your_password",
  "setup_token": "optional_setup_token"
}
```

成功后返回：

```json
{
  "message": "system initialized successfully"
}
```

#### 密码登录

`POST /api/v1/auth/login`

```json
{
  "password": "your_password"
}
```

成功后返回：

```json
{
  "token": "session_token_string",
  "expires_at": "2024-01-01T00:00:00Z",
  "auth_method": "password"
}
```

#### 验证会话

`GET /api/v1/auth/validate`

该接口用于验证当前凭证是否有效，并返回认证类型与会话时间信息。

```json
{
  "valid": true,
  "auth_method": "session",
  "expires_at": "...",
  "created_at": "...",
  "last_used": "..."
}
```

#### 修改密码

`POST /api/v1/auth/change-password`

```json
{
  "current_password": "old_password",
  "new_password": "new_password"
}
```

成功后返回：

```json
{
  "message": "password changed successfully"
}
```

#### Passkey 接口

##### 登录开始

`GET /api/v1/auth/passkey/login/begin`

该接口用于发起 Passkey 登录挑战，前端拿到挑战参数后应调用浏览器 WebAuthn 能力继续流程。

##### 登录完成

`POST /api/v1/auth/passkey/login/finish`

该接口用于提交浏览器返回的签名结果，服务端验证通过后会建立登录态。

##### 注册开始

`GET /api/v1/auth/passkey/register/begin`

该接口用于获取 Passkey 注册参数，通常在已登录状态下调用。

##### 注册完成

`POST /api/v1/auth/passkey/register/finish`

该接口用于提交注册凭证并完成设备绑定。

##### 列出 Passkey

`GET /api/v1/auth/passkeys`

该接口用于获取当前账号已注册的 Passkey 设备列表。

##### 删除 Passkey

`DELETE /api/v1/auth/passkeys/:credential_id`

该接口用于删除指定 Passkey。

兼容删除接口：

`POST /api/v1/auth/passkeys/:credential_id/delete`

#### API Token 管理

##### 创建 Token

`POST /api/v1/auth/tokens`

该接口用于创建 API Token。请求体包含 `name`、可选 `token_type` 和可选 `ip_allowlist`。`token_type` 支持 `full`、`upload`、`list`，默认值为 `full`。

```json
{
  "name": "Token Description",
  "token_type": "full",
  "ip_allowlist": ["192.168.1.1/32", "10.0.0.0/8"]
}
```

成功后返回一次性可见的原始令牌：

```json
{
  "token": "raw_token_string",
  "raw_token": "raw_token_string"
}
```

##### 获取 Token 列表

`GET /api/v1/auth/tokens`

该接口用于获取当前账号下的 Token 列表。

##### 删除 Token

`DELETE /api/v1/auth/tokens/:id`

该接口用于删除指定 Token。

兼容删除接口：

`POST /api/v1/auth/tokens/:id/delete`

#### Token 日志

`GET /api/v1/auth/tokens/logs`

该接口支持分页查询，常用参数为 `page` 和 `page_size`。

```json
{
  "data": [APITokenLog Object],
  "total": 100,
  "page": 1,
  "size": 50
}
```

`DELETE /api/v1/auth/tokens/logs`

该接口用于清理 Token 日志。请求需要提供 `days`，表示清理多少天之前的记录。

兼容清理接口：

`POST /api/v1/auth/tokens/logs/cleanup`

```json
{
  "deleted": 120,
  "cutoff": "2026-02-03T00:00:00Z"
}
```

#### 安全日志

`GET /api/v1/auth/security/logs`

该接口用于查看近期安全事件和关键操作，支持按分页参数查询，并支持通过 `failed_only` 过滤失败登录。

```json
{
  "data": [
    {
      "id": 1,
      "category": "auth",
      "level": "warning",
      "action": "login_failed",
      "message": "failed login attempt",
      "method": "POST",
      "path": "/api/v1/auth/login",
      "ip_address": "127.0.0.1",
      "username": "admin",
      "created_at": "2026-02-20T12:34:56Z"
    }
  ],
  "total": 42,
  "page": 1,
  "size": 20
}
```

### 2.3 媒体管理

媒体管理接口基路径为 `/api/v1/images`。虽然路径保留了历史命名，但实际对象已经是媒体，包含图片与视频。

#### 上传媒体

`POST /api/v1/images`

该接口支持多文件上传，并支持全局元数据和按文件元数据两种写法。请求使用 `multipart/form-data`，核心字段是 `file`。你可以设置 `route`、`description`、`tags` 和 `custom_name` 作为全局默认值，也可以通过 `metadata` 为每个文件单独指定这些值。

转换参数仅对图片生效。`convert=true` 时可配合 `target_format`、`quality` 和 `effort` 进行格式转换。视频不会执行图片转换流程。

`metadata` 结构如下：

```json
[
  {
    "description": "desc1",
    "tags": ["tag1"],
    "routes": ["route1"],
    "custom_name": "file1.png",
    "client_index": 0
  }
]
```

响应是逐文件结果数组。成功项会返回哈希、尺寸、媒体类型信息与访问链接，失败项会返回稳定错误码和错误信息。

```json
[
  {
    "client_index": 0,
    "success": true,
    "hash": "...",
    "file_name": "...",
    "url": "http://...",
    "path": "...",
    "mime": "video/mp4",
    "width": 1920,
    "height": 1080,
    "duration_seconds": 37
  },
  {
    "client_index": 1,
    "success": false,
    "file_name": "bad.txt",
    "code": "unsupported_file_type",
    "message": "unsupported file type: text/plain"
  }
]
```

#### 获取媒体列表

`GET /api/v1/images`

该接口支持分页、标签筛选和文件名模糊查询，常用参数为 `page`、`page_size`、`tag` 和 `file_name`。

```json
{
  "data": [Image Object],
  "total": 100,
  "page": 1,
  "size": 20
}
```

#### 获取媒体详情

`GET /api/v1/images/:hash/info`

详情包含通用文件信息、可选的图像尺寸与视频时长、描述标签、上传来源以及路由别名。

```json
{
  "hash": "...",
  "file_name": "...",
  "mime_type": "video/mp4",
  "size": 1024,
  "width": 800,
  "height": 600,
  "duration_seconds": 37,
  "description": "...",
  "tags": ["tag1", "tag2"],
  "uploaded_by_token_id": 12,
  "uploaded_by_token_name": "Upload Token",
  "uploaded_by_token_type": "upload",
  "routes": ["route1", "route2"],
  "created_at": "...",
  "updated_at": "..."
}
```

#### 更新媒体信息

`PATCH /api/v1/images/:hash`

该接口可更新描述、标签、文件名和路由。传入 `routes` 时会覆盖旧路由集合。

```json
{
  "description": "New Description",
  "tags": ["new", "tags"],
  "file_name": "new_name.png",
  "routes": ["route1", "route2"]
}
```

#### 删除媒体

`DELETE /api/v1/images/:hash`

该接口会删除原文件、关联缩略图和数据库记录。

兼容删除接口：

`POST /api/v1/images/:hash/delete`

### 2.4 标签接口

`GET /api/v1/tags`

该接口返回标签与使用次数，支持 `limit` 参数。默认值为 200，最大值为 1000。

```json
{
  "data": [
    { "tag": "cat", "count": 12 },
    { "tag": "sunset", "count": 5 }
  ]
}
```

### 2.5 路由管理

路由管理接口基路径为 `/api/v1/routes`。

#### 获取路由列表

`GET /api/v1/routes`

该接口用于分页查询系统中已注册的路由别名。

#### 删除路由

`DELETE /api/v1/routes/:route`

该接口用于删除路由别名，不会删除对应媒体文件。

兼容删除接口：

`POST /api/v1/routes/:route/delete`
