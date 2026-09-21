[![Language](https://img.shields.io/badge/Language-Java-blue.svg)](https://java.dev)
[![Version](https://img.shields.io/github/v/release/melvek/TaskSSH?include_prereleases)](https://github.com/melvek/TaskSSH/releases/latest)
![Supports](https://img.shields.io/badge/Supports-Windows,%20Linux-orange)
[![LICENSE](https://img.shields.io/github/license/melvek/TaskSSH)](LICENSE)

TaskSSH 是一个基于 JSch 的轻量级运维工具，支持对同一组服务器批量上传、下载文件及批量执行远程命令。

采用「Action + Task」模型：Action 是原子能力（执行命令、上传文件、下载文件），Task 是由若干 Action 组成的有序任务。
通过 `inventory.yaml` 清单文件定义服务器组、主机、认证信息、业务参数和任务，即可一键完成批量部署、文件推送与命令执行。

---

## 环境要求

- JDK 8 或更高版本
- 远程服务器需支持 SSH / SFTP

---

## 安装

从 [GitHub Releases 页面](https://github.com/melvek/TaskSSH/releases) 下载预编译的 JAR。

也可以编译安装：

```bash
git clone --depth 1 https://github.com/melvek/TaskSSH.git
cd TaskSSH
mvn clean package
```

构建完成后，在 `target/` 目录下生成 `taskssh-x.x.x.jar`。

创建别名便于使用：

```bash
alias taskssh='java -jar /path/to/taskssh-x.x.x.jar'
```

---

## 快速开始

### 1. 创建清单文件 `inventory.yaml`

```yaml
global_vars:
  port: 22
  username: deploy
  password: 2dO7ObeRBjqyuKkMpV6Xkg==
  command: "systemctl restart my-app"

servers:
  prod-trans:
    vars:
      service_path: /opt/trans/
      command: "sh /opt/trans/restart.sh"
    hosts:
      prod_trans_1: 192.168.1.1
      prod_trans_2:
        host: 192.168.1.2
        port: 2222
        username: root
        password: 2dO7ObeRBjqyuKkMpV6Xkg==
        remote_path: /data/app/

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

说明：

- `global_vars`：全局默认参数，被服务器组/主机继承
- `servers`：服务器组定义
    - `vars`：该组的变量，合并到组内所有主机
    - `hosts`：主机列表，支持两种写法
        - 简写：`主机名: IP`，使用全局参数
        - 完整：`主机名: { host, port, username, password, ... }`
- `tasks`：用户自定义任务，可覆盖内置任务，也可新增
- 除 `host`、`port`、`username`、`password` 外的字段进入 `extraFields`，可用于变量替换

### 2. 加密密码

明文密码存在安全风险，TaskSSH 使用 Jasypt 加密。运行 `taskssh encrypt` 生成密文，填入 `inventory.yaml`。

```bash
taskssh encrypt "your-password"
# 输出：2dO7ObeRBjqyuKkMpV6Xkg==
```

注意：主密钥 `SEC_KEY` 硬编码在 `EncryptTool` 中。

---

## 命令总览

```
taskssh <task> <server group / hosts...> [options]
```

### 全局选项

| 选项   | 长选项           | 参数       | 说明                        |
|------|---------------|----------|---------------------------|
| `-i` | `--inventory` | `file`   | 清单文件（默认 `inventory.yaml`） |
| `-l` | `--list`      |          | 仅显示服务器列表                  |
| `-P` | `--port`      | `int`    | 覆盖端口                      |
| `-u` | `--user`      | `string` | 覆盖用户名                     |
| `-p` | `--password`  | `string` | 覆盖密码（不推荐）                 |
| `-y` | `--yes`       |          | 跳过执行前确认                   |
| `-v` | `--version`   |          | 显示版本                      |
| `-h` | `--help`      |          | 显示帮助（列出所有 Action）         |

### 工具命令

| 命令        | 说明       |
|-----------|----------|
| `encrypt` | 加密字符串    |
| `decrypt` | 解密字符串    |

```bash
# 加密
taskssh encrypt "your-password"
# 输出：2dO7ObeRBjqyuKkMpV6Xkg==

# 解密
taskssh decrypt "2dO7ObeRBjqyuKkMpV6Xkg=="
# 输出：your-password
```

---

## 认证方式

支持以下认证方式，按优先级尝试：

| 方式    | 配置字段            | 说明                     |
|-------|-----------------|------------------------|
| 公钥认证  | `identity_file` | 私钥文件路径，`~` 自动展开        |
| 私钥口令  | `passphrase`    | 与密码同样加密存储              |
| 密码认证  | `password`      | 加密存储，或运行时终端输入          |

优先级：`publickey > keyboard-interactive > password`

### 配置示例

```yaml
hosts:
  prod_1:
    host: 1.2.3.4
    username: deploy
    identity_file: ~/.ssh/id_rsa
    passphrase: "密文"
    # 公钥失败时回退密码
    # password: "密文"
```

### 终端交互输入

清单未配置密码或口令时，会从终端读取：

```
deploy@1.2.3.4's password: 
```

多台主机共享密码缓存，只弹一次。

---

## 内置任务

内置任务由工具自带，无需在 `inventory.yaml` 中定义。用户若在 `tasks` 中定义同名任务，则覆盖内置任务。

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
taskssh command prod-trans -i inventory.yaml
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

上传行为：

1. `-d` 以 `/` 结尾视为目录，最终路径为 `目录 + 本地文件名`
2. 目标已存在且为目录，上传到该目录内
3. 目标已存在且为文件：
    - 默认报错
    - `-F` 直接覆盖
    - `-F -B` 备份原文件（追加时间戳）后覆盖
4. 父目录不存在时直接报错，不自动创建

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

下载行为：

1. `-d` 以 `/` 结尾视为目录，按主机标识分组：`dest/<主机标识>/<文件名>`
2. `-d` 是文件路径，直接下载到该文件
3. 本地文件已存在：
    - 默认报错
    - `-F` 直接覆盖
    - `-F -B` 备份本地原文件（追加时间戳）后覆盖
4. 本地目录不存在时自动创建
5. 不支持远程目录，需先在远程打包

---

## 自定义任务

在 `inventory.yaml` 的 `tasks` 段中定义。每个任务包含若干 `steps`，按声明顺序执行。

```yaml
tasks:
  release:
    steps:
      - name: "上传新版本"
        action: push
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

### 示例：用自定义任务实现部署

```yaml
tasks:
  deploy:
    steps:
      - name: "上传应用包"
        action: push
        with:
          file: "${package_path}"
          dest: "${service_path}"
          force: true
      - name: "重启服务"
        action: command
        with:
          command: "${command}"
```

```bash
taskssh deploy prod-trans -f app.jar -y
```

---

## 变量替换

清单中的任意变量均可在路径、命令、参数中通过 `${key}` 引用。

### 变量来源与优先级

按优先级从低到高：

| 层级  | 来源                         | 说明          |
|-----|----------------------------|-------------|
| 1   | `global_vars`              | 全局默认        |
| 2   | `servers.<group>.vars`     | 服务组         |
| 3   | `hosts.<host>.extraFields` | 单台服务器       |
| 4   | CLI 参数                     | 命令行指定，优先级最高 |

### 严格模式

未解析的 `${var}` 会直接报错，不会静默保留。这样能避免拼写错误导致命令执行到意外路径。

命令中如需 shell 变量，用不带花括号的写法：

```bash
taskssh command prod -e "echo $HOME"
```

`$HOME` 原样传给 shell，`${HOME}` 会被当作 TaskSSH 变量处理。

### 内置变量

| 变量     | 说明                 |
|--------|--------------------|
| `date` | 当前日期，格式 `yyyyMMdd` |

---

## 更新日志

[CHANGELOG](CHANGELOG.md)
