# DeltaScope 妗岄潰绔疄鏂借鍒?
> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 涓?deltascope 鏋勫缓 Wails + React + TailwindCSS 妗岄潰绔簲鐢紝鎻愪緵 analyze/review 鍥惧舰鐣岄潰鍜岄厤缃鐞?
**Architecture:** 鐜版湁 CLI 鏍稿績閫昏緫鎷嗗垎鍒?backend 鍖咃紝Wails 鎻愪緵妗岄潰妗嗘灦锛孯eact 鍓嶇閫氳繃 IPC 璋冪敤鍚庣

**Tech Stack:** Wails, Go, React, TailwindCSS

---

## 鏂囦欢缁撴瀯棰勮

```
deltascope/
鈹溾攢鈹€ main.go                    # CLI 鍏ュ彛锛堥噸鏋勫悗锛?鈹溾攢鈹€ go.mod
鈹溾攢鈹€ wails.json                 # 鏂板缓锛歐ails 閰嶇疆
鈹溾攢鈹€ app.go                     # 鏂板缓锛歐ails 搴旂敤鍏ュ彛
鈹?鈹溾攢鈹€ backend/                   # 鏂板缓锛氬悗绔€昏緫鍖?鈹?  鈹溾攢鈹€ config.go              # 閰嶇疆璇诲啓
鈹?  鈹溾攢鈹€ git.go                 # Git 鎿嶄綔
鈹?  鈹溾攢鈹€ analyze.go             # analyze 鍛戒护閫昏緫
鈹?  鈹斺攢鈹€ review.go              # review 鍛戒护閫昏緫
鈹?鈹斺攢鈹€ frontend/                  # 鏂板缓锛歊eact 鍓嶇
    鈹溾攢鈹€ package.json
    鈹溾攢鈹€ src/
    鈹?  鈹溾攢鈹€ App.tsx
    鈹?  鈹溾攢鈹€ pages/
    鈹?  鈹?  鈹溾攢鈹€ ConfigPage.tsx
    鈹?  鈹?  鈹溾攢鈹€ AnalyzePage.tsx
    鈹?  鈹?  鈹斺攢鈹€ ReviewPage.tsx
    鈹?  鈹斺攢鈹€ main.tsx
    鈹斺攢鈹€ tailwind.config.js
```

---

## Task 1: 閲嶆瀯 main.go - 鎷嗗垎 config.go

**Files:**
- Create: `backend/config.go`
- Modify: `main.go`

### Step 1.1: 鍒涘缓 backend 鐩綍鍜?config.go

```bash
mkdir -p backend
```

### Step 1.2: 鍒涘缓 backend/config.go

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

### Step 1.3: 淇敼 main.go 浣跨敤 backend.Config

鍦?`main.go` 椤堕儴 import 娣诲姞锛?```go
import (
    // ... 鐜版湁 import ...
    "deltascope/backend"
)
```

鏇挎崲 `loadConfig()` 鍑芥暟锛?190-1218 琛岋級涓猴細
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

### Step 1.4: 楠岃瘉 CLI 浠嶇劧鍙敤

```bash
go build -o deltascope.exe main.go
./deltascope.exe -h
```

Expected: 甯姪淇℃伅姝ｅ父鏄剧ず

### Step 1.5: 鎻愪氦

```bash
git add backend/config.go main.go
git commit -m "refactor: extract config logic to backend/config.go"
```

---

## Task 2: 閲嶆瀯 main.go - 鎷嗗垎 git.go

**Files:**
- Create: `backend/git.go`
- Modify: `main.go`

### Step 2.1: 鍒涘缓 backend/git.go锛屾彁鍙?Git 鐩稿叧鍑芥暟

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

### Step 2.2: 淇敼 main.go 浣跨敤 backend/git.go

