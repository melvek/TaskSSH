# 更新日志

## [2.1.0] - 2026-09-25

### Breaking Changes

- **vault 密钥外置**：加密密钥不再硬编码在二进制中
  - 默认交互输入
  - 支持 `--secret-key-file` 指定密钥文件（普通文本或可执行脚本），默认位置 `~/.taskssh/vault-key`
  - **旧密文无法解密**，需用 `taskssh encrypt` 重新加密
  - **移除 `push -B/--backup`**：不再支持覆盖前自动备份
  - 需要备份时，在任务里增加 `command` 步骤执行备份命令
  - **移除 `fetch -F/--force` 和 `fetch -B/--backup`**：
    - `fetch` 默认覆盖本地已存在文件
  - **强制密码加密**：清单文件中的 `password` 必须是 `taskssh encrypt` 生成的密文

### Add

- **script 任务**：新增 `script` action，将本地 shell 脚本上传到远程主机并执行
  - CLI：`taskssh script <host> -f <file> [-d <dest>] [-R] [-F]`
  - 清单：`action: script`，参数 `file` / `dest` / `remove` / `force`
  - 默认上传到 `/tmp`，执行后可选删除，覆盖已存在文件可选
- **push 目录上传**：`push` 支持目录
  - 不传 `-z`：递归逐文件上传，不依赖远程工具
  - 传 `-z`：本地 zip 打包上传，远程 `unzip` 解包，依赖远程 `unzip`
- **fetch 目录下载**：`fetch` 支持目录
  - 不传 `-z`：递归逐文件下载，不依赖远程工具
  - 传 `-z`：远程 `zip -r` 打包，下载后本地 `archive/zip` 解压，依赖远程 `zip`
  - 新增 `-T/--tmp-dir`：指定远程临时目录，默认 `/tmp`，仅 `-z` 时生效
- **连接复用**：每台主机连接一次，执行完整条任务后关闭，连接次数从「主机数 × step 数」降为「主机数」
- **`--connect-timeout`**：新增连接超时配置，单位秒，默认 `10`
- **`--secret-key-file` / `-V`**：指定密钥文件或可执行脚本
  - 普通文本：内容即密钥
  - 可执行文件：执行输出作为密钥
  - 支持 `~` 展开
- **`--dry-run`**：新增 dry-run 模式，连接远程但不执行命令和文件操作
  - 输出每台主机每个 step 的解析后参数
  - 跳过执行前确认，不输出执行摘要
- **CLI 变量注入**：新增 `-D/--define` 参数
  - 格式：`-D key=value`，可重复
  - 优先级最高，覆盖 `global_vars` / 组 `vars` / 主机 `Extra`
  - key 和 value 均做 trim
  - 拒绝覆盖保留变量（`date` / `execId`）
  - 用法示例：`taskssh revert web_server -D revert_id=20260925_0K8X7A2B`
- **主机查找增强**：支持按组名、`组名/主机名`、组内主机名查找
  - 组名精确匹配：展开组内所有主机
  - `组名/主机名` 格式：精确到单台主机，如 `web_server/web_1`
  - 组内主机名唯一匹配：如 `web_1`，自动定位到所属组
  - 多组同名主机时报错，并提示使用 `组名/主机名` 格式
  - 独立主机名或 IP 仍按原逻辑处理  

### Changed

- **输出**：单 step 任务不再输出 `[STEP 1/1]` 行
- **`push`**：`-f` 的说明从「文件」改为「文件或目录」
- **`encrypt`**：交互输入密钥两次（输入 + 确认）
- **`decrypt`**：交互输入密钥一次
- **`fetch`**：默认覆盖本地文件，不再需要 `-F`
- **`script -R/--remove`**：短选项从 `-R` 改为 `-r`

### Performance

- **SFTP 会话复用**：每个 `Client` 持有复用的 SFTP 会话，避免每次文件操作都创建新 channel
  - 影响 `push` / `fetch` 目录场景，一次任务期间的 SFTP channel 创建次数从数百次降为 1 次
  - 跨机房网络下收益更明显

