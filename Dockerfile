# 多阶段构建：前端 React -> Go 后端 -> 运行时镜像
# 产物是单个静态 Go 二进制 + 打包后的前端静态文件，供 Render 等平台部署。

# ---------- 前端构建 ----------
FROM node:20-alpine AS web
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# ---------- 后端构建 ----------
FROM golang:1.26-alpine AS build
WORKDIR /app/server
ENV CGO_ENABLED=0
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN go build -o blog .

# ---------- 运行时 ----------
FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/server/blog ./blog
COPY --from=web /app/web/dist ./web/dist
ENV PORT=8080
EXPOSE 8080
CMD ["./blog"]
