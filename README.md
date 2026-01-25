# AnzuImg

## 简介

一个基于 Go 语言 + Nuxt 打造的图床后。使用了[vipsgen](https://github.com/cshum/vipsgen)（[libvips](https://github.com/libvips/libvips)的 Go 语言实现），理论上能够良好地处理主流文件格式以及新兴的`webp`、`avif`、`jpxl`格式

## 部署

使用前请先配置.env文件，样例文件位于`deploy/.env.example`中

```
# PostgreSQL
ANZUIMG_DB_USER=anzuuser
ANZUIMG_DB_PASSWORD=anzupass
ANZUIMG_DB_NAME=anzuimg

# CORS Configuration
# CORS 配置，配置允许的Origin
ANZUIMG_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# Trusted proxies (Gin ClientIP)
# 逗号分隔；建议只信任宿主机 Nginx / 本机回环 / docker 内网段
ANZUIMG_TRUSTED_PROXIES=127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16

# Setup protection: require header X-Setup-Token for POST /api/v1/auth/setup
# 首次初始化时请临时设置一个随机值，初始化完成后可移除。
ANZUIMG_SETUP_TOKEN=

# Upload limits (MB)
ANZUIMG_MAX_UPLOAD_MB=110
ANZUIMG_MAX_UPLOAD_FILE_MB=60
ANZUIMG_MAX_UPLOAD_FILES=20

# Cookie SameSite: Lax / Strict / None
ANZUIMG_COOKIE_SAMESITE=Lax

# Passkey Configuration (WebAuthn/FIDO2)
ANZUIMG_PASSKEY_RP_ID=localhost
ANZUIMG_PASSKEY_RP_ORIGIN=http://localhost:8080
ANZUIMG_PASSKEY_RP_DISPLAY_NAME=AnzuImg

# Storage Configuration
# Storage type: "local" or "cloud"
ANZUIMG_STORAGE_TYPE=local
ANZUIMG_STORAGE_BASE=./data/images

# CloudFlare R1 Configuration
ANZUIMG_CLOUD_ENDPOINT=https://r2.cloudflarestorage.com
ANZUIMG_CLOUD_BUCKET=your-bucket-name
ANZUIMG_CLOUD_REGION=auto
ANZUIMG_CLOUD_ACCESS_KEY=your-access-key-id
ANZUIMG_CLOUD_SECRET_KEY=your-secret-access-key
ANZUIMG_CLOUD_USE_SSL=true

```

使用docker

```bash
docker compose -f deploy/docker-compose.yml -d
```

> [!NOTE]
>
> 首次运行，需要完成**系统初始化**
>
> 生产环境建议在 `.env` 中设置一次性初始化口令 `ANZUIMG_SETUP_TOKEN`（初始化完成后可移除），用于防止初始化接口被抢占.

> [!CAUTION]
>
> 如果没有配置 `ANZUIMG_SETUP_TOKEN`，后端将只接受本地路径访问进行初始化
> 务必正确配置CORS

### 生产部署建议

虽然说，Nuxt前端的Nitro服务器配置了Proxy代理，可以代理请求到后端，但是还是推荐把 **宿主机 Nginx 作为唯一对外入口**，即：

- `/api/**`、`/i/**` 直接反代到后端 `127.0.0.1:8080`
- 其他页面与静态资源反代到前端 `127.0.0.1:3000`

示例配置见 [`deploy/nginx/anzuimg.conf.example`](deploy/nginx/anzuimg.conf.example)。

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
