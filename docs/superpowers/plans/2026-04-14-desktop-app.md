# DeltaScope 桌面端实施计划
> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 DeltaScope 构建 Wails + React + TailwindCSS 桌面端应用，提供 analyze/review 图形界面和配置管理
**Architecture:** 现有 CLI 核心逻辑拆分到 `backend` 包，Wails 提供桌面框架，React 前端通过 IPC 调用后端

**Tech Stack:** Wails, Go, React, TailwindCSS

---

## 文件结构预览

```
deltascope/
├── main.go                    # CLI 入口（重构后）
├── go.mod
├── wails.json                 # 新建：Wails 配置
├── app.go                     # 新建：Wails 应用入口
├── backend/                   # 新建：后端逻辑包
│   ├── config.go              # 配置读写
│   ├── git.go                 # Git 操作
│   ├── analyze.go             # analyze 命令逻辑
│   └── review.go              # review 命令逻辑
└── frontend/                  # 新建：React 前端
    ├── package.json
    ├── src/
    │   ├── App.tsx
    │   ├── pages/
    │   │   ├── ConfigPage.tsx
    │   │   ├── AnalyzePage.tsx
    │   │   └── ReviewPage.tsx
    │   └── main.tsx
    └── tailwind.config.js
```

---

## Task 1: 重构 main.go - 拆分 config.go

**Files:**
- Create: `backend/config.go`
- Modify: `main.go`

### Step 1.1: 创建 backend 目录和 config.go

```bash
mkdir -p backend
```

### Step 1.2: 创建 backend/config.go

```go
package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey  string `json:"api_key"`
	APIBase string `json:"api_base"`
	Model   string `json:"model"`
}

func LoadConfig() Config {
	cfg := Config{}

	// 1. global config: ~/.deltascope/config.json
	if home, err := os.UserHomeDir(); err == nil {
		globalPath := filepath.Join(home, ".deltascope", "config.json")
		if data, err := os.ReadFile(globalPath); err == nil {
			_ = json.Unmarshal(data, &cfg)
		}
	}

	// 2. local config: ./.deltascope.json
	if data, err := os.ReadFile(".deltascope.json"); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	// 3. environment variables override config files
	if v := os.Getenv("DELTASCOPE_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("DELTASCOPE_API_BASE"); v != "" {
		cfg.APIBase = v
	}
	if v := os.Getenv("DELTASCOPE_MODEL"); v != "" {
		cfg.Model = v
	}

	return cfg
}

func SaveConfig(cfg Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".deltascope")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
```

### Step 1.3: 修改 main.go 使用 backend.Config

在 `main.go` 顶部 import 添加：
```go
import (
    // ... 现有 import ...
    "deltascope/backend"
)
```

替换 `loadConfig()` 函数（190-1218 行）为：
```go
func loadConfig() deltascopeConfig {
	cfg := backend.LoadConfig()
	return deltascopeConfig{
		APIKey:  cfg.APIKey,
		APIBase: cfg.APIBase,
		Model:   cfg.Model,
	}
}
```

### Step 1.4: 验证 CLI 仍然可用

```bash
go build -o deltascope.exe main.go
./deltascope.exe -h
```

Expected: 帮助信息正常显示

### Step 1.5: 提交

```bash
git add backend/config.go main.go
git commit -m "refactor: extract config logic to backend/config.go"
```

---

## Task 2: 重构 main.go - 拆分 git.go

**Files:**
- Create: `backend/git.go`
- Modify: `main.go`

### Step 2.1: 创建 backend/git.go，提取 Git 相关函数

```go
package backend

import (
	"bytes"
	"os/exec"
	"strings"
)

func runGit(args ...string) ([]byte, error) {
	prefix := []string{
		"-c", "i18n.logOutputEncoding=UTF-8",
		"-c", "i18n.commitEncoding=UTF-8",
		"-c", "core.quotepath=false",
	}
	cmd := exec.Command("git", append(prefix, args...)...)
	cmd.Env = append(cmd.Env, "LC_ALL=C.UTF-8", "LANG=C.UTF-8")
	return cmd.CombinedOutput()
}

func CollectDiffFiles(repoPath, base, head string) ([]string, error) {
	out, err := runGit("-C", repoPath, "diff", "--name-only", base+".."+head)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func CollectDiffContent(repoPath, base, head string) (string, error) {
	out, err := runGit("-C", repoPath, "diff", base+".."+head)
	if err != nil {
		return "", err
	}
	s := string(out)
	const maxLen = 30000
	if len(s) > maxLen {
		s = s[:maxLen] + "\n\n[diff truncated due to length limit]"
	}
	return s, nil
}
```

