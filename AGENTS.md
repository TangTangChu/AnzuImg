# AGENTS.md

这是AnzuImg的Agents.md

## 总体开发要求

- 对于后端的验证，请使用docker compose up -d --build backend进行构建验证，Windows上没有相关环境，Build和test可能会出问题，所以必须使用docker
- 对于前端的验证，请使用pnpx nuxt typeckeck，不可使用pnpm build
- 安全性优先，对于引入的端口需要考虑安全性
- 不可留历史包袱，不写向后兼容代码（但是对于一个接口同时允许POST/PUT的这种情况是允许的，因为这是生产环境所限）

## 具体的前端要求

- UX设计时，务必遵循统一的UX风格，本项目虽然使用了MD3颜色，但是不代表遵循MD3规范，本项目不严格遵循MD3规范
- 禁止使用hover时放大、上浮效果，禁止使用渐变色、强阴影
- 禁止边框包裹边框的设计
