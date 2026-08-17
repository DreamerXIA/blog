# 04 — 日志类型过滤

**What to build:** 访客能按类型只看某类日志。

**Blocked by:** 02 — 日志闭环

**Status:** ready-for-agent

- [ ] `GET /api/logs?type=work|study|daily|summary` 只返回对应类型的日志
- [ ] 前端列表有类型过滤 tab（全部 / 工作 / 学习 / 日报 / 总结），点击切换列表
- [ ] API 层测试覆盖过滤逻辑
