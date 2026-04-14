# DeltaScope

DeltaScope 是一个面向 Git 仓库的缺陷分析与代码变更风险评估工具，使用 Go 开发，当前同时提供 CLI 和 Wails 桌面端。

当前版本聚焦两个核心场景：

- `analyze`：分析一段时间内的缺陷修复提交，输出统计报告和可视化看板
- `review`：对比两个分支或提交的差异，调用 LLM 生成影响范围和测试建议

## 功能概览

### 1. analyze

用于扫描指定时间范围内的 Git 提交，识别缺陷修复类提交，并按任务号聚合输出分析结果。

主要能力：

- 支持相对时间范围，如 `1m`、`3m`、`6m`、`7d`、`30d`
- 支持绝对日期范围 `--from` / `--to`
- 支持识别 hotfix 分支提交
- 输出 Markdown、CSV、JSON 和 HTML 看板

输出文件：

- `report.md`
- `report.csv`
- `report.json`，需显式传 `--json`
- `dashboard.html`，默认生成，可通过 `--charts` 控制

### 2. review

用于比较两个分支或提交之间的 diff，并调用兼容 Chat Completions 接口的 LLM 生成变更影响分析。

主要能力：

- 获取变更文件列表
- 截取并上传 diff 内容
- 生成影响范围清单
- 生成测试场景建议
- 输出 `review.md`

## 启动项目

### 1. 启动桌面端开发模式

首次使用前，先安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

然后在项目根目录启动：

```bash
"$(go env GOPATH)/bin/wails" dev
```

说明：

- 会同时启动 Wails 开发进程和前端 Vite 开发服务器
- 代码修改后会自动热更新
- 如果本机没有安装前端依赖，命令会自动安装

### 2. 构建桌面端

构建可执行文件但跳过平台打包：

```bash
"$(go env GOPATH)/bin/wails" build -nopackage -clean
```

如果只是想验证桌面端是否能编译，这是最稳妥的方式。

### 3. 启动 CLI

构建 CLI：

```bash
go build -o deltascope.exe main.go
```

查看帮助：

```bash
./deltascope.exe --help
```

也可以直接构建完整项目：

```bash
go build ./...
```

## 使用方式

### 桌面端

当前程序行为如下：

- 直接运行 `deltascope` 会启动桌面端
- 运行 `deltascope desktop` 也会启动桌面端
- 运行 `deltascope analyze ...` 或 `deltascope review ...` 会进入 CLI 模式

### CLI 查看帮助

```bash
./deltascope.exe --help
```

### analyze 示例

分析最近 3 个月的缺陷修复提交：

```bash
deltascope analyze --repo . --since 3m --out ./deltascope-reports --charts --json
```

分析指定日期范围：

```bash
deltascope analyze --repo . --from 2026-01-01 --to 2026-03-31 --out ./deltascope-reports
```

常用参数：

- `--repo`：Git 仓库路径，默认 `.` 
- `--since`：相对时间范围
- `--from` / `--to`：绝对时间范围，格式 `YYYY-MM-DD`
- `--out`：输出目录，默认 `./deltascope-reports`
- `--branch`：hotfix 分支匹配模式，默认 `hotfix/*`
- `--prefix`：缺陷修复提交前缀，默认 `fix:`
- `--json`：额外生成 `report.json`
- `--charts`：生成 `dashboard.html`
- `--open`：生成后自动打开 HTML 看板

说明：

- `--since` 与 `--from/--to` 不能同时使用
- 不传时间参数时，默认分析最近 6 个月
- 当前版本中，不带子命令直接运行会启动桌面端；如需 CLI，请显式使用 `analyze`

### review 示例

对比 `origin/develop` 和当前分支：

```bash
deltascope review --base origin/develop --head HEAD --api-key "$DELTASCOPE_API_KEY"
```

指定完整 LLM 配置：

```bash
deltascope review \
  --base origin/develop \
  --head feature/my-change \
  --api-key "$DELTASCOPE_API_KEY" \
  --api-base "$DELTASCOPE_API_BASE" \
  --model "$DELTASCOPE_MODEL" \
  --out ./deltascope-reports
```

常用参数：

- `--repo`：Git 仓库路径，默认 `.`
- `--base`：基线分支或提交，默认 `origin/develop`
- `--head`：目标分支或提交，默认 `HEAD`
- `--out`：输出目录，默认 `./deltascope-reports`
- `--api-key`：LLM API Key
- `--api-base`：LLM API Base URL
- `--model`：模型名

输出文件：

- `review.md`

## 配置

`review` 命令支持以下配置来源，优先级从高到低：

1. 命令行参数
2. 环境变量
3. 本地配置文件 `./.deltascope.json`
4. 全局配置文件 `~/.deltascope/config.json`

支持的环境变量：

- `DELTASCOPE_API_KEY`
- `DELTASCOPE_API_BASE`
- `DELTASCOPE_MODEL`

配置文件示例：

```json
{
  "api_key": "your-api-key",
  "api_base": "https://api.example.com/v1",
  "model": "your-model"
}
```

## 运行要求

- 本机需要安装 Git，并确保 `git` 在 `PATH` 中
- `review` 依赖可访问的 LLM 接口
- `review` 请求的是 Chat Completions 风格接口，程序会自动补全 `/v1/chat/completions`

## 项目结构

```text
.
├── app.go
├── main.go
├── go.mod
├── dashboard.html
├── backend/
│   ├── config.go
│   └── git.go
├── frontend/
├── wails.json
└── docs/
```

## 当前状态

当前仓库已经包含桌面端基础实现：

- Wails 桌面壳
- React + TailwindCSS 前端
- 配置页
- Analyze 页面
- Review 页面

目前 `analyze` / `review` 的核心逻辑仍主要复用现有 `main.go` 路径，后续可以继续按计划拆分到 `backend/` 的独立文件中。
