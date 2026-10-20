[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://go.dev)
[![Version](https://img.shields.io/github/v/release/melvek/TaskSSH?include_prereleases)](https://github.com/melvek/TaskSSH/releases/latest)
![Supports](https://img.shields.io/badge/Supports-Windows,%20Linux,%20macOS-orange)
[![License](https://img.shields.io/github/license/melvek/TaskSSH)](LICENSE)

![TaskSSH](docs/assets/images/logo.svg)

# TaskSSH

TaskSSH 是一个轻量级、跨平台的 SSH 批量运维工具。它将服务器、变量和任务保存在 `inventory.yaml` 中，通过单个可执行文件对一台或多台主机执行命令、传输文件，或运行可复用的多步骤任务。

TaskSSH 不需要在远程主机安装 Agent，也不要求用户安装 Python。它适合希望用清晰、可审查的配置完成日常批量操作，同时不需要完整自动化平台的场景。

## 设计原则

- **简单上手**：下载单个可执行文件即可使用；目标主机通过位置参数指定，操作内容通过选项明确表达。
- **复用 SSH**：通过 SSH 连接远程主机，不在目标机器安装额外 Agent 或开放服务端口。
- **配置可读、可审查**：以 YAML 管理主机组、变量和任务，让操作流程便于团队理解与代码审查。
- **默认谨慎**：不指定目标主机时直接报错；使用 `--dry-run` 预览任务，或使用 `--confirm` 在执行前确认目标。
- **跨平台运行**：提供 Windows、Linux 和 macOS 的预构建版本。

## 安装

从 [Releases](https://github.com/melvek/TaskSSH/releases/latest) 下载对应平台的压缩包，解压后即可运行 `taskssh` 或 `taskssh.exe`。

Linux / macOS 可按需添加执行权限：

```bash
chmod +x taskssh
```

从源码构建需要 Go 1.27 或更高版本：

```bash
git clone https://github.com/melvek/TaskSSH.git
cd TaskSSH
go build -o taskssh .
```

## 快速开始

### 直接执行

每个命令使用一个主机 pattern 位置参数。多个主机或组可在同一个 pattern 中用逗号组合；在 PowerShell、Bash 或 Zsh 中，建议将包含逗号、空格或 `&`、`!` 等符号的 pattern 用引号括起来。

```bash
# 在单台主机执行命令
taskssh command web1 -e "uptime"

# 在多个主机组执行命令
taskssh command "prod,web_server" -e "date"

# 上传文件
taskssh push prod -f ./dist/app.jar -d /opt/myapp/

# 下载日志
taskssh fetch prod -f /var/log/myapp.log -d ./logs/

# 查看 dry-run 计划，不连接远程主机
taskssh deploy prod --dry-run
```

`-H/--host` 已移除。主机 pattern 语法及筛选示例见[主机 pattern 文档](https://taskssh.mestrap.com/document.html#host-pattern)。

### 使用清单和任务

创建 `inventory.yaml`，定义默认连接信息、主机组和可复用任务：

```yaml
defaults:
  username: deploy

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
      - name: Upload application
        action: push
        with:
          file: ./dist/app.jar
          dest: "{{service_path}}"
      - name: Restart application
        action: command
        with:
          command: "systemctl restart myapp"
```

执行任务前，可先查看将要操作的主机和步骤：

```bash
taskssh release prod --dry-run
```

确认清单中的 SSH 主机密钥已通过可信渠道验证，并已记录到 OpenSSH `known_hosts` 文件。TaskSSH 不会自动信任未知或已变更的主机密钥。

### 管理加密密钥

使用 Vault 密钥加密清单中需要保存的密码：

```bash
taskssh keygen
taskssh encrypt
```

`keygen` 默认将密钥写入 `~/.taskssh/vault-key`。请安全备份密钥；丢失后，使用该密钥加密的内容将无法解密。

## 功能概览

- **任务编排**：组合 `command`、`push`、`fetch`、`script` 操作为有序任务。
- **变量与条件**：支持默认值、组、主机和命令行变量；用 `when` 按条件执行步骤。
- **批量控制**：支持并发连接，以及通过 `serial`、`min_available` 配置滚动批次。
- **文件传输**：支持文件与目录；可选择逐文件传输或通过 `-z` 打包传输。
- **认证方式**：支持 SSH Agent、公钥、keyboard-interactive 和密码认证。
- **安全保护**：加密保存密码；校验 SSH 主机密钥；限制文件下载路径风险。
- **跨平台**：支持 Windows、Linux 和 macOS。

## 文档与帮助

- [完整使用文档](https://taskssh.mestrap.com/document.html)：清单格式、主机 pattern、Action、变量、认证与 CLI 选项。
- [问题反馈与功能建议](https://github.com/melvek/TaskSSH/issues)：提交 Bug 或讨论改进建议。
- 命令行帮助：`taskssh --help` 或 `taskssh <command> --help`。

## 参与贡献

欢迎提交 Bug 报告、文档改进和代码贡献。提交改动前，请阅读[贡献指南](CONTRIBUTING.md)，并运行相关 Go 测试：

```bash
go test ./...
```

## 许可证

TaskSSH 使用 MIT License，详见 [LICENSE](LICENSE)。

## 更新日志

查看 [CHANGELOG.md](CHANGELOG.md) 了解版本变更。
