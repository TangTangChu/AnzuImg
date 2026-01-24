# AnzuImg

## 简介

一个基于 Go 语言 + Nuxt 打造的图床后。使用了[vipsgen](https://github.com/cshum/vipsgen)（[libvips](https://github.com/libvips/libvips)的GO语言实现），理论上能够良好地处理主流文件格式以及新兴的`webp`、`avif`、`jpxl`格式

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
> 首次运行，前端将引导进行**初始密码设置**

> [!CAUTION]
>
> 务必正确配置CORS

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
