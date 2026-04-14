# DeltaScope

DeltaScope 是一个面向 Git 仓库的缺陷分析与代码变更风险评估 CLI 工具，使用 Go 开发。

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

## 构建

```bash
go build -o deltascope.exe main.go
```

## 使用方式

查看帮助：

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
- 不带子命令直接运行时，默认执行 `analyze`

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
├── main.go
├── go.mod
├── dashboard.html
├── backend/
│   ├── config.go
│   └── git.go
└── docs/
```

## 当前状态

当前仓库实现的是 CLI 版本。桌面版相关内容仍在 `docs/` 中，属于设计和计划文档，不是当前已完成交付的一部分。
