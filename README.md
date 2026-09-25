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
- **Action + Task 模型**：`command`、`push`、`fetch`、`script` 四种原子操作自由组合。
- **目录传输**：`push` / `fetch` 支持目录，默认递归，可选 zip 打包。
- **密码加密存储**：使用 AES-256-GCM 加密密码，清单中不出现明文。
- **密钥外置**：加密密钥由用户管理，不硬编码在二进制中。
- **跨平台**：支持 Windows、Linux、macOS。
- **并发执行**：可控制并发数，默认串行。
- **Dry-run**：`--dry-run` 显示每台主机每个步骤将执行的内容，不实际执行。

## 支持平台

| 平台 | 架构 |
|---|---|
| Windows | amd64 |
| Linux | amd64 |
| Linux | arm64 |
| macOS | amd64 (Intel) |
| macOS | arm64 (Apple Silicon) |

## 安装

前往 [Releases](https://github.com/melvek/TaskSSH/releases) 页面，下载对应平台的压缩包，解压后得到 `taskssh` / `taskssh.exe`。

Linux / macOS 下添加执行权限：

```bash
chmod +x taskssh
```

## 快速开始

### 1. 创建密钥文件

用于加密清单中的密码。只需创建一次：

```bash
mkdir -p ~/.taskssh
openssl rand -base64 32 > ~/.taskssh/vault-key
chmod 600 ~/.taskssh/vault-key
```

### 2. 加密密码

```bash
taskssh encrypt "your-password"
# 输出：xxxxx
```

### 3. 创建 `inventory.yaml`

```yaml
global_vars:
  username: deploy
  password: "xxxxx"        # 上一步的密文

servers:
  prod:
    vars:
      service_path: /opt/myapp/
    hosts:
      web1: 192.168.1.10
      web2: 192.168.1.11

tasks:
  release:
    steps:
      - name: "上传新版本"
        action: push
        with:
          file: "./dist/app.jar"
          dest: "${service_path}"
          force: true
      - name: "重启服务"
        action: command
        with:
          command: "systemctl restart myapp"
        delay: 5
```

### 4. 执行任务

```bash
taskssh release prod -y
```

## 文档

完整文档见 [document.html](https://taskssh.mestrap.com/document.html)，包含：

- 清单文件详解（`global_vars` / `servers` / `tasks`）
- Action 参考（`command` / `push` / `fetch` / `script`）
- 变量替换规则
- Vault 密钥管理
- 命令行选项
- 认证方式
- 常见问题

## 更新日志

[CHANGELOG](CHANGELOG.md)
