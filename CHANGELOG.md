# 更新日志

## [1.3.1] - 2026-09-22

### Add

- **HostVars.port**：支持未配置连接端口时，使用默认值`22`

### Fixed

- 修复密码加密及解密时，密码可能会留存在 shell 历史记录导致密码泄露的问题

---

## [1.3.0] - 2026-09-21

### Add

- **批量下载**：新增 `fetch` 任务，从远程主机下载文件到本地
  - `-f, --file` 远程文件路径
  - `-d, --dest` 本地目标路径
  - `-F, --force` 覆盖本地已存在文件
  - `-B, --backup` 覆盖前备份本地原文件
  - `dest` 以 `/` 结尾时，按主机标识分组存放：`dest/<主机标识>/<文件名>`
  - 本地目录不存在时自动创建
  - 不支持远程目录，需先在远程打包
- **SSH 会话工厂**：抽取 `SshSessionFactory`，统一连接、认证、超时
- **私钥认证**：支持 `identity_file` 配置，`~` 自动展开
- **私钥口令**：支持 `passphrase` 配置，与密码同样加密存储

### Changed

- **SSH 工具类签名**：`JschCommandExecutor` / `JschFileUploader` 改为接收 `HostVars`
- **`SshUserInfo` 增强**：支持 passphrase 与终端交互输入
- **输出日志**: 优化了输出内容可读性

### Fixed

- 修复跨组主机名重复导致的静默覆盖问题，主机标识改为 `组名/主机名`

---

## [1.2.1] - 2026-09-20

### Add

- **加解密工具独立入口**：新增 taskssh encrypt / taskssh decrypt 子命令，无需依赖清单文件即可单独使用
- **SSH登录**：支持终端交互输入密码

### Changed

- **`push` 不再自动解压**：上传压缩包后不执行解压，如需解压请在任务中显式添加 `command` 步骤，执行远程解压命令

### Fixed

- 修复 `JschCommandExecutor` 中输出重定向可能会导致输出信息异常的错误

---

## [1.2.0] - 2026-09-19

### Breaking Changes

- **`push` 默认行为变更**：远程文件已存在时不再自动备份覆盖，改为报错。需显式使用 `-F` 才覆盖，`-F -B` 备份后覆盖
- **变量严格模式**：未解析的 `${var}` 会直接报错，不再静默保留

### Add

- **覆盖策略**：`push` 新增 `-F` / `-B` 参数，控制覆盖和备份行为
- **步骤等待**：`Step` 新增 `delay` 字段，执行后按配置等待再进入下一步
- **CLI 值支持变量**：CLI 参数值可引用 `${var}`，每台主机按各自变量展开

### Changed

- **变量优先级调整**：CLI 参数 > `step.with` > 变量池（`global_vars` / `group` / `host`）
- **`step.with` 定位调整**：只作参数容器，不再参与变量池
- **CLI 结构优化**：参数解析不再受 `-i` 位置限制，`taskssh -h` 可正常输出帮助

### Fixed

- 修复 `-e` 无法覆盖内置任务中 `${command}` 的问题
- 修复变量未解析时静默保留的问题
- 修复 `taskssh -h` 被误判为任务名的问题
- 修复大输出场景下远程命令退出码可能取错的问题
- 修复备份文件命名可能重复的问题

---

## [1.1.1] - 2026-09-17

### Add

- **编码规范**：新增 `CONTRIBUTING.md`，遵循 `Alibaba Java Coding Guidelines`
- **覆盖策略**：`push` 新增 `-F` / `-B` 参数
- **步骤等待**：`Step` 新增 `delay` 字段

### Changed

- **`push` 默认行为变更**：由"自动备份 + 覆盖"改为"已存在则报错"
- **变量优先级调整**：CLI 参数优先级最高

### Fixed

- 修复 `-e` 无法覆盖内置任务中 `${command}` 的问题
- 修复变量未解析时静默保留的问题

---

## [1.1.0] - 2026-09-16

### Add

- **任务机制**：支持将多个原子操作串联为有序流程
- **Action 抽象**：内置 `command`、`push` 两个 Action
- **内置流程**：`command` / `push` / `deploy`
- **自定义流程**：在清单的 `tasks` 段中定义
- **CLI 入口统一**：`taskssh <task> [hosts...] [options]`
- **变量递归替换**：`${key}` 最多展开 4 层
- **Action 帮助**：`taskssh -h` 列出所有 Action 及参数
- **统一异常**：`TaskException` 携带主机名与步骤名

### Changed

- **CLI 结构重构**：由「命令 + 参数」改为「流程名 + 参数」
- **`command` / `push` / `deploy` 降级为内置流程**，调用方式不变

### Removed

- `DeployCommand` 等独立命令类
- `TaskSSHCommand` 抽象基类

---

## [1.0.0] - 2024-05-20

### 初始版本

- 基于 JSch 的轻量级 SSH 运维工具
- 支持 `push` / `command` / `deploy`
- 支持 YAML 清单文件定义服务器组与主机
- 支持变量替换、密码加密、执行摘要

---