# 01 — 项目脚手架与健康检查

**What to build:** 初始化仓库与前后端骨架，使 Go 服务启动后能返回健康检查，React 应用能在开发模式下加载并触达后端 API，证明前后端联通。

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] `git init` 完成，仓库采用前端与后端分离的清晰目录结构
- [ ] Go 模块初始化，`GET /api/health` 返回 200 与预期 JSON
- [ ] React（Vite）应用能启动，开发模式下 `/api` 请求代理到 Go 后端
- [ ] 前端能成功调用 `/api/health` 并展示结果（证明前后端联通）
