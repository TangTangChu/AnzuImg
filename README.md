# AnzuImg

## 简介

一个基于 Go 语言 + Nuxt 打造的图床后。使用了[vipsgen](https://github.com/cshum/vipsgen)（[libvips](https://github.com/libvips/libvips)的 Go 语言实现），理论上能够良好地处理主流文件格式以及新兴的`webp`、`avif`、`jpxl`格式

## 部署

使用前请先配置.env文件，样例文件位于`deploy/.env.example`中

```

# 后端核心配置

# 监听地址
ANZUIMG_SERVER_ADDR=:8080
# 优雅停机等待秒数，默认 10
ANZUIMG_SHUTDOWN_TIMEOUT_SEC=10

# API 前缀配置
# API 路由前缀，例如 /kotori，留空表示无前缀
# 默认为空，此时 API 路由为 /api/v1/...
# 设置为 /kotori 时，API 路由为 /kotori/api/v1/...
ANZUIMG_API_PREFIX=

# 数据库配置
# 主机地址，Docker 部署通常填 db
ANZUIMG_DB_HOST=db
ANZUIMG_DB_PORT=5432
ANZUIMG_DB_USER=anzuuser
ANZUIMG_DB_PASSWORD=anzupass
ANZUIMG_DB_NAME=anzuimg
# SSL 模式，支持 disable require verify-full，默认 disable
ANZUIMG_DB_SSLMODE=disable

# 存储配置
# 存储类型，local 或 cloud，默认 local
ANZUIMG_STORAGE_TYPE=local
# 本地存储路径，仅在 STORAGE_TYPE=local 时使用
ANZUIMG_STORAGE_BASE=/data/images

# S3 或 S3 兼容云存储配置，仅在 STORAGE_TYPE=cloud 时使用
ANZUIMG_CLOUD_ENDPOINT=s3.amazonaws.com
ANZUIMG_CLOUD_BUCKET=anzuimg-bucket
ANZUIMG_CLOUD_REGION=us-east-1
ANZUIMG_CLOUD_ACCESS_KEY=
ANZUIMG_CLOUD_SECRET_KEY=
ANZUIMG_CLOUD_USE_SSL=true

# 网络与安全配置
# 允许跨域访问的源，逗号分隔，填写前端访问地址
ANZUIMG_ALLOWED_ORIGINS=http://localhost:9200
# 信任代理网段，用于获取真实 IP，逗号分隔
ANZUIMG_TRUSTED_PROXIES=127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
# 初始化设置 Token，留空则不校验，首次部署建议设置
ANZUIMG_SETUP_TOKEN=
# Cookie SameSite 策略，Lax Strict None，默认 Lax
ANZUIMG_COOKIE_SAMESITE=Lax
# 是否启用会话严格 IP 绑定，默认 false
ANZUIMG_STRICT_SESSION_IP=false

# Passkey WebAuthn 配置
# Relying Party ID，通常为域名
ANZUIMG_PASSKEY_RP_ID=localhost
# Relying Party Origin，浏览器访问的完整 URL
ANZUIMG_PASSKEY_RP_ORIGIN=http://localhost:9200
# Relying Party Display Name，注册时显示的应用名称
ANZUIMG_PASSKEY_RP_DISPLAY_NAME=AnzuImg

# 上传限制配置
# 单次请求的最大体积，单位 MB，默认 110
ANZUIMG_MAX_UPLOAD_MB=110
# 单个文件的最大体积，单位 MB，默认 60
ANZUIMG_MAX_UPLOAD_FILE_MB=60
# 单次请求的最大文件数，默认 20
ANZUIMG_MAX_UPLOAD_FILES=20

# 前端配置
# 后端 API 地址
ANZUIMG_FRONTEND_BACKEND_URL=http://backend:8080
# 前端子路径部署，例如 /app/，建议以 / 结尾
ANZUIMG_FRONTEND_APP_BASE_URL=/clannd/
# 前端 API 路由前缀，应与 ANZUIMG_API_PREFIX 一致
# 留空或 / 表示不加前缀，此时前端 API 为 /api/v1/...
# 设置为 /kotori 时，前端 API 为 /kotori/api/v1/...
ANZUIMG_FRONTEND_API_PREFIX=/kotori

```

使用docker

```bash
docker compose -f deploy/docker-compose.yml up -d
```

> [!CAUTION]
>
> 首次运行，前端将引导进行初始化密码，如果没有配置 `ANZUIMG_SETUP_TOKEN`，后端将只接受本地路径访问进行初始化
>
> 如果使用了CDN服务，请关闭严格IP模式，否则无法正常访问控制服务
>
> 务必正确配置CORS
>
> cloud存储策略我没有测试过

### 生产部署建议

虽然 Nuxt 前端的 Nitro 支持 proxy，但生产环境仍建议把 **宿主机 Nginx 作为唯一对外入口**。

本项目支持子路径部署和 API 前缀，推荐对外暴露 3 个入口：

- **图床前端**：`/clannd/`
- **图床后端 API**：`/kotori/`，通常只承载 `/api/v1/*` 与 `/health`
- **图片直链**：`/i/`

对应前端环境变量：

- `ANZUIMG_FRONTEND_APP_BASE_URL=/clannd/`，前端挂载路径，建议以 `/` 结尾
- `ANZUIMG_FRONTEND_API_PREFIX=/kotori`，置空或 `/` 表示不加前缀
- `ANZUIMG_FRONTEND_BACKEND_URL=http://backend:8080`，前端 SSR 和 Nitro 侧用于 proxy 的后端地址

后端如需原生前缀支持可设置：

- `ANZUIMG_API_PREFIX=/kotori`，后端路由将变为 `/kotori/api/v1/...`

示例 Nginx 配置见 [`deploy/nginx/anzuimg.conf.example`](deploy/nginx/anzuimg.conf.example)。

## API

请阅读[API文档](README/API.md)

## 示例

### 登录页

![image1](README/image1.webp)

登录的时候支持密码登录与Passkey登录（Passkey需要自行注册）

### 图库页

![image7](README/image7.webp)

点击图片可以查看详情

![image5](README/image5.webp)

> [!TIP]
>
> 在此模态框下，可以拖拽图片、缩放图片、编辑图片信息等操作

### 上传页

![image2](README/image2.webp)

> [!TIP]
>
> 上传页支持**多文件上传**，也支持对图片格式进行**转换**

![image8](README/image8.webp)

### 路由管理页

![image3](README/image3.webp)

> [!TIP]
>
> 路由就是把图片映射到指定路径，即`/i/r/指定路由`下，这样可以不用一大串哈希值，并且可以实现换图片但url不更换（比如头像url）
>
> - 通过路由访问
>
> ![image9](README/image9.webp)
>
> - 通过哈希值访问
>
> ![image10](README/image10.webp)

### 设置页

![image4](README/image4.webp)
