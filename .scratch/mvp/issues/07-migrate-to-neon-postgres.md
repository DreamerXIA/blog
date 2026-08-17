# 07 — 迁移到 Neon Postgres 持久化

**What to build:** 将持久化从 SQLite 迁移到 Neon 托管 Postgres，使数据在 Render 重新部署/重启后仍保留；对外 API 契约与前端交互完全不变。

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

## Problem Statement

系统当前用 SQLite（`modernc.org/sqlite`）持久化，但 Render 免费实例的磁盘是临时盘，重新部署或重启即清空。结果是：Owner 辛苦记录的工作日志、学习日志、日报、阶段总结，以及维护好的个人信息，会在下一次部署后「回档」——访客/导师看到的内容退回到种子默认值。数据无法持久是当前最影响使用的问题。

## Solution

把持久化从 SQLite 迁移到 Neon 托管的 Postgres（免费档 $0，挂起后数据仍保留）。数据落到云端持久化，重新部署/重启后仍在。对外 API 契约、前端交互、领域模型全部不变——访客和 Owner 感知不到底层数据库换了，只是数据不再丢。

## User Stories

1. As a 访客，I want 打开页面即看到 Owner 已发布的日志与个人信息，so that 我能快速了解他是谁。
2. As a 访客，I want 服务重新部署或重启后这些内容仍完整保留（不回档到默认值），so that 我读到的进展是连续的。
3. As a 访客，I want 换设备或刷新后仍看到相同内容，so that 我确认数据是服务端持久化的。
4. As a 访客，I want 按类型过滤、点开日志详情的行为与迁移前完全一致，so that 我无感知地继续使用。
5. As the Owner，I want 我创建的工作日志在部署后仍存在，so that 我不必反复重录。
6. As the Owner，I want 我创建的学习日志、日报、阶段总结在重启后仍存在，so that 记录不丢失。
7. As the Owner，I want 我更新的个人信息在重启后仍保留，so that 主页不会退回种子默认值。
8. As the Owner，I want 写入仍受 token 保护，so that 迁移后权限模型不变。
9. As the Owner，I want 本地开发用与线上一致的数据层（连云端 Postgres），so that 本地验证的行为与生产一致、避免测试与生产发散。
10. As the Owner，I want 首次连接空数据库时自动建表并种子个人信息，so that 新库也能直接可用。
11. As the Owner，I want 自动化测试连一个独立的 Postgres 测试库，so that 测试跑在真实数据库上而非 mock 或 SQLite。
12. As the Owner，I want 现有 API 契约（路径、请求/响应 JSON）完全不变，so that 前端无需改动。
13. As 导师/团队负责人，I want 近期日报与总结在任意时间访问都完整可读，so that 我能持续追踪进展。
14. As 运维/平台，I want 健康检查端点仍可用且不因数据库探活增加成本，so that 在线判断不依赖数据库唤醒。

## Implementation Decisions

- 持久化驱动从 `modernc.org/sqlite` 换成 `pgx/v5/stdlib`（`database/sql` 适配器），保留现有 store 结构；弃 lib/pq（维护模式）。
- 全切 Postgres，不保留 SQLite 双模式；本地开发与生产都连 Neon（dev 用独立库）。
- 连接配置通过 `DATABASE_URL`（生产）与 `TEST_DATABASE_URL`（测试）环境变量传入，Neon 连接串需含 SSL。
- 领域模型不变：日志（type 枚举 work/study/daily/summary）、个人信息（单行）字段与 JSON 结构完全一致。
- schema：沿用启动时自动建表 + seed；DDL 由 SQLite 方言改写为 Postgres（自增主键用 identity）；seed 用 `ON CONFLICT DO NOTHING` 保证幂等。
- SQL 方言改写：`?` 占位符 → `$n`；`LastInsertId()` → `RETURNING id`。
- 时间戳保持 TEXT（RFC3339 UTC），列表排序仍按 id 倒序。
- 健康检查不触碰数据库，避免 Render 探活打醒 Neon 烧算力。
- 旧 SQLite 临时盘数据弃置，不写迁移脚本。
- 部署配置新增 `DATABASE_URL`。

## Testing Decisions

- 唯一测试 seam 仍是 HTTP API 边界（与现有测试一致），用 `httptest` 对真实 Postgres（Neon 测试库）发起请求。
- 好测试只断言外部可见行为：状态码、JSON 结构、跨请求持久化、鉴权拒绝、校验错误；不窥探数据库内部或 handler 结构。
- 测试隔离：`newTestServer` 连 `TEST_DATABASE_URL` 后 `TRUNCATE ... RESTART IDENTITY` 重置；缺 `TEST_DATABASE_URL` 时 `t.Fatal` 报错而非 skip。
- 先例：现有 `handlers_test.go` 的 API 层测试（健康检查、创建→持久化→列表、鉴权拒绝、校验、过滤）——断言不变，仅底层数据库换成 Postgres。

## Out of Scope

- 日志编辑与删除（仍不在 MVP）
- 日报自动聚合
- 标签、图片、评论、分页
- 引入迁移工具（golang-migrate / goose）
- 时间戳切换为 TIMESTAMPTZ
- 消除 Render 免费实例的 15 分钟冷启动（Neon 只解决数据持久化，不解决冷启动）
- SQLite/Postgres 双模式支持
- 现有 SQLite 数据的迁移脚本

## Further Notes

- 需 Owner 先在 Neon 控制台建项目并拿到生产、测试两个连接串，否则本地 `go test` 会因缺 `TEST_DATABASE_URL` 报错中止。
- Neon 免费档 5 分钟无活动即挂起，首查冷启动约几百毫秒；Render 免费档仍有 15 分钟无流量 spin down。
- 相关决策见 `docs/adr/0002-postgres-neon.md`；单服务架构决策（ADR-0001）仍有效。
