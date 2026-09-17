# SHUZ-152 项目访问权限改造与部署记录

本文汇总 SHUZ-152 的产品规则、代码改动、数据变更、测试结果、部署状态和回退方式。仓库中不保存数据库口令、JWT、邮件凭据、固定验证码或生产 Compose 环境变量。

## 交付结论

- 项目创建人可在项目侧栏的“访问授权”中启用“仅限选中成员”，并维护可访问成员。
- 项目创建人始终保留访问权，不会被自己的授权名单锁在项目之外。
- 未授权成员看不到项目条目、标题或任何项目内容；直链及关联接口统一按资源不存在返回 `404`。
- `owner`、`admin` 不自动绕过项目权限，也不能代替创建人修改授权。
- 历史项目不再根据工作区角色、负责人或资源上传人推断创建人；创建人只采用新建时记录或工作区 owner 明确确认的数据。
- 功能已部署到正式入口 `https://team.shuzhixinghua.com`。原正式应用容器、切换前数据库备份和上传文件备份仍保留，可回退。

## 数据模型与迁移

### 迁移 500：项目权限字段

`project` 表新增：

- `created_by UUID`：项目创建人；历史未知数据允许为 `NULL`。
- `access_restricted BOOLEAN NOT NULL DEFAULT false`：是否仅允许指定成员访问。
- `allowed_user_ids UUID[] NOT NULL DEFAULT '{}'`：获得访问权的成员用户 ID。

迁移是加列操作，不创建 foreign key，不修改原项目的默认开放状态。旧版本应用可忽略这些字段，因此应用回退不要求立即恢复数据库。

### 迁移 501：纠正错误的历史归属

早期测试迁移曾把历史项目统一归属到工作区最早的 owner。迁移 501 清除该推断，同时保留迁移 500 之后新建项目的真实创建人。若历史项目已经启用限制访问，迁移会停止并要求先人工核实，避免删除有效授权管理人。

### 迁移 502：固化已确认创建人

工作区 owner 确认后的归属如下：

| 创建人 | 项目 |
| --- | --- |
| 张昕伟 | 艾多美、光彩养老事业 |
| 彭博 | AI Native、ModelForge、WOS、Everfree 雪蕾、GDM 饮食管理、上音、小礼树、胡庆余堂、普为、广科院智能推荐训练与推理系统、智能硬件、营养师系统 |

迁移使用工作区 UUID 和项目 UUID 双重限定，在其他部署中不会命中。回退脚本只清理由该迁移写入且仍保持预期创建人的记录，避免覆盖后续不同的人工调整。

正式库执行迁移前已另存上述项目记录；数据更新在单个事务内完成，并核对 14 个项目均已有创建人。项目的 `access_restricted` 和 `allowed_user_ids` 没有随创建人补录而改变。

## 后端改动

- 新增 `PUT /api/projects/{id}/access`，只允许当前项目创建人调用。
- 校验选中用户必须属于当前工作区；重复成员自动去重；单次最多 1000 人。
- task token 不能管理项目授权。
- 项目查询响应新增 `created_by`、`access_restricted`、`access_allowed`、`can_manage_access` 和 `allowed_user_ids`。
- 非创建人不会收到授权名单；未授权成员不会收到任何项目响应。
- 项目访问规则覆盖项目详情与修改、项目资源、任务详情与修改、列表、搜索、表格查询、附件下载、自动化与聊天项目上下文等路径。
- WebSocket 按每个接收者重新判断权限；保留权限者收到正常项目更新，失去权限者只收到不含项目内容的 `project:access_changed` 失效事件并清除缓存。
- 删除事件保持可达，避免被删除记录无法再用于权限判断时客户端留下陈旧数据。

## 前端与 API 客户端改动

