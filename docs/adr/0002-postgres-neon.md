---
status: accepted
---

# SQLite → Neon Postgres 迁移

系统原用 SQLite（`modernc.org/sqlite`）持久化，但 Render 免费实例的磁盘是临时盘，重新部署/重启即清空，导致线上数据无法持久。为让数据真正保留，迁移到 Neon 托管的 Postgres（免费档 $0，挂起后数据仍保留）。代价：本地开发与测试需连 Neon 云（本机无 Postgres、无 Docker），且 Neon 免费档 5 分钟无活动即挂起、首次请求冷启动约几百毫秒到几秒。

## Considered Options

- **Render 持久盘**：SQLite 文件落持久盘即零代码改动，但需升级付费实例（Starter 起）。否：要花钱。
- **Render 托管 Postgres**：同样要重写 SQL，且比 Neon 贵。否。
- **SQLite/Postgres 双模式**（本地 SQLite + 生产 PG）：本地零依赖，但需维护两套 SQL 方言，且测试跑 SQLite、生产跑 PG，正是本项目测试决策里警告的「测试与生产发散」。否。

## 关键子决策

- **驱动**：`github.com/jackc/pgx/v5/stdlib`（`database/sql` 适配器，保留 `Store{ db *sql.DB }` 结构），弃 lib/pq（维护模式）。需把 `?` 占位符改 `$n`、`LastInsertId()` 改 `RETURNING id`、`AUTOINCREMENT` 改 identity。
- **本地开发**：全切 Postgres，本地直接连 Neon 云（dev 用独立库），不保留 SQLite 双模式。
- **测试**：连 Neon 测试库（`TEST_DATABASE_URL`），`newTestServer` 用 `TRUNCATE ... RESTART IDENTITY` 重置隔离；缺 `TEST_DATABASE_URL` 时 `t.Fatal` 报错而非 skip。
- **时间戳**：保持 TEXT（RFC3339），API 契约不变。
- **建表**：沿用启动时 `CREATE TABLE IF NOT EXISTS` + seed（`ON CONFLICT DO NOTHING` 幂等），不引入迁移工具。
- **健康检查**：`/api/health` 不 ping DB，避免 Render 探活打醒 Neon 烧算力。

## Consequences

- 数据真正持久，不再随 Render 部署丢失。
- 本地 `go test` 需联网 + 一个 Neon 测试库，冷启动约 1s。
- Render 免费档自身的 15 分钟无流量 spin down 仍在——Neon 解决的是数据丢失，不是冷启动。
