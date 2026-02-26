FROM node:24-slim AS builder

WORKDIR /app

RUN npm config set registry https://registry.npmmirror.com

RUN npm install -g pnpm

RUN pnpm config set registry https://registry.npmmirror.com

COPY frontend/AnzuImg/package.json frontend/AnzuImg/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/AnzuImg .

ARG ANZUIMG_FRONTEND_BACKEND_URL=http://backend:8080
ARG ANZUIMG_FRONTEND_APP_BASE_URL=/
ARG ANZUIMG_FRONTEND_API_PREFIX=/kotori

ENV ANZUIMG_FRONTEND_BACKEND_URL=${ANZUIMG_FRONTEND_BACKEND_URL}
ENV ANZUIMG_FRONTEND_APP_BASE_URL=${ANZUIMG_FRONTEND_APP_BASE_URL}
ENV ANZUIMG_FRONTEND_API_PREFIX=${ANZUIMG_FRONTEND_API_PREFIX}

RUN pnpm build

FROM node:24-slim

WORKDIR /app

COPY --from=builder /app/.output /app/.output

ENV PORT=3000
ENV HOST=0.0.0.0
ENV NUXT_HOST=0.0.0.0
ENV NUXT_PORT=3000

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