- 项目详情侧栏新增“访问授权”设置，仅在 `can_manage_access=true` 时展示。
- 创建人可切换限制访问、勾选工作区成员并保存；创建人本人固定保留。
- 未授权项目不再进入客户端数据集；旧服务端缺少新字段时，Zod schema 使用兼容默认值，保证已安装客户端仍可工作。
- 更新英文、简体中文、日文和韩文文案。
- TanStack Query mutation 保存后刷新项目详情与相关列表缓存。

## 测试与验收

功能分支交付时通过：

- 前端测试 179 项。
- 后端测试 53 项。
- 实时通信测试 43 项。
- 双账号浏览器验收：创建人设置授权、成员访问、未授权拒绝、撤销后立即失效、非创建人无法修改授权。
- 数据迁移回归：全新安装不伪造创建人、历史错误归属可修复、重复执行安全、新项目真实创建人与已有访问策略不被覆盖。

重放到 2026-09-17 最新 `main` 后再次通过：

- Go：`TestProjectAccessLifecycle`。
- core schema：172 项测试。
- 项目授权 UI：2 个测试文件、7 项测试。
- `@multica/core` 与 `@multica/views` TypeScript 类型检查。
- `@multica/core` 与 `@multica/views` ESLint 检查为 0 error；输出的 warning 均来自最新 `main` 的既有文件。
- 迁移 500、501、502 的隔离 PostgreSQL 回归，包含重复执行、工作区边界与 down 回退。
- `git diff --check`。

## 测试环境记录

- 地址：`http://47.94.141.187:3100`。
- 使用独立前端、后端、PostgreSQL、Docker 网络、数据卷和网关。
- 测试数据库由正式库副本净化而来：停用自动化与集成、撤销运行时和任务令牌、取消未完成运行，并使用独立 JWT。
- 测试站已按要求移除额外 Basic Auth，直接进入 Multica 邮箱登录页。
- 测试环境与正式库隔离，不能把净化后的测试库覆盖回正式库。

## 正式环境记录

- 正式入口、HTTPS、邮箱登录、端口映射、运行时绑定、项目资源和上传文件保持不变。
- 当前应用容器为 `multica-production-shuz152-backend` 和 `multica-production-shuz152-frontend`。
- 原 `multica-backend-1`、`multica-frontend-1` 已停止但未删除；原镜像和配置已保留。
- 正式数据库仍为原 `multica-postgres-1`，没有用测试副本覆盖。
- 上线前保存数据库归档、上传文件归档、原容器元数据、配置指纹和 nginx 配置。
- 日常启动与回退命令见 `deploy/shuz152/PRODUCTION.md`；测试环境隔离要求见 `deploy/shuz152/README.md`。

## 回退策略

1. 应用回退：执行服务器保留的 `rollback.sh`，停止新容器并启动原前后端，继续使用当前数据库和上传文件。
2. 权限代码回退：旧应用忽略迁移 500 新增字段，无需恢复切换前数据库。
3. 创建人数据回退：迁移 502 的 down 文件仅清除仍与确认结果一致的 14 条归属。
4. 灾难恢复：仅在确认需要回到切换前完整数据时使用预迁移数据库与上传文件归档；普通应用回退不得执行此步骤。

## 相关文件

- `server/migrations/500_project_access.*.sql`：权限字段。
- `server/migrations/501_correct_historical_project_creators.*.sql`：历史错误归属修复。
- `server/migrations/502_backfill_confirmed_project_creators.*.sql`：已确认项目创建人。
- `server/internal/handler/project_access.go`：授权规则和更新接口。
- `server/internal/handler/project_access_realtime.go`：实时消息过滤。
- `server/internal/handler/project_access_test.go`：后端授权回归。
- `deploy/shuz152/PROJECT_ACCESS_INVISIBILITY_DESIGN.zh.md`：未授权项目完全隐藏的后端技术设计。
- `packages/views/projects/components/project-access-settings.tsx`：授权设置 UI。
- `deploy/shuz152/README.md`：测试环境隔离与操作说明。
- `deploy/shuz152/PRODUCTION.md`：正式环境切换与回退说明。
