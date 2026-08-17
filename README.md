# 个人主页与成长日志

一个展示个人信息、记录工作日志 / 学习日志 / 日报 / 阶段总结的全栈系统。访客只读查看，Owner 通过 token 写入。导师或团队负责人打开页面即可快速了解「你是谁、最近在做什么、遇到什么问题、下一步计划做什么」。

- 前端：React 18 + Vite（单页滚动布局）
- 后端：Go（标准库 `net/http`，Go 1.22+ 方法路由）
- 持久化：PostgreSQL（Neon 托管，`pgx/v5` 驱动）
- 架构：单一 Go 服务同时托管打包后的前端静态文件与 JSON API

## 线上访问

- 公网地址：https://blog-6f1n.onrender.com
- 免费档无流量约 15 分钟后 spin down，首次访问有冷启动延迟（Neon 免费档 5 分钟无活动亦会挂起）。

## 本地启动

### 环境要求

- Go 1.22+（本项目使用 1.26）
- Node.js 18+ 与 npm
- 无需本地安装数据库，但需一个可连的 Postgres（Neon 免费档即可），通过 `DATABASE_URL` 指定

### 方式一：开发模式（前后端分离，热更新）

分别开两个终端：

```bash
# 终端 1：启动后端 API（默认 :8080）
cd server
go run .

# 终端 2：启动前端 Vite dev server（默认 :5173，/api 代理到 :8080）
cd web
npm install
npm run dev
```

浏览器打开 http://localhost:5173 。

### 方式二：生产形态（单服务，一条命令）

```bash
./run.sh
```

等价于「构建前端 → `go run .` 托管前端静态文件与 API」，打开 http://localhost:8080 即可看到完整系统。

### 环境变量

| 变量 | 默认值 | 说明 |
| :--- | :--- | :--- |
| `PORT` | `8080` | 服务监听端口 |
| `OWNER_TOKEN` | `dev-token` | 写入（创建日志 / 更新个人信息）所需的 token，生产环境务必改为强密钥 |
| `DATABASE_URL` | 无（必填） | Postgres 连接串（Neon 控制台获取，需含 SSL） |
| `TEST_DATABASE_URL` | 无 | 测试用 Postgres 连接串；缺省时 `go test` 报错中止 |
| `STATIC_DIR` | 自动探测 | 前端静态目录；默认依次尝试 `../web/dist`、`web/dist` |

## 技术选型与原因

- **React**：需求指定前端必须使用 React，生态成熟、组件化便于组织单页布局。
- **Go 后端**：需求指定后端使用 Go。理由：单一静态二进制部署简单、标准库 `net/http`（1.22+ 方法路由 + 路径参数）对 MVP 的 6 个 API 已足够，无需引入 gin/chi 等框架；并发模型适合作为单服务同时承载静态文件与 API。
- **PostgreSQL（Neon 托管，`pgx/v5` 驱动）**：数据持久化到托管 Postgres，不再随 Render 临时盘重置。详见 `docs/adr/0002-postgres-neon.md`。代价是 Neon 免费档 5 分钟无活动即挂起、首次请求冷启动约几百毫秒，且本地开发/测试需联网连 Neon 云。
- **单服务架构**：MVP 下前后端分离托管（如 Vercel + 独立 API）只会增加部署与跨域复杂度，单进程即可完成「前端页面 → API → 持久化」完整闭环。详见 `docs/adr/0001-single-service-sqlite.md`。

## 数据模型

两个表，均由服务首次启动时自动建表并种子初始化：

**`logs`（日志）**

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | BIGINT | 主键，自增（identity） |
| `type` | TEXT | 枚举：`work` / `study` / `daily` / `summary` |
| `title` | TEXT | 标题（必填） |
| `content` | TEXT | 正文（必填） |
| `created_at` | TEXT | 创建时间（RFC3339 UTC） |
| `updated_at` | TEXT | 更新时间 |

