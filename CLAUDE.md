# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 椤圭洰姒傝堪

杩欐槸涓€涓?Git 浠撳簱閫氱敤鐨勭己闄峰垎鏋愪笌鍙樻洿椋庨櫓璇勪及 CLI 宸ュ叿锛屼娇鐢?Go 璇█寮€鍙戙€?
- **浜у搧瀹氫綅**: Git 浠撳簱缂洪櫡鍒嗘瀽涓庡彉鏇撮闄╄瘎浼板伐鍏凤紝鍓嶅悗绔€氱敤锛屼笉渚濊禆璇█鏍?- **鏍稿績鍔熻兘**:
  - `analyze`: 鍒嗘瀽鎸囧畾鏃堕棿鑼冨洿鐨?bug 淇鎻愪氦锛岃緭鍑烘姤鍛?鍥捐〃
  - `review`: 瀵规瘮鍒嗘敮锛屼娇鐢?AI 鍒嗘瀽褰卞搷鑼冨洿銆侀闄╂姤鍛娿€佹祴璇曡鐐?
## 椤圭洰缁撴瀯

```
deltascope/
鈹溾攢鈹€ main.go          # 涓荤▼搴忎唬鐮侊紙鍗曟枃浠跺簲鐢級
鈹溾攢鈹€ go.mod           # Go 妯″潡瀹氫箟
鈹溾攢鈹€ dashboard.html   # 鍐呭祵鐨?HTML 鐪嬫澘妯℃澘锛?/go:embed锛?鈹溾攢鈹€ 瀹氫綅.md          # 浜у搧瀹氫綅鏂囨。
鈹斺攢鈹€ deltascope.exe      # 缂栬瘧鍚庣殑鍙墽琛屾枃浠讹紙Windows锛?```

## 甯哥敤鍛戒护

### 鏋勫缓

```bash
go build -o deltascope.exe main.go
```

### 杩愯

```bash
# 鍒嗘瀽鏈€杩?3 涓湀鐨勭己闄?deltascope analyze --repo . --since 3m --out ./deltascope-reports --charts --json

# 鍒嗘瀽鎸囧畾鏃ユ湡鑼冨洿
deltascope analyze --repo . --from 2026-01-01 --to 2026-03-31 --out ./deltascope-reports

# AI 浠ｇ爜鍙樻洿璇勫
deltascope review --base origin/develop --head HEAD --api-key $API_KEY
```

## 鏍稿績鏋舵瀯

### 涓昏鏁版嵁缁撴瀯

- `Commit`: Git 鎻愪氦淇℃伅
- `DefectIssue`: 缂洪櫡闂锛堟寜浠诲姟鍙峰垎缁勶級
- `analysisSummary`: 鍒嗘瀽鎽樿缁熻

### 鍏抽敭鍑芥暟

- `runAnalyze()`: 鎵ц缂洪櫡鍒嗘瀽鍛戒护
- `runReview()`: 鎵ц AI 浠ｇ爜璇勫鍛戒护
- `collectCommits()`: 鏀堕泦 Git 鎻愪氦鍘嗗彶
- `groupDefects()`: 鎸変换鍔″彿鍒嗙粍缂洪櫡
- `summarizeCommits()`: 鐢熸垚鍒嗘瀽鎽樿
- `callAIReview()`: 璋冪敤 LLM API 杩涜浠ｇ爜璇勫

### 杈撳嚭鏂囦欢

鍒嗘瀽鍛戒护浼氬湪杈撳嚭鐩綍鐢熸垚锛?- `report.md`: 浜哄伐鍙鐨?Markdown 鎶ュ憡
- `report.csv`: CSV 鏍煎紡鏁版嵁锛堝甫 BOM锛孍xcel 鍙嬪ソ锛?- `report.json`: JSON 鏍煎紡鏁版嵁锛堝彲閫夛紝`--json`锛?- `dashboard.html`: 鍙鍖栫湅鏉匡紙鍙€夛紝`--charts`锛?
璇勫鍛戒护鐢熸垚锛?- `review.md`: 浠ｇ爜鍙樻洿褰卞搷鍒嗘瀽鎶ュ憡

## 閰嶇疆

鏀寔涓夌閰嶇疆鏂瑰紡锛堜紭鍏堢骇浠庨珮鍒颁綆锛夛細
1. 鍛戒护琛屽弬鏁?2. 鐜鍙橀噺锛歚DELTASCOPE_API_KEY`, `DELTASCOPE_API_BASE`, `DELTASCOPE_MODEL`
3. 閰嶇疆鏂囦欢锛?   - 鍏ㄥ眬锛歚~/.deltascope/config.json`
   - 鏈湴锛歚./.deltascope.json`

## 寮€鍙戞敞鎰忎簨椤?
- 杩欐槸涓€涓崟鏂囦欢 Go 搴旂敤锛屾墍鏈変唬鐮佸湪 `main.go` 涓?- `dashboard.html` 閫氳繃 `//go:embed` 鍐呭祵鍒板彲鎵ц鏂囦欢涓?- 榛樿浣跨敤 DeepSeek API 杩涜 AI 璇勫锛屽彲閫氳繃 `--api-base` 鍒囨崲
- Git 鍛戒护閫氳繃 `exec.Command` 璋冪敤锛岄渶纭繚 Git 鍦?PATH 涓?