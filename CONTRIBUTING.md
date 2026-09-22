# 贡献指南

感谢你对 TaskSSH 的关注！本文档用于规范代码贡献流程。

### 开发环境

| 要求  | 版本                |
|-----|-------------------|
| Go  | 1.22+             |
| Git | 2.x+              |
| IDE | VS Code（推荐）或 GoLand |

#### 推荐插件

VS Code 用户建议安装：

| 插件 | 说明 |
|---|---|
| **Go** | 官方插件，语法、补全、调试 |
| **GitLens** | Git 增强 |
| **EditorConfig** | 统一缩进、换行 |

Go 插件会自动执行 `gofmt`、`goimports`、`go vet`。首次打开项目时，插件会提示安装工具（`gopls`、`dlv` 等），点 "Install All" 即可。

#### 代码规范

Go 不依赖手册，规范由工具链强制：

| 工具 | 命令 | 作用 |
|---|---|---|
| `gofmt` | `go fmt ./...` | 格式化，保存时自动执行 |
| `go vet` | `go vet ./...` | 静态检查 |
| `golangci-lint` | `golangci-lint run` | 集成多种 linter（推荐） |

参考文档：

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide)

### 贡献流程

1. **提交 Issue**：在 [Issues](https://github.com/melvek/TaskSSH/issues) 说明改动意图
2. **Fork 仓库**：从 `main` 分支创建你的工作分支
3. **编写代码**：遵循下方的代码与注释规约
4. **本地验证**：

   ```bash
   go mod tidy
   go build ./...
   go test ./...
   go vet ./...
   ```

5. **提交 PR**：向 `main` 分支发起 Pull Request

### 代码规约

#### 强制规则

**1. 格式化**

所有代码必须经过 `gofmt`。缩进用 **Tab**，不用空格。

```bash
go fmt ./...
```

**2. 命名**

| 类型 | 规则 | 例子 |
|---|---|---|
| 包名 | 小写，单词，无下划线 | `config`、`ssh` |
| 导出标识符 | 大写开头 | `Connect`、`Upload` |
| 未导出标识符 | 小写开头 | `buildAuthMethods` |
| 缩写词 | 全大写或全小写 | `HTTPClient`、`userID` |

**注意**：`userID` 不是 `userId`，`HTTPClient` 不是 `HttpClient`。

**3. 错误处理**

错误必须处理，不能忽略。

```go
// ✅ 正确
f, err := os.Open(path)
if err != nil {
    return fmt.Errorf("open %s: %w", path, err)
}
defer f.Close()

// ❌ 错误
f, _ := os.Open(path)
```

**规则**：

- 错误信息小写，不加标点
- 用 `%w` 包装，保留调用链
- 不 panic（除非真的无法恢复）

**4. 导出标识符必须有注释**

```go
// Host 描述主机配置。
type Host struct {
    Host string
}

// Connect 建立 SSH 连接。
func Connect(host *Host) (*Client, error) {
    // ...
}
```

**注释以标识符名开头，用完整句子。**

**5. 包注释**

每个包的 `doc.go` 或主文件顶部要有包注释。

```go
// Package config 提供清单文件解析。
package config
```

#### 推荐规则

**6. 中文注释优先**

与其用"半吊子"英文，不如用中文把问题说清楚。专有名词与关键字保持英文原文即可。

**7. 代码修改时同步修改注释**

尤其是参数、返回值、核心逻辑的改动。

**8. 删除未使用的字段、方法、变量**

Go 编译器会报错未使用的变量和 import，但未使用的函数、字段需要手动清理。

**9. 接口要小**

Go 偏爱小接口，通常 1-3 个方法。

```go
// ✅ 小接口
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

**10. 并发安全**

goroutine 必须有退出条件，共享数据用 `sync` 包或 channel 保护。

```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(it Item) {
        defer wg.Done()
        process(it)
    }(item)
}
wg.Wait()
```

#### 参考规则

**11. 谨慎注释掉代码**

被注释的代码应说明原因。永久不用的代码直接删除（代码仓库保存了历史）。

**12. 注释力求精简**

好的命名和代码结构是自解释的，避免过度注释。

反例：

```go
// put elephant into fridge
put(elephant, fridge)
```

正例：方法名和参数名已经说明意图，无需额外注释。

**13. 特殊标记**

`TODO` 和 `FIXME` 需注明标记人与时间：

```go
// TODO(your-name, 2026/09/22): 支持多主机并行
```

### 项目结构

```
taskssh/
├── main.go                     程序入口
├── cmd/                        CLI 命令层
│   ├── root.go                 根命令与全局 flag
│   ├── command.go              command 子命令
│   ├── push.go                 push 子命令
│   ├── fetch.go                fetch 子命令
│   ├── encrypt.go              encrypt / decrypt
│   ├── dynamic.go              动态任务分发
│   └── run.go                  公共执行逻辑
├── internal/                   内部包（不对外暴露）
│   ├── config/                 清单解析
│   ├── ssh/                    SSH 通信
│   ├── action/                 原子动作
│   ├── task/                   任务执行
│   ├── resolve/                变量替换
│   ├── secret/                 加解密
│   └── utils/                  工具与常量
└── testdata/                   测试数据
```

### 测试要求

- 新增功能必须包含单元测试
- 修改现有功能，测试需同步更新
- 测试文件命名：`xxx_test.go`，与源码同目录
- 测试数据放 `testdata/` 目录

```bash
# 跑所有测试
go test ./...

# 详细输出
go test -v ./...

# 覆盖率
go test -cover ./...
```

### 提交信息规范

使用 Conventional Commits 格式：

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

**Type 类型：**

| 类型         | 说明      |
|------------|---------|
| `feat`     | 新功能     |
| `fix`      | Bug 修复  |
| `docs`     | 文档更新    |
| `style`    | 代码格式调整  |
| `refactor` | 重构      |
| `test`     | 测试相关    |
| `chore`    | 构建/工具相关 |

**示例：**

```
feat(ssh): 支持 keyboard-interactive 认证
fix(push): 修复 Windows 路径分隔符问题
docs(readme): 更新 Go 版安装说明
```

### 构建与发布

```bash
# 本地编译
go build -o taskssh main.go

# 多平台编译
make build-all

# Windows 一键发布
build.bat
```

### 许可协议

贡献的代码将在 [MIT](LICENSE) 许可下发布。提交 PR 即表示你同意此条款。

### 联系方式

如有疑问，请通过 [Issues](https://github.com/melvek/TaskSSH/issues) 联系。