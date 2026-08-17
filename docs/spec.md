# 个人主页与成长日志系统 — MVP Spec

## Problem Statement

一位学习者需要向导师或团队负责人展示自己（是谁、技术方向、当前学习目标），并持续记录工作日志、学习日志、日报与阶段总结。目前仓库为空，没有任何可供展示或记录的系统，导师无法了解其近期进展、遇到的问题与下一步计划。

## Solution

交付一个最小可用的全栈系统：一个 React 单页前端，展示个人信息、支持「写日志」表单与带类型过滤的日志列表；一个 Go 后端，将数据持久化到 SQLite，并同时托管 JSON API 与打包后的前端静态文件。访客只读，Owner 通过 token 写入。可本地运行并部署到 Render 提供公网访问。

> 注：本文中「持久化到 SQLite」的决定已被 [ADR-0002](./adr/0002-postgres-neon.md) 取代（迁移到 Neon 托管 Postgres）；其余需求（领域模型、API 契约、权限、测试 seam）仍有效。

## User Stories

1. As a 访客，I want 打开页面即看到 Owner 的姓名、联系方式、技术方向与学习目标，so that 我能快速了解他是谁。
2. As a 访客，I want 看到按时间倒序排列的日志列表，so that 我能读到他的近期活动。
3. As a 访客，I want 按类型（工作/学习/日报/总结）过滤日志，so that 我能聚焦某类内容。
4. As a 访客，I want 点开一条日志查看完整内容，so that 我能了解细节。
5. As a 访客，I want 换设备或刷新后仍看到已发布内容，so that 我确认数据是服务端持久化的。
6. As a 访客，I want 页面可公网访问，so that 我不必本地启动就能查看。
7. As the Owner，I want 创建一条工作日志（选类型、填标题和正文），so that 我记录我的工作。
8. As the Owner，I want 创建一条学习日志，so that 我记录学习过程。
9. As the Owner，I want 创建一条日报，so that 我总结当天内容。
10. As the Owner，I want 创建一条阶段总结，so that 我记录阶段性结论。
11. As the Owner，I want 更新我的个人信息，so that 主页保持最新。
12. As the Owner，I want 我的写入受 token 保护，so that 随机访客不能篡改内容。
13. As the Owner，I want 首次运行即有种子个人信息，so that 主页不会在编辑前为空。
14. As the Owner，I want 本地以最少配置启动系统，so that 我能开发与验证。
15. As the Owner，I want 清晰的 README（运行/验证/技术选型/AI 使用），so that 我能说明交付。
16. As a 导师/团队负责人，I want 快速看到近期日报与总结，so that 我了解其进展与问题。
17. As a 导师/团队负责人，I want 一屏直达所有日报，so that 我不必逐条翻找每日进展。
18. As a 运维/平台，I want 一个健康检查端点，so that 我能确认服务在线。

## Implementation Decisions

- 架构：单一 Go 服务同时托管打包后的 React 静态文件与 JSON API（见 ADR 0001）。
- 持久化：SQLite，驱动使用 `modernc.org/sqlite`（纯 Go、免 CGO）（见 ADR 0001）。
- 领域模型：单一「日志」实体，用 `type` 枚举区分 `work` / `study` / `daily` / `summary`；「日报」是独立类型，不自动聚合当日其他日志。
- 日志字段：`id`、`type`、`title`（必填）、`content`（必填）、`created_at`、`updated_at`；无 tags、无独立 date 字段（日报的"当天"用 `created_at` 表示）。
- 个人信息字段：`name`、`phone`、`email`、`tech_direction`、`learning_goals`，单行数据，种子初始化默认值，Owner 可编辑。
- 权限：读取完全公开；写入（创建日志、更新个人信息）需在请求中携带 Owner token（后端环境变量中配置的密钥）。
- API 契约：
  - `GET /api/health` — 健康检查（公开）
  - `GET /api/profile` — 读取个人信息（公开）
  - `PUT /api/profile` — 更新个人信息（需 token）
  - `GET /api/logs?type=work|study|daily|summary` — 日志列表，时间倒序，可按类型过滤（公开）
  - `POST /api/logs` — 创建日志（需 token）
  - `GET /api/logs/{id}` — 日志详情（公开）
- 后端路由使用 Go 标准库 `net/http`（1.22+ 方法路由），不引入 gin/chi 等框架。
- `type` 枚举以英文字符串存储，前端映射为中文标签。
- 前端：React 18 + Vite，单页滚动布局，自上而下为「个人信息区 → 『写日志』按钮（展开内联表单）→ 日志列表（带类型过滤 tab）」。

## Testing Decisions

- 唯一自动化测试 seam：HTTP API 边界。用 `net/http/httptest` 对一个真实 SQLite 实例（临时文件或内存）发起请求，断言状态码、JSON 结构、跨请求持久化、鉴权拒绝、校验错误。
- 好测试的标准：只测公开 HTTP 契约与外部可见行为，不窥探数据库内部或 handler 结构体。
- 被测模块：profile 的 GET/PUT、log 的 GET 列表（含 type 过滤）/POST/详情、token 鉴权、健康检查。
- 先例：无（全新项目），本次确立「以 API 层测试为准」的惯例。
- 前端通过浏览器手工验证（Playwright 可用于验证，但不作为 MVP 的自动化测试 seam）。

## Out of Scope

- 用户注册/登录体系（仅 owner 单 token）。
- 日志的编辑与删除（MVP 仅创建 + 查看；编辑/删除属于加分方向，暂缓）。
- 日报自动聚合。
- 标签、图片、评论、分页。
- 托管 Postgres 或持久盘（SQLite 临时盘的取舍已在 ADR/README 说明）。
- 多路由前端（已选单页滚动）。
- 完整的自动化 E2E 测试套件。

## Further Notes

- 部署目标为 Render，需重新登录 `gh` 并新建仓库供 Render 连接；SQLite 在免费实例的临时盘特性需在 README 中说明。
- README 需包含 AI 使用说明。
- 已形成的领域词汇表见 `CONTEXT.md`，架构决策见 `docs/adr/0001-single-service-sqlite.md`。