**`profile`（个人信息，单行）**

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | BIGINT | 恒为 1，保证单行 |
| `name` | TEXT | 姓名 |
| `phone` | TEXT | 电话 |
| `email` | TEXT | 邮箱 |
| `tech_direction` | TEXT | 技术方向 |
| `learning_goals` | TEXT | 学习目标 |

「日报」是独立日志类型（`type = daily`），不自动聚合当日其他日志。

## API 契约

| 方法 | 路径 | 鉴权 | 说明 |
| :--- | :--- | :--- | :--- |
| GET | `/api/health` | 公开 | 健康检查，返回 `{"status":"ok"}` |
| GET | `/api/profile` | 公开 | 读取个人信息 |
| PUT | `/api/profile` | token | 更新个人信息 |
| GET | `/api/logs?type=work\|study\|daily\|summary` | 公开 | 日志列表，时间倒序，可按类型过滤 |
| POST | `/api/logs` | token | 创建日志 |
| GET | `/api/logs/{id}` | 公开 | 日志详情 |

鉴权：写入请求需携带请求头 `Authorization: Bearer <OWNER_TOKEN>`。

## 需求理解

核心目标是让「别人」（导师 / 团队负责人）无需任何说明即可快速了解一个人：

1. **是谁** —— 个人信息（姓名、联系方式、技术方向、学习目标）；
2. **最近在做什么** —— 工作日志、学习日志；
3. **每天进展与问题** —— 日报；
4. **阶段性结论** —— 阶段总结。

因此信息架构自上而下为「个人信息区 → 写日志入口 → 带类型过滤的日志列表」，访客一屏即可覆盖「了解一个人 + 浏览其近期活动」，Owner 通过 token 维护内容。

## 核心用户流程

- **访客**：打开页面 → 看到个人信息 → 浏览按时间倒序的日志列表 → 点 tab 按类型过滤 → 点某条日志展开完整内容 → 刷新 / 换设备内容仍在（服务端持久化）。
- **Owner**：点「写日志」→ 选类型、填标题正文、填 token → 提交后列表立即出现；点「编辑个人信息」→ 修改字段、填 token → 保存后主页更新。

## 已完成功能

- 个人信息展示与 Owner 编辑（首次运行自动种子默认值）
- 日志创建（工作 / 学习 / 日报 / 总结）、列表展示（时间倒序）、详情展开
- 按类型过滤日志
- 写入受 Owner token 保护，缺失或错误 token 返回 401
- 服务端持久化（PostgreSQL/Neon），刷新 / 换设备后内容可见
- 健康检查端点 `/api/health`
- 生产形态单服务（Go 托管前端静态文件 + API）
- API 层自动化测试（`net/http/httptest` 覆盖读写、鉴权拒绝、校验错误、过滤）

## 未完成内容与后续计划

- 日志的编辑与删除（MVP 仅「创建 + 查看」）
- 日报格式化 / 一键复制汇报内容
- 日志分页、标签、图片、评论
- Render 免费实例无流量约 15 分钟后 spin down，首次访问有冷启动延迟；Neon 免费档 5 分钟无活动即挂起（数据不丢，但首查约几百毫秒）
- 完整的 E2E 测试（当前以 API 层测试为准，前端以浏览器手工验证）

## AI 使用说明

本项目使用 Claude Code 完成开发：根据 `docs/spec.md` 与 `need.md` 生成领域模型、API 契约与实现方案；AI 产出 Go 后端（store / handlers / 路由）、React 前端与 API 测试，并由人工审查、调整与验证。测试与浏览器验证过程中发现的编码 / 路径 / 静态目录等问题由 AI 与人工共同修正。所有 AI 生成的代码均经 `go vet`、`go test` 及浏览器手工验证通过。

## 验证方式

```bash
# 后端 API 测试（需先 export TEST_DATABASE_URL=<Neon 测试库连接串>）
cd server && go test ./...

# 手动验证核心闭环（生产形态）
./run.sh
# 浏览器打开 http://localhost:8080，依次验证：
# 1. 看到个人信息
# 2. 写日志（token 用 dev-token）→ 列表出现
# 3. 刷新页面 → 日志仍在
# 4. 切换类型 tab → 过滤生效
```
