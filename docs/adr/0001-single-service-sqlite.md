# 单服务架构 + SQLite 持久化

系统采用单一 Go 服务，同时托管 React 打包后的静态前端与 JSON API，持久化使用 SQLite（`modernc.org/sqlite` 纯 Go 驱动）。选单服务是因为 MVP 规模下前后端分离托管（Vercel + 独立 Go API）只会增加部署与跨域复杂度，而单服务一个进程即可完成「前端页面 → API → 持久化」全部闭环；选 SQLite 而非托管 Postgres 是为了零外部依赖、本地零配置启动，代价是线上免费实例磁盘临时性带来的数据可重置风险，需在 README 中说明并在后续必要时切换到持久盘或托管 Postgres。
