import { useEffect, useState } from "react";
import ConfigPage from "./pages/ConfigPage";
import AnalyzePage from "./pages/AnalyzePage";
import ReviewPage from "./pages/ReviewPage";
import { loadConfig, runAnalyze, runReview } from "./lib/wails";
import type { AnalyzeParams, AnalyzeResult, Config, ReviewParams, ReviewResult } from "./types";
import { StatusPill } from "./components/workbench";

type PageKey = "config" | "analyze" | "review";

const defaultConfig: Config = {
  api_key: "",
  api_base: "",
  model: "",
};

const initialAnalyzeParams: AnalyzeParams = {
  repo: ".",
  since: "3m",
  from: "",
  to: "",
  out_dir: "./deltascope-reports",
  branch: "hotfix/*",
  prefix: "fix:",
  generate_json: false,
  generate_charts: true,
};

const initialReviewParams: ReviewParams = {
  repo: ".",
  base: "origin/develop",
  head: "HEAD",
  out_dir: "./deltascope-reports",
  api_key: "",
  api_base: "",
  model: "",
};

const pageMeta: Record<PageKey, { title: string; description: string }> = {
  config: {
    title: "连接配置",
    description: "保存默认模型连接，Review 会优先使用这套配置。",
  },
  analyze: {
    title: "Analyze",
    description: "扫描缺陷修复历史，输出报告和 dashboard。",
  },
  review: {
    title: "Review",
    description: "对比代码差异，生成风险判断和评审报告。",
  },
};

export default function App() {
  const [page, setPage] = useState<PageKey>("config");
  const [config, setConfig] = useState<Config>(defaultConfig);
  const [bootMessage, setBootMessage] = useState("正在连接运行时");
  const [analyzeParams, setAnalyzeParams] = useState<AnalyzeParams>(initialAnalyzeParams);
  const [analyzeRunning, setAnalyzeRunning] = useState(false);
  const [analyzeMessage, setAnalyzeMessage] = useState("等待开始");
  const [analyzeResult, setAnalyzeResult] = useState<AnalyzeResult | null>(null);
  const [reviewParams, setReviewParams] = useState<ReviewParams>(initialReviewParams);
  const [reviewRunning, setReviewRunning] = useState(false);
  const [reviewMessage, setReviewMessage] = useState("等待开始");
  const [reviewResult, setReviewResult] = useState<ReviewResult | null>(null);
  const configReady = Boolean(config.api_key && config.api_base && config.model);

  useEffect(() => {
    loadConfig()
      .then((nextConfig) => {
        setConfig(nextConfig);
        setBootMessage("配置已加载");
      })
      .catch((error) => {
        setBootMessage(error instanceof Error ? error.message : "加载配置失败");
      });
  }, []);

  const currentPage = pageMeta[page];

  async function handleAnalyzeRun() {
    setAnalyzeRunning(true);
    setAnalyzeMessage("正在执行 Analyze...");
    try {
      const nextResult = await runAnalyze(analyzeParams);
      setAnalyzeResult(nextResult);
      setAnalyzeMessage("已完成");
    } catch (error) {
      setAnalyzeMessage(error instanceof Error ? error.message : "Analyze 执行失败");
    } finally {
      setAnalyzeRunning(false);
    }
  }

  async function handleReviewRun() {
    setReviewRunning(true);
    setReviewMessage("正在执行 Review...");
    try {
      const nextResult = await runReview({
        ...reviewParams,
        api_key: reviewParams.api_key || config.api_key,
        api_base: reviewParams.api_base || config.api_base,
        model: reviewParams.model || config.model,
      });
      setReviewResult(nextResult);
      setReviewMessage("已完成");
    } catch (error) {
      setReviewMessage(error instanceof Error ? error.message : "Review 执行失败");
    } finally {
      setReviewRunning(false);
    }
  }

  return (
    <div className="app-shell">
      <aside className="app-sidebar">
        <div className="sidebar-brand">
          <span className="sidebar-brand-mark" />
          <div>
            <p className="sidebar-brand-name">DeltaScope</p>
            <p className="sidebar-brand-copy">代码分析桌面工具</p>
          </div>
        </div>

        <nav className="sidebar-nav" aria-label="主导航">
          <button
            type="button"
            onClick={() => setPage("config")}
            className={["sidebar-nav-item", page === "config" ? "sidebar-nav-item--active" : ""].join(" ")}
          >
            配置
          </button>
          <button
            type="button"
            onClick={() => setPage("analyze")}
            className={["sidebar-nav-item", page === "analyze" ? "sidebar-nav-item--active" : ""].join(" ")}
          >
            Analyze
          </button>
          <button
            type="button"
            onClick={() => setPage("review")}
            className={["sidebar-nav-item", page === "review" ? "sidebar-nav-item--active" : ""].join(" ")}
          >
            Review
          </button>
        </nav>
      </aside>

      <div className="app-main">
        <header className="topbar">
          <div className="topbar-heading">
            <h1>{currentPage.title}</h1>
            <p>{currentPage.description}</p>
          </div>

          <div className="topbar-meta">
            <div className="meta-badge">
              <span>运行状态</span>
              <strong>{bootMessage}</strong>
            </div>
            <div className="meta-badge">
              <span>默认模型</span>
              <StatusPill tone={configReady ? "ready" : "warning"}>
                {configReady ? config.model : "未完成配置"}
              </StatusPill>
            </div>
          </div>
        </header>

        <main className="content-area">
          {page === "config" ? <ConfigPage config={config} onChange={setConfig} /> : null}
          {page === "analyze" ? (
            <AnalyzePage
              params={analyzeParams}
              onParamsChange={setAnalyzeParams}
              running={analyzeRunning}
              message={analyzeMessage}
              result={analyzeResult}
              onRun={handleAnalyzeRun}
            />
          ) : null}
          {page === "review" ? (
            <ReviewPage
              config={config}
              params={reviewParams}
              onParamsChange={setReviewParams}
              running={reviewRunning}
              message={reviewMessage}
              result={reviewResult}
              onRun={handleReviewRun}
            />
          ) : null}
        </main>
      </div>
    </div>
  );
}