### Fix

- 修复变量加载顺序错误，导致全局参数意外覆盖了服务器参数的错误
- 修复FetchAvtion本地路径解析异常导致下载失败的问题

---

## [2.0.1] - 2026-09-23

### Add

- **并发执行**：新增`-c/--concurrency` 并发执行参数，默认值 1 表示串行执行，同时优化了并行时的日志输出

### Changed

- **优化信息输出**：当任务只有一个步骤时，不再输出步骤进度信息

### Fix

- 

---

## [2.0.0] - 2026-09-22

> **重大版本更新：Java → Go 重写。**
>
> 本项目从 Java + JSch 完全重写为 Go，目标是单文件分发、跨平台编译、更好的性能和可维护性。
> Java 版已迁移至 `taskssh-java-final` 分支，除重大 BUG 外不再更新。
> 主分支现为 Go 版，将持续演进。

### ⚠️ Breaking Changes

- **运行环境变更**：不再需要 JDK，改为单文件二进制
- **密文格式不兼容**：加密算法从 Jasypt（PBEWithMD5AndDES）改为 AES-256-GCM
  - Java 版生成的密文无法在 Go 版解密
  - 迁移时需用 Go 版重新加密密码：`taskssh encrypt "密码"`
- **配置文件兼容**：`inventory.yaml` 结构完全不变，用户无需修改清单

### Add

- **单文件分发**：编译产出单个二进制，无需任何运行时
  - Windows：`taskssh.exe`
  - Linux：`taskssh`
  - macOS：`taskssh`

### Changed

- **异常处理**：从 Java 异常体系改为 Go 的 `error` 返回值
- **日志输出**：格式对齐 Java 版
- **上传下载输出**：输出完整路径（本地绝对路径 + 远程完整路径）
- **包结构**：按 Go 惯例重新组织

### Removed

- 移除 Java 版全部代码（迁移至 `taskssh-java-final` 分支）

### Performance

| 维度 | Java 版 | Go 版 |
|---|---|---|
| 启动时间 | ~500ms | ~10ms |
| 内存占用 | ~50MB | ~5MB |
| 二进制大小 | JAR + JDK (~200MB) | 单文件 (~6MB) |
| 部署 | 需装 JDK | 拷贝即用 |

### Migration Guide

从 Java 版迁移到 Go 版：

1. **下载 Go 版**：下载对应平台的 `taskssh` / `taskssh.exe`
2. **保留清单文件**：`inventory.yaml` 结构不变，可直接使用
3. **重新加密密码**：`taskssh encrypt "你的密码"`，把新密文填回清单
4. **验证**：`taskssh command prod -e "date" -y`

## [1.3.1] - 2026-09-22

> **这是 Java 版 TaskSSH 的最后一个稳定版本。**
>
> 后续开发将转向 Go 语言重写，目标是单文件分发、跨平台编译、更好的并发能力。
> Java 版不再新增功能，仅做必要的问题修复。Go 版将继承全部功能并持续演进。

### Fixed

- **跨组主机名重复导致静默覆盖**：不同服务器组中存在同名主机时，后解析的组会覆盖先解析的组，导致部分主机被遗漏
  - 现在主机标识改为 `组名/主机名`（如 `prod/web_1`），支持跨组同名
  - 独立主机（不在任何组内）仍使用原始主机名作为标识

### Changed

- **主机展示格式**：目标主机列表和日志中的主机名，从 `web_1` 变为 `prod/web_1`，语义更明确

### Notes

- 版本号 `1.3.1` 为 Java 版收官版本
- Go 版仓库将独立维护，功能对齐后发布
- 现有 `inventory.yaml` 清单结构不变，Go 版完全兼容
- 密码密文格式不兼容，迁移时需用 Go 版重新加密

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

## [1.0.0] - 2026-09-10

### 初始版本

- 基于 JSch 的轻量级 SSH 运维工具
- 支持 `push` / `command` / `deploy`
- 支持 YAML 清单文件定义服务器组与主机
- 支持变量替换、密码加密、执行摘要

---