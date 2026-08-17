# 03 — 个人信息展示与编辑

**What to build:** 主页展示个人信息（姓名 / 电话 / 邮箱 / 技术方向 / 学习目标），owner 可更新；首次运行有种子数据。

**Blocked by:** 01 — 项目脚手架与健康检查

**Status:** ready-for-agent

- [ ] 服务端有个人信息单行数据，首次运行自动种子默认值
- [ ] `GET /api/profile` 返回个人信息
- [ ] `PUT /api/profile` 更新个人信息，且需 owner token
- [ ] 前端主页顶部展示个人信息
- [ ] owner 可编辑并保存个人信息，刷新后更新生效
- [ ] API 层测试覆盖读写与鉴权
