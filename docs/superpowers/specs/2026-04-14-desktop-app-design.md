# DeltaScope 桌面端设计文档
**日期**: 2026-04-14  
**目标**: Wails + React + TailwindCSS 桌面端应用（Windows 优先）

---

## 1. 项目背景与目标

### 背景
- 现有 DeltaScope CLI 工具功能完整，但团队成员中的非技术人员不会使用命令行。
- 报告展示和分享还不够方便。

### 目标用户
- 内部团队约 20 人，包括开发和测试人员。

### 核心目标
- 提供图形界面，让命令行操作转为点击操作。
- 数据存储在用户本地，无需服务端支持。

---

## 2. 技术选型

| 层级 | 技术 | 说明 |
|------|------|------|
| 桌面框架 | Wails | Go 后端 + Web 前端，跨平台，包体积较小 |
| 前端框架 | React | 组件化开发，生态成熟 |
| UI 样式 | TailwindCSS | 快速构建美观界面 |
| 目标平台 | Windows | 优先支持，后续可扩展到 macOS / Linux |

---

## 3. 第一版功能范围

### 3.1 配置页面
- 表单字段：
  - API Key
  - API Base URL
  - Model 名称
- 功能：
  - 保存配置到 `~/.deltascope/config.json`
  - 从配置文件加载已保存的配置
  - 验证配置（可选，测试 API 连通性）

### 3.2 Analyze 页面
- 表单字段：
  - 仓库路径（支持文件夹选择器）
  - 时间范围选择：
    - 相对时间：`1m / 3m / 6m / 7d / 30d / 90d / 180d`
    - 或绝对日期范围（`from/to`）
  - 输出目录
  - 选项：生成 JSON、生成 Charts
- 功能：
  - 点击“运行分析”按钮执行
  - 显示进度（loading spinner）
  - 运行完成后，嵌入展示 `dashboard.html`

### 3.3 Review 页面
- 表单字段：
  - 仓库路径（支持文件夹选择器）
  - Base 分支 / 提交
  - Head 分支 / 提交
  - 输出目录
- 功能：
  - 点击“运行评审”按钮执行
  - 显示进度（loading spinner）
  - 运行完成后，展示 `review.md`（Markdown 渲染）

---

## 4. 架构设计

### 4.1 整体架构

```text
┌────────────────────────────────────────────────────────────┐
│               React + TailwindCSS 前端                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  配置页面     │  │ Analyze 页面 │  │ Review 页面  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└──────────────────────────┬─────────────────────────────────┘
                           │ Wails IPC
┌──────────────────────────▼─────────────────────────────────┐
│                        Go 后端                             │
│  复用现有 CLI 核心逻辑                                      │
│  - backend.Analyze()                                       │
│  - backend.Review()                                        │
│  - backend.LoadConfig() / SaveConfig()                     │
└────────────────────────────────────────────────────────────┘
```

### 4.2 目录结构

```text
deltascope/
├── main.go                    # CLI 入口（保持不变）
├── go.mod
├── wails.json                 # Wails 配置
├── app.go                     # Wails 应用入口
├── backend/                   # 后端逻辑（从 main.go 拆分）
│   ├── analyze.go             # analyze 命令逻辑
│   ├── review.go              # review 命令逻辑
│   ├── config.go              # 配置读写
│   └── git.go                 # Git 操作（从 main.go 提取）
└── frontend/                  # React + TailwindCSS 前端
    ├── package.json
    ├── src/
    │   ├── App.tsx
    │   ├── pages/
    │   │   ├── ConfigPage.tsx
    │   │   ├── AnalyzePage.tsx
    │   │   └── ReviewPage.tsx
    │   └── components/
    │       └── ...
    └── tailwind.config.js
```

### 4.3 代码复用策略

将现有 `main.go` 中的函数拆分到 `backend/` 包：

| 原函数 | 新位置 | 说明 |
|--------|--------|------|
| `runAnalyze()` | `backend.Analyze()` | 保持逻辑不变，改为返回结果而非直接输出 |
| `runReview()` | `backend.Review()` | 同上 |
| `loadConfig()` | `backend.LoadConfig()` | 移到独立文件 |
| `collectCommits()` 等 Git 操作 | `backend/git.go` | 提取为独立模块 |

**注意：**
- CLI 和桌面端共用同一套 `backend/` 逻辑。
- CLI 的 `main.go` 保持可用，不破坏现有功能。

---

## 5. 交互流程

### 5.1 Analyze 流程

```text
用户输入参数
  -> 点击“运行”
  -> 显示 Loading
  -> 调用 backend.Analyze()
  -> 运行完成
  -> 读取生成的 dashboard.html
  -> 嵌入前端展示
```

### 5.2 Review 流程

```text
用户输入参数
  -> 点击“运行”
  -> 显示 Loading
  -> 调用 backend.Review()
  -> 运行完成
  -> 读取生成的 review.md
  -> Markdown 渲染展示
```

---

## 6. Wails IPC 接口定义

```go
// app.go 中定义的绑定方法

type App struct {
    // ...
}

// 配置相关
func (a *App) LoadConfig() (Config, error)
func (a *App) SaveConfig(cfg Config) error

// Analyze 相关
func (a *App) RunAnalyze(params AnalyzeParams) (AnalyzeResult, error)

// Review 相关
func (a *App) RunReview(params ReviewParams) (ReviewResult, error)

// 工具方法
func (a *App) SelectDirectory() (string, error) // 打开文件夹选择器
```

---

## 7. 非目标（第一版不做）

- 用户账号 / 权限系统
- 服务端数据共享
- 报告历史存档与搜索
- 报告对比功能
- 定时自动扫描
- macOS / Linux 支持（后续可扩展）

---

## 8. 成功标准

1. 团队中的开发和测试都能通过图形界面使用 analyze / review 功能。
2. 不破坏现有 CLI 的功能。
3. 代码复用率大于 80%（CLI 和桌面端共用 backend 逻辑）。