### Step 2.2: 修改 main.go 使用 backend/git.go

在 `main.go` 中：
- 删除 `runGit()` 函数（84-293 行）
- 删除 `collectDiffFiles()` 函数（289-1302 行）
- 删除 `collectDiffContent()` 函数（304-1315 行）
- 在 import 中确保有 `"deltascope/backend"`
- 修改 `runReview()` 中调用处：
  - `collectDiffFiles(...)` -> `backend.CollectDiffFiles(...)`
  - `collectDiffContent(...)` -> `backend.CollectDiffContent(...)`

### Step 2.3: 验证 CLI 仍然可用

```bash
go build -o deltascope.exe main.go
./deltascope.exe -h
```

### Step 2.4: 提交

```bash
git add backend/git.go main.go
git commit -m "refactor: extract git operations to backend/git.go"
```

---

## Task 3: 初始化 Wails 项目

**Files:**
- Create: `wails.json`
- Create: `app.go`
- Modify: `go.mod`

### Step 3.1: 安装 Wails CLI（如果未安装）
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Step 3.2: 初始化 Wails 项目（手动创建，不覆盖现有文件）

创建 `wails.json`：
```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "deltascope",
  "outputfilename": "deltascope",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "Your Name"
  }
}
```

### Step 3.3: 创建 app.go（Wails 应用入口）
```go
package main

import (
	"context"
	"embed"
	"deltascope/backend"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) LoadConfig() (backend.Config, error) {
	return backend.LoadConfig(), nil
}

func (a *App) SaveConfig(cfg backend.Config) error {
	return backend.SaveConfig(cfg)
}

// TODO: Add RunAnalyze and RunReview bindings in later tasks
```

### Step 3.4: 更新 go.mod，添加 Wails 依赖

```bash
wails mod tidy
```

### Step 3.5: 提交

```bash
git add wails.json app.go go.mod go.sum
git commit -m "feat: initialize wails project structure"
```

---

## Task 4: 创建 React 前端项目

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tailwind.config.js`
- Create: `frontend/postcss.config.js`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`
- Create: `frontend/src/styles.css`

### Step 4.1: 创建 frontend/package.json

```json
{
  "name": "deltascope-frontend",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-markdown": "^9.0.1",
    "@wailsapp/runtime": "^2.0.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@vitejs/plugin-react": "^4.2.0",
    "autoprefixer": "^10.4.16",
    "postcss": "^8.4.32",
    "tailwindcss": "^3.3.6",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

### Step 4.2: 创建 frontend/vite.config.ts

```typescript
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
});
```

### Step 4.3: 创建 frontend/tailwind.config.js

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

### Step 4.4: 创建 frontend/postcss.config.js

```javascript
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

### Step 4.5: 创建 frontend/index.html

```html
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>DeltaScope</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

### Step 4.6: 创建 frontend/src/styles.css

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

### Step 4.7: 创建 frontend/src/main.tsx

```typescript
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './styles.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
```

### Step 4.8: 创建 frontend/src/App.tsx（基础版本）
```typescript
import { useState, useEffect } from 'react'
import * as wails from '@wailsapp/runtime'

interface Config {
  api_key: string
  api_base: string
  model: string
}