鍦?`main.go` 涓細
- 鍒犻櫎 `runGit()` 鍑芥暟锛?84-293 琛岋級
- 鍒犻櫎 `collectDiffFiles()` 鍑芥暟锛?289-1302 琛岋級
- 鍒犻櫎 `collectDiffContent()` 鍑芥暟锛?304-1315 琛岋級
- 鍦?import 涓‘淇濇湁 `"deltascope/backend"`
- 淇敼 `runReview()` 涓皟鐢ㄥ锛?  - `collectDiffFiles(...)` 鈫?`backend.CollectDiffFiles(...)`
  - `collectDiffContent(...)` 鈫?`backend.CollectDiffContent(...)`

### Step 2.3: 楠岃瘉 CLI 浠嶇劧鍙敤

```bash
go build -o deltascope.exe main.go
./deltascope.exe -h
```

### Step 2.4: 鎻愪氦

```bash
git add backend/git.go main.go
git commit -m "refactor: extract git operations to backend/git.go"
```

---

## Task 3: 鍒濆鍖?Wails 椤圭洰

**Files:**
- Create: `wails.json`
- Create: `app.go`
- Modify: `go.mod`

### Step 3.1: 瀹夎 Wails CLI锛堝鏋滄湭瀹夎锛?
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Step 3.2: 鍒濆鍖?Wails 椤圭洰锛堟墜鍔ㄥ垱寤猴紝涓嶈鐩栫幇鏈夋枃浠讹級

鍒涘缓 `wails.json`锛?```json
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

### Step 3.3: 鍒涘缓 app.go锛圵ails 搴旂敤鍏ュ彛锛?
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

### Step 3.4: 鏇存柊 go.mod锛屾坊鍔?Wails 渚濊禆

```bash
wails mod tidy
```

### Step 3.5: 鎻愪氦

```bash
git add wails.json app.go go.mod go.sum
git commit -m "feat: initialize wails project structure"
```

---

## Task 4: 鍒涘缓 React 鍓嶇椤圭洰

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tailwind.config.js`
- Create: `frontend/postcss.config.js`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`
- Create: `frontend/src/styles.css`

### Step 4.1: 鍒涘缓 frontend/package.json

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

### Step 4.2: 鍒涘缓 frontend/vite.config.ts

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

### Step 4.3: 鍒涘缓 frontend/tailwind.config.js

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

### Step 4.4: 鍒涘缓 frontend/postcss.config.js

```javascript
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

### Step 4.5: 鍒涘缓 frontend/index.html

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

### Step 4.6: 鍒涘缓 frontend/src/styles.css

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

### Step 4.7: 鍒涘缓 frontend/src/main.tsx

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

### Step 4.8: 鍒涘缓 frontend/src/App.tsx锛堝熀纭€鐗堟湰锛?
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
      .then(() => alert('閰嶇疆宸蹭繚瀛?))
      .catch((err: Error) => alert('淇濆瓨澶辫触: ' + err.message))
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-2xl font-bold mb-8">DeltaScope 閰嶇疆</h1>
        
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
            淇濆瓨閰嶇疆
          </button>
        </div>
      </div>
    </div>
  )
}

