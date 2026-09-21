## 贡献指南

感谢你对 TaskSSH 的关注！本文档基于 [Alibaba Java Coding Guidelines](https://alibaba.github.io/p3c/) 制定，用于规范代码贡献流程。

### 开发环境

| 要求    | 版本                |
|-------|-------------------|
| JDK   | 8+                |
| Maven | 3.6+              |
| IDE   | IntelliJ IDEA（推荐） |

#### 安装编码规约插件

IDEA 中安装 `Alibaba Java Coding Guidelines` 插件：

1. `File` → `Settings` → `Plugins` → `Marketplace`
2. 搜索 “Alibaba Java Coding Guidelines”，安装后重启
3. 通过 `Tools` → `阿里编码规约` → `编码规约扫描` 检查代码

### 贡献流程

1. **提交 Issue**：在 [Issues](https://github.com/melvek/TaskSSH/issues) 说明改动意图
2. **Fork 仓库**：从 `main` 分支创建你的工作分支
3. **编写代码**：遵循下方的注释规约
4. **本地验证**：运行 `mvn clean test` 确保通过
5. **提交 PR**：向 `main` 分支发起 Pull Request

### 注释规约

以下规则遵循 Alibaba Java Coding Guidelines 的注释规约部分 。

#### 强制规则

**1. 类、类属性、类方法必须使用 Javadoc**

使用 `/** 内容 */` 格式，不得使用 `// xxx` 方式。Javadoc 能在 IDE 中悬浮提示，也能正确生成文档。

```java
/**
 * 执行远程命令
 */
public class CommandAction implements TaskAction {
```

**2. 抽象方法必须用 Javadoc 注释**

所有抽象方法（包括接口方法）必须说明返回值、参数、异常，以及方法的功能。

```java
/**
 * 对单台主机执行整条任务
 *
 * @param task     任务定义
 * @param hostVars 目标主机变量
 * @throws TaskException 当某个步骤执行失败时抛出
 */
public void executeOnHost(Task task, HostVars hostVars) throws TaskException {
```

**3. 类必须添加创建者和创建日期**

```java
/**
 * @author your-name
 * @date 2026/09/17
 */
public class TaskExecutor {
```

**4. 方法内部注释**

单行注释在被注释语句上方另起一行，使用 `//`。多行注释使用 `/* */`，与代码对齐。

```java
// 解析最终目标路径
String finalPath = resolveTargetPath(sftp, localFile, remoteTarget);
```

**5. 枚举字段必须有注释**

```java
public enum Color {
    /** 黑色 */
    BLACK("\033[0;30m"),
    /** 红色 */
    RED("\033[0;31m");
}
```

#### 推荐规则

**6. 代码修改时同步修改注释**

尤其是参数、返回值、异常、核心逻辑的改动。

**7. 中文注释优先**

与其用“半吊子”英文，不如用中文把问题说清楚。专有名词与关键字保持英文原文即可 。

**8. 删除未使用的字段、方法、变量**

保持代码整洁。

#### 参考规则

**9. 谨慎注释掉代码**

被注释的代码应说明原因。永久不用的代码直接删除（代码仓库保存了历史）。

**10. 注释力求精简**

好的命名和代码结构是自解释的，避免过度注释。

反例：
```java
// put elephant into fridge
put(elephant, fridge);
```

正例：方法名和参数名已经说明意图，无需额外注释。

**11. 特殊标记**

`TODO` 和 `FIXME` 需注明标记人与时间，及时处理 ：

```java
// TODO(your-name, 2026/09/17): 支持 zip 格式解压
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
feat(push): support directory upload with auto-compress
fix(command): handle exit code correctly
docs(readme): add Task usage examples
```

### 许可协议

贡献的代码将在 [MIT](LICENSE) 许可下发布。提交 PR 即表示你同意此条款。

### 联系方式

如有疑问，请通过 [Issues](https://github.com/melvek/TaskSSH/issues) 联系。