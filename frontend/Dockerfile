FROM node:24-slim AS builder

WORKDIR /app

RUN npm config set registry https://registry.npmmirror.com

RUN npm install -g pnpm

RUN pnpm config set registry https://registry.npmmirror.com

COPY frontend/AnzuImg/package.json frontend/AnzuImg/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/AnzuImg .

ARG BACKEND_URL=http://backend:8080
ENV BACKEND_URL=${BACKEND_URL}

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