export default App
```

### Step 4.9: 瀹夎鍓嶇渚濊禆

```bash
cd frontend
npm install
cd ..
```

### Step 4.10: 鎻愪氦

```bash
git add frontend/
git commit -m "feat: add react frontend base structure"
```

---

## Task 5: 瀹炵幇 Analyze 鍚庣閫昏緫 + 鍓嶇椤甸潰

**Files:**
- Create: `backend/analyze.go`
- Modify: `app.go`锛堟坊鍔?RunAnalyze 缁戝畾锛?- Create: `frontend/src/pages/AnalyzePage.tsx`
- Modify: `frontend/src/App.tsx`锛堟坊鍔犲鑸級

### Step 5.1: 鍒涘缓 backend/analyze.go

锛堜粠 main.go 鎻愬彇 runAnalyze 閫昏緫锛屼慨鏀逛负杩斿洖缁撴灉鑰岄潪鐩存帴杈撳嚭鏂囦欢锛?
### Step 5.2: 淇敼 app.go 娣诲姞 RunAnalyze 缁戝畾

### Step 5.3: 鍒涘缓 frontend/src/pages/AnalyzePage.tsx

### Step 5.4: 淇敼 App.tsx 娣诲姞瀵艰埅鍜岃矾鐢?
### Step 5.5: 鎻愪氦

```bash
git add backend/analyze.go app.go frontend/src/pages/AnalyzePage.tsx frontend/src/App.tsx
git commit -m "feat: add analyze page and backend logic"
```

---

## Task 6: 瀹炵幇 Review 鍚庣閫昏緫 + 鍓嶇椤甸潰

**Files:**
- Create: `backend/review.go`
- Modify: `app.go`锛堟坊鍔?RunReview 缁戝畾锛?- Create: `frontend/src/pages/ReviewPage.tsx`
- Modify: `frontend/src/App.tsx`

### Step 6.1: 鍒涘缓 backend/review.go

锛堜粠 main.go 鎻愬彇 runReview 閫昏緫锛屼慨鏀逛负杩斿洖缁撴灉锛?
### Step 6.2: 淇敼 app.go 娣诲姞 RunReview 缁戝畾

### Step 6.3: 鍒涘缓 frontend/src/pages/ReviewPage.tsx

### Step 6.4: 淇敼 App.tsx 瀹屽杽瀵艰埅

### Step 6.5: 鎻愪氦

```bash
git add backend/review.go app.go frontend/src/pages/ReviewPage.tsx
git commit -m "feat: add review page and backend logic"
```

---

## Task 7: 鏈€缁堥泦鎴愪笌娴嬭瘯

**Files:**
- Modify: `frontend/src/App.tsx`锛堟坊鍔?Markdown 娓叉煋锛?- Modify: `frontend/src/pages/AnalyzePage.tsx`锛堝祵鍏?dashboard.html锛?- Modify: `frontend/src/pages/ReviewPage.tsx`锛堝睍绀?review.md锛?
### Step 7.1: 娣诲姞 react-markdown 渚濊禆

```bash
cd frontend
npm install react-markdown
cd ..
```

### Step 7.2: 瀹屽杽 Analyze 椤甸潰灞曠ず

### Step 7.3: 瀹屽杽 Review 椤甸潰灞曠ず

### Step 7.4: 瀹屾暣娴嬭瘯

```bash
wails dev
```

### Step 7.5: 鏋勫缓鐢熶骇鐗堟湰

```bash
wails build
```

### Step 7.6: 鎻愪氦

```bash
git add frontend/src/App.tsx frontend/src/pages/
git commit -m "feat: complete desktop app with report display"
```

---

## 璁″垝鑷垜瀹℃煡

**1. Spec 瑕嗙洊妫€鏌ワ細**
- 鉁?閰嶇疆椤甸潰 - Task 4, 5, 6
- 鉁?Analyze 椤甸潰 - Task 5
- 鉁?Review 椤甸潰 - Task 6
- 鉁?Wails + React + TailwindCSS - Task 3, 4
- 鉁?浠ｇ爜澶嶇敤 - Task 1, 2
- 鉁?Windows 浼樺厛 - 闅愬惈鍦ㄦ墍鏈変换鍔′腑

**2. Placeholder 妫€鏌ワ細**
- 鈿狅笍 Task 5, 6, 7 鐨勮缁嗕唬鐮佸潡寰呰ˉ鍏咃紙鐩墠鏄瑕侊級

**3. 绫诲瀷涓€鑷存€э細**
- 鉁?Config 绫诲瀷鍦ㄥ墠鍚庣涓€鑷?- 鉁?鎵€鏈夋枃浠惰矾寰勬槑纭?
---

Plan complete and saved to `docs/superpowers/plans/2026-04-14-desktop-app.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
