# 02 — 日志闭环（核心 tracer bullet）

**What to build:** 用户能在页面上创建一条日志，内容保存到服务端 SQLite，刷新或换浏览器后仍能看到；写入受 owner token 保护。

**Blocked by:** 01 — 项目脚手架与健康检查

**Status:** ready-for-agent

- [ ] 服务端有日志的持久化表（type / title / content / created_at / updated_at）
- [ ] `POST /api/logs` 创建日志；缺失或错误 token 时拒绝写入
- [ ] `GET /api/logs` 返回时间倒序的日志列表
- [ ] `GET /api/logs/{id}` 返回单条日志
- [ ] 前端有「写日志」表单（选类型、填标题和正文），提交后列表立即出现该日志
- [ ] 刷新页面或换浏览器（无缓存）后该日志仍可见
- [ ] API 层测试覆盖：创建→持久化→列表、鉴权拒绝、标题/正文必填校验
