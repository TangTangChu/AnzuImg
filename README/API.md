## 认证

系统支持 Web 会话（Session）和 API 令牌（API Token）两种认证凭证。凭证可通过以下任意方式传递，优先级从上到下：

1. **Cookie**: `anzuimg_session=<token>` (主要用于 Web 端)
2. **HTTP Header**: `Authorization: Bearer <token>` (推荐用于 API)
3. **HTTP Header**: `X-Session-Token: <token>`

## 基础路径

- **图片资源**: `/i`
- **API 接口**: `/api/v1`

---

## 1. 图片资源

用于公开访问已上传的图片资源。

### 获取图片

`GET /i/:hash`

获取指定 Hash 的原图。

- **参数**:
  - `hash` (path): 图片的唯一 Hash 值。

### 获取缩略图

`GET /i/:hash/thumbnail`

获取指定 Hash 图片的缩略图。

- **参数**:
  - `hash` (path): 图片的唯一 Hash 值。

### 通过自定义路由获取图片

`GET /i/r/:route`

通过预设的自定义路由别名访问图片。

- **参数**:
  - `route` (path): 自定义路由别名。

---

## 2. 管理接口

除认证接口外的所有管理接口均需要通过身份验证。

### 2.1 系统健康

#### 健康检查

`GET /health`

- **响应**: `200 OK` (无响应体)

#### Ping (Auth Required)

`GET /api/v1/ping`

- **响应**: `200 OK` (无响应体)

### 2.2 认证管理

Base URL: `/api/v1/auth`

#### 检查初始化状态

`GET /api/v1/auth/status`

检查系统是否已完成初始化（是否存在管理员账号）。

- **响应**:
  ```json
  {
    "initialized": true
  }
  ```

#### 系统初始化

`POST /api/v1/auth/setup`

设置管理员初始密码。仅在 `initialized: false` 时可用。

> [!NOTE]
> 若服务端设置了 `ANZUIMG_SETUP_TOKEN`，则此接口需要携带请求头：`X-Setup-Token: <token>`。

- **请求**:
  ```json
  {
    "password": "your_password" // min 8 chars
  }
  ```
- **响应**:
  ```json

  ```

#### 标签列表

`GET /api/v1/tags`

- **Query 参数**:
  - `limit`: 返回数量上限 (默认 200，最大 1000)。

- **响应**:
  ```json
  {
    "data": [
      { "tag": "cat", "count": 12 },
      { "tag": "sunset", "count": 5 }
    ]
  }
  ```
  {
  "message": "system initialized successfully"
  }
  ```

  ```

#### 密码登录

`POST /api/v1/auth/login`

- **请求**:
  ```json
  {
    "password": "your_password"
  }
  ```
- **响应**:
  ```json
  {
    "token": "session_token_string",
    "expires_at": "2024-01-01T00:00:00Z",
    "auth_method": "password"
  }
  ```

#### 验证会话

`GET /api/v1/auth/validate`

验证当前 Token 是否有效。

- **响应**:
  ```json
  {
    "valid": true,
    "auth_method": "session", // or "api_token"
    "expires_at": "...",
    "created_at": "...",
    "last_used": "..."
  }
  ```

#### 修改密码

`POST /api/v1/auth/change-password`

- **请求**:
  ```json
  {
    "current_password": "old_password",
    "new_password": "new_password"
  }
  ```
- **响应**:
  ```json
  {
    "message": "password changed successfully"
  }
  ```

#### Passkey (WebAuthn) 相关

- `GET /api/v1/auth/passkey/login/begin`: 开始 Passkey 登录流程。
- `POST /api/v1/auth/passkey/login/finish`: 完成 Passkey 登录流程。
- `GET /api/v1/auth/passkey/register/begin`: 开始 Passkey 注册（需登录）。
- `POST /api/v1/auth/passkey/register/finish`: 完成 Passkey 注册（需登录）。
- `GET /api/v1/auth/passkeys`: 列出已注册的 Passkey。
- `DELETE /api/v1/auth/passkeys/:credential_id`: 删除指定 Passkey。

#### API Token 管理

#### 创建 Token

`POST /api/v1/auth/tokens`

- **请求**:
  ```json
  {
    "name": "Token Description",
    "ip_allowlist": ["192.168.1.1/32", "10.0.0.0/8"] // 可选
  }
  ```
- **响应**:
  ```json
  {
    "token": "raw_token_string", // 仅显示一次
    "raw_token": "raw_token_string"
  }
  ```

#### 获取 Token 列表

`GET /api/v1/auth/tokens`

- **响应**: `[APIToken Object]` 列表

#### 删除 Token

`DELETE /api/v1/auth/tokens/:id`

---

### 2.3 图片管理

Base URL: `/api/v1/images`

#### 上传图片

`POST /api/v1/images`

支持多文件上传和元数据设置。

- **Content-Type**: `multipart/form-data`
- **参数**:
  - `file`: 文件数据 (支持多个)。
  - `route`: 全局路由别名 (逗号分隔，可选)。
  - `description`: 全局描述 (可选)。
  - `tags`: 全局标签 (逗号分隔，可选)。
  - `custom_name`: 全局自定义文件名 (可选)。
  - `convert`: `true/false` (是否转换格式)。
  - `target_format`: `jpeg/png/webp` (转换目标格式)。
  - `quality`: `1-100` (压缩质量)。
  - `effort`: `0-6` (编码努力程度)。
  - `metadata`: JSON 字符串，用于为每个文件单独指定元数据。
    ```json
    [
      {
        "description": "desc1",
        "tags": ["tag1"],
        "routes": ["route1"],
        "custom_name": "file1.png"
      }
    ]
    ```

- **响应**: 上传结果列表。
  ```json
  [
    {
      "success": true,
      "hash": "...",
      "file_name": "...",
      "url": "http://...",
      "path": "...",
      "width": 100,
      "height": 100,
      ...
    }
  ]
  ```

#### 图片列表

`GET /api/v1/images`

- **Query 参数**:
  - `page`: 页码 (默认 1)。
  - `page_size`: 每页数量 (默认 20)。
  - `tag`: 按标签筛选。
  - `file_name`: 按文件名模糊搜索。

- **响应**:
  ```json
  {
    "data": [Image Object],
    "total": 100,
    "page": 1,
    "size": 20
  }
  ```

#### 获取图片详情

`GET /api/v1/images/:hash/info`

- **响应**:
  ```json
  {
    "hash": "...",
    "file_name": "...",
    "mime_type": "...",
    "size": 1024,
    "width": 800,
    "height": 600,
    "description": "...",
    "tags": ["tag1", "tag2"],
    "routes": ["route1", "route2"],
    "created_at": "...",
    "updated_at": "..."
  }
  ```

#### 更新图片信息

`PATCH /api/v1/images/:hash`

- **请求**:
  ```json
  {
    "description": "New Description",
    "tags": ["new", "tags"],
    "file_name": "new_name.png",
    "routes": ["route1", "route2"] // 会覆盖原有路由
  }
  ```
- **响应**: 更新后的 Image 对象。

#### 删除图片

`DELETE /api/v1/images/:hash`

删除图片及其关联的缩略图和数据库记录。

---

### 2.4 路由管理

Base URL: `/api/v1/routes`

#### 获取所有路由

`GET /api/v1/routes`

列出系统中所有注册的图片路由别名。

- **响应**: `[ImageRoute Object]` 列表

#### 删除路由

`DELETE /api/v1/routes/:route`

仅删除路由别名，不删除对应图片。