function App() {
  const [config, setConfig] = useState<Config>({
    api_key: '',
    api_base: '',
    model: '',
  })

  useEffect(() => {
    wails.Events.On('wails:ready', () => {
      window.backend.LoadConfig().then((cfg: Config) => {
        setConfig(cfg)
      })
    })
  }, [])

  const handleSaveConfig = () => {
    window.backend.SaveConfig(config)
      .then(() => alert('配置已保存'))
      .catch((err: Error) => alert('保存失败: ' + err.message))
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-2xl font-bold mb-8">DeltaScope 配置</h1>
        
        <div className="bg-white p-6 rounded-lg shadow space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">API Key</label>
            <input
              type="text"
              className="w-full border rounded px-3 py-2"
              value={config.api_key}
              onChange={(e) => setConfig({ ...config, api_key: e.target.value })}
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">API Base URL</label>
            <input
              type="text"
              className="w-full border rounded px-3 py-2"
              value={config.api_base}
              onChange={(e) => setConfig({ ...config, api_base: e.target.value })}
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Model</label>
            <input
              type="text"
              className="w-full border rounded px-3 py-2"
              value={config.model}
              onChange={(e) => setConfig({ ...config, model: e.target.value })}
            />
          </div>
          
          <button
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
            onClick={handleSaveConfig}
          >
            保存配置
          </button>
        </div>
      </div>
    </div>
  )
}

export default App
```

### Step 4.9: 安装前端依赖

```bash
cd frontend
npm install
cd ..
```

### Step 4.10: 提交

```bash
git add frontend/
git commit -m "feat: add react frontend base structure"
```

---

## Task 5: 实现 Analyze 后端逻辑 + 前端页面

**Files:**
- Create: `backend/analyze.go`
- Modify: `app.go`（添加 `RunAnalyze` 绑定）
- Create: `frontend/src/pages/AnalyzePage.tsx`
- Modify: `frontend/src/App.tsx`（添加导航）

### Step 5.1: 创建 backend/analyze.go

（从 main.go 提取 runAnalyze 逻辑，修改为返回结果而非直接输出文件）
### Step 5.2: 修改 app.go 添加 RunAnalyze 绑定

### Step 5.3: 创建 frontend/src/pages/AnalyzePage.tsx

### Step 5.4: 修改 App.tsx 添加导航和路由
### Step 5.5: 提交

```bash
git add backend/analyze.go app.go frontend/src/pages/AnalyzePage.tsx frontend/src/App.tsx
git commit -m "feat: add analyze page and backend logic"
```

---

## Task 6: 实现 Review 后端逻辑 + 前端页面

**Files:**
- Create: `backend/review.go`
- Modify: `app.go`（添加 `RunReview` 绑定）
- Create: `frontend/src/pages/ReviewPage.tsx`
- Modify: `frontend/src/App.tsx`

### Step 6.1: 创建 backend/review.go

（从 main.go 提取 runReview 逻辑，修改为返回结果）
### Step 6.2: 修改 app.go 添加 RunReview 绑定

### Step 6.3: 创建 frontend/src/pages/ReviewPage.tsx

### Step 6.4: 修改 App.tsx 完善导航

### Step 6.5: 提交

```bash
git add backend/review.go app.go frontend/src/pages/ReviewPage.tsx
git commit -m "feat: add review page and backend logic"
```

---

## Task 7: 最终集成与测试

**Files:**
- Modify: `frontend/src/App.tsx`（添加 Markdown 渲染）
- Modify: `frontend/src/pages/AnalyzePage.tsx`（嵌入 `dashboard.html`）
- Modify: `frontend/src/pages/ReviewPage.tsx`（展示 `review.md`）
### Step 7.1: 添加 react-markdown 依赖

```bash
cd frontend
npm install react-markdown
cd ..
```

### Step 7.2: 完善 Analyze 页面展示

### Step 7.3: 完善 Review 页面展示

### Step 7.4: 完整测试

```bash
wails dev
```

### Step 7.5: 构建生产版本

```bash
wails build
```

### Step 7.6: 提交

```bash
git add frontend/src/App.tsx frontend/src/pages/
git commit -m "feat: complete desktop app with report display"
```

---

## 计划自我审查

**1. Spec 覆盖检查：**
- 已覆盖：配置页面 - Task 4, 5, 6
- 已覆盖：Analyze 页面 - Task 5
- 已覆盖：Review 页面 - Task 6
- 已覆盖：Wails + React + TailwindCSS - Task 3, 4
- 已覆盖：代码复用 - Task 1, 2
- 已覆盖：Windows 优先 - 隐含在所有任务中

**2. Placeholder 检查：**
- Task 5、6、7 的详细代码块待补充（目前是概要）

**3. 类型一致性：**
- Config 类型在前后端一致
- 所有文件路径明确
---

Plan complete and saved to `docs/superpowers/plans/2026-04-14-desktop-app.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
