[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://go.dev)
[![Version](https://img.shields.io/github/v/release/melvek/TaskSSH?include_prereleases)](https://github.com/melvek/TaskSSH/releases/latest)
![Supports](https://img.shields.io/badge/Supports-Windows,%20Linux,%20macOS-orange)
[![LICENSE](https://img.shields.io/github/license/melvek/TaskSSH)](LICENSE)

![logo](docs/assets/logo.svg)

TaskSSH 是一个单文件、无依赖的 SSH 批量运维工具。你只需要写一个 `inventory.yaml`，定义服务器组、认证信息和任务步骤，就能一次对多台服务器执行命令、上传文件、下载文件。

如果你厌倦了用 Shell 循环 + `sshpass` 管理服务器，又觉得 Ansible 对小型场景太重，TaskSSH 可能正适合你。

## 特性

- **单文件、无依赖**：下载即用，不需要 Python、Agent 或复杂运行环境。
- **YAML 声明式配置**：服务器、变量、任务步骤全部写在 `inventory.yaml`。
- **Action + Task 模型**：`command`、`push`、`pull` 三种原子操作自由组合。
- **密码加密存储**：使用 AES-256-GCM 加密密码，避免清单文件中出现明文。
- **跨平台**：支持 Windows、Linux、macOS。
- **并发执行**：可控制并发数，默认串行，适合不同风险偏好的场景。
- **执行前确认**：默认执行前会确认，避免误操作；可用 `-y` 跳过。
- **Shell Tab 补全**：支持 Bash / PowerShell 补全，提升日常使用效率。

---

## 支持平台

| 平台 | 架构 |
|---|---|
| Windows | amd64 |
| Linux | amd64 |
| Linux | arm64 |
| macOS | amd64 (Intel) |
| macOS | arm64 (Apple Silicon) |

---

## 安装

### 下载二进制

前往 [Releases](https://github.com/melvek/TaskSSH/releases) 页面，下载对应平台的压缩包：

- Windows amd64
- Linux amd64 / arm64

解压后得到：

- `taskssh` / `taskssh.exe`
- `inventory.yaml` 示例
- `README.md`

Linux / macOS 下建议添加执行权限：

```bash
chmod +x taskssh
```

Windows 下直接运行 `taskssh.exe` 即可

---

## 快速开始

### 1. 创建`inventory.yaml`清单文件 

```yaml
global_vars:
  username: deploy
  password: 2dO7ObeRBjqyuKkMpV6Xkg==

servers:
  prod-trans:
    vars:
      service_path: /opt/trans/
    hosts:
      prod_trans_1: 192.168.1.1
      prod_trans_2:
        host: 192.168.1.2
        port: 2222
        username: root
        password: 2dO7ObeRBjqyuKkMpV6Xkg==

tasks:
  release:
    steps:
      - name: "上传新版本"
        action: push
        with:
          file: "./dist/${app_name}-${version}.jar"
          dest: "${service_path}"
          force: true
          backup: true
      - name: "重启服务"
        action: command
        with:
          command: "systemctl restart ${app_name}"
```

### 2. 加密密码（可选）

明文密码存在安全风险。TaskSSH 使用 AES-256-GCM 加密，运行 `taskssh encrypt` 生成密文，填入 `inventory.yaml`。

```bash
taskssh encrypt "your-password"
# 输出：xxxxx
```

把输出的密文填入 `inventory.yaml` 的 `password` 字段。

### 3. 执行任务

```bash
taskssh release prod-trans
```

TaskSSH 会读取 `inventory.yaml`，对 `prod-trans` 组下的所有主机依次执行 `release` 任务。

---

## 清单文件 `inventory.yaml`

TaskSSH 使用一个 YAML 文件描述所有内容。

### `global_vars`

全局默认变量，会被服务器组和主机继承。

```yaml
global_vars:
  username: deploy
  password: "encrypted-password"
  app_name: myapp
```

### `servers`

服务器组定义。每个组可以有自己的 `vars`，组内所有主机会合并这些变量。

```yaml
servers:
  prod:
    vars:
      service_path: /opt/myapp/
    hosts:
      web1: 192.168.1.10
      web2:
        host: 192.168.1.11
        port: 2222
        username: root
        password: "encrypted-password"
```

主机支持两种写法：

- 简写：`主机名: IP`
- 完整写法：包含 `host`、`port`、`username`、`password` 等字段

除 `host`、`port`、`username`、`password` 外的字段会进入 `extraFields`，可用于变量替换。

### `tasks`

任务由若干步骤组成，步骤按顺序执行。

```yaml
tasks:
  release:
    steps:
      - name: 上传新版本
        action: push
        with:
          file: "./dist/${app_name}-${version}.jar"
          dest: "${service_path}"
          force: true
          backup: true

      - name: 重启服务
        action: command
        with:
          command: "systemctl restart ${app_name}"
```

### Action 类型

| Action | 说明 | 常用字段 |
|---|---|---|
| `command` | 在远程主机执行命令 | `command` |
| `push` | 上传本地文件到远程主机 | `file`、`dest`、`force`、`backup` |
| `fetch` | 从远程主机下载文件 | 参考项目示例 |

---

## 变量替换

TaskSSH 支持在 `inventory.yaml` 中使用 `${变量名}` 进行替换。

变量来源包括：

- `global_vars`
- `servers.<组名>.vars`
- 主机自身的 `vars`
- 执行时传入的变量

示例：

```yaml
global_vars:
  app_name: myapp

servers:
  prod:
    vars:
      service_path: /opt/myapp/
    hosts:
      web1: 192.168.1.10

tasks:
  restart:
    steps:
      - name: 重启服务
        action: command
        with:
          command: "systemctl restart ${app_name}"
```

---

### 全局选项

| 选项   | 长选项           | 参数       | 说明                        |
|------|---------------|----------|---------------------------|
| `-i` | `--inventory` | `file`   | 清单文件（默认 `inventory.yaml`） |
| `-l` | `--list`      |          | 仅显示服务器列表                  |
| `-P` | `--port`      | `int`    | 覆盖端口                      |
| `-u` | `--user`      | `string` | 覆盖用户名                     |
| `-p` | `--password`  | `string` | 覆盖密码（不推荐）                 |
| `-c` | `--concurrency` | `int`  | 并发数，默认 `1`（串行）             |
| `-y` | `--yes`       |          | 跳过执行前确认                   |
| `-v` | `--version`   |          | 显示版本                      |
| `-h` | `--help`      |          | 显示帮助（列出所有 Action）         |

### 并发执行

`-c/--concurrency` 控制并发连接数，默认 `1`（串行）。

```bash
# 串行（默认），输出实时、有序
taskssh command prod -e "date"

# 10 台并发
taskssh command prod -e "date" -c 10

# 50 台并发
taskssh release prod -c 50 -y
```

### 工具命令

| 命令        | 说明       |
|-----------|----------|
| `encrypt` | 加密字符串    |
| `decrypt` | 解密字符串    |

```bash
# 加密
taskssh encrypt "your-password"

# 解密
taskssh decrypt "xxxxx"
```

---

## 基础任务

基础任务包括 `command`、`push`、`fetch` 三类，可满足常规运维操作。

基础任务可独立使用，也可作为 `tasks` 中 `step` 的 `action` 进行自由组合，来完成复杂的批量运维任务。

内置任务由工具自带，如果在 `inventory.yaml` 中的 `tasks` 中定义同名任务，将会覆盖内置任务。

| 任务        | 说明          |
|-----------|-------------|
| `command` | 执行远程命令      |
| `push`    | 上传文件到远程     |
| `fetch`   | 从远程下载文件     |

### command — 执行远程命令

| 选项   | 长选项         | 参数       | 说明                           |
|------|-------------|----------|------------------------------|
| `-e` | `--execute` | `string` | 要执行的命令（也可从清单 `command` 字段读取） |

示例：

```bash
taskssh command web-server-01 -e "ls -la /opt"
```

### push — 上传文件

| 选项   | 长选项        | 参数               | 说明                              |
|------|------------|------------------|---------------------------------|
| `-f` | `--file`   | `file or folder` | 本地文件或文件夹                        |
| `-d` | `--dest`   | `path`           | 远程目标路径（也可从清单 `service_path` 读取） |
| `-F` | `--force`  |                  | 覆盖已存在的远程文件                      |
| `-B` | `--backup` |                  | 覆盖前备份原文件                        |

示例：

```bash
taskssh push app-server -f app.jar -d /opt/app/
taskssh push prod-trans -f app.jar -F -B
```

说明：

1. `-d` 以 `/` 结尾视为目录，最终路径为 `目录 + 本地文件名`
2. 目标已存在且为目录，上传到该目录内
3. 目标已存在且为文件：
    - 默认报错
    - `-F` 直接覆盖
    - `-F -B` 备份原文件（追加时间戳）后覆盖
4. 若远程目录不存在时会报错，程序不会自动创建远程目录，若有需要，可提前使用 `command` 命令创建远程目录

### fetch — 下载文件

| 选项   | 长选项        | 参数     | 说明                |
|------|------------|--------|-------------------|
| `-f` | `--file`   | `path` | 远程文件路径            |
| `-d` | `--dest`   | `path` | 本地目标路径            |
| `-F` | `--force`  |        | 覆盖本地已存在文件         |
| `-B` | `--backup` |        | 覆盖前备份本地原文件        |

示例：

```bash
# 下载到目录（按主机标识分组）
taskssh fetch prod -f /var/log/app.log -d ./logs/

# 下载到指定文件
taskssh fetch prod_web_1 -f /var/log/app.log -d ./app.log

# 覆盖并备份
taskssh fetch prod -f /var/log/app.log -d ./logs/ -F -B
```

说明：

1. `-d` 以 `/` 结尾视为目录，按主机标识分组：`dest/<主机标识>/<文件名>`
2. `-d` 是文件路径，直接下载到该文件
3. 本地文件已存在：
    - 默认报错
    - `-F` 直接覆盖
    - `-F -B` 备份本地原文件（追加时间戳）后覆盖
4. 本地目录不存在时自动创建
5. 不支持远程目录下载，需先在远程打包

---

## 定义任务流程

在 `inventory.yaml` 的 `tasks` 段中定义。每个任务包含若干 `steps`，按声明顺序执行。

```yaml
tasks:
  release:
    steps:
      - name: "上传新版本"
        action: push
        delay: 5
        with:
          file: "./dist/${app_name}-${version}.jar"
          dest: "${service_path}"
      - name: "重启服务"
        action: command
        with:
          command: "systemctl restart ${app_name}"
```

字段说明：

| 字段       | 说明                               |
|----------|----------------------------------|
| `name`   | 步骤名，仅用于日志                        |
| `action` | 动作类型，对应 `taskssh -h` 中列出的 Action |
| `with`   | 动作参数，值支持 `${var}` 变量替换           |
| `delay`  | 执行后等待秒数，默认 `0`                   |

执行：

```bash
taskssh release prod-trans -i inventory.yaml -y
```

---

## 变量替换

清单中的任意变量均可在路径、命令、参数中通过 `${key}` 引用。

### 变量来源与优先级

按优先级从低到高：

| 层级  | 来源                     | 说明          |
|-----|------------------------|-------------|
| 1   | `global_vars`          | 全局默认        |
| 2   | `servers.<group>.vars` | 服务组         |
| 3   | `hosts.<host>.vars`    | 单台服务器       |
| 4   | `CLI` 参数               | 命令行指定，优先级最高 |

---

## 安全建议

- 使用 `taskssh encrypt` 加密密码，不要在 `inventory.yaml` 中保存明文。
- 为批量操作创建专用账号，并遵循最小权限原则。
- 生产环境执行前，先用 `taskssh -l` 确认目标服务器。
- 高风险任务建议保持默认串行执行，确认无误后再提高并发。
- 不要将包含加密密码的 `inventory.yaml` 提交到公开仓库。

---

## 适用场景

TaskSSH 适合需要“对多台服务器做同样操作”的轻量场景：

- 批量上传 JAR / 二进制 / 配置文件
- 批量重启服务、重载配置
- 从多台服务器拉取日志
- 批量执行诊断命令
- 小团队 CI/CD 辅助脚本
- 临时运维，不想引入重型配置管理工具

---

## 与 Ansible、Shell 脚本的对比

| 方案 | 依赖 | 配置方式 | 学习成本 | 适合场景 |
|---|---|---|---|---|
| TaskSSH | 单文件，无依赖 | YAML | 低 | 小到中型批量运维 |
| Ansible | Python 等 | YAML Playbook | 中 | 中到大型配置管理 |
| Shell + sshpass | ssh、sshpass 等 | Shell 脚本 | 低但维护难 | 临时、一次性任务 |

TaskSSH 不试图替代 Ansible。它更专注于“用最简单的方式，安全地批量执行 SSH 操作”。

---

## 更新日志

[CHANGELOG](CHANGELOG.md)
