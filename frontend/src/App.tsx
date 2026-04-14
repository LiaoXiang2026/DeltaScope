import { useEffect, useState } from "react";
import ConfigPage from "./pages/ConfigPage";
import AnalyzePage from "./pages/AnalyzePage";
import ReviewPage from "./pages/ReviewPage";
import { loadConfig } from "./lib/wails";
import type { Config } from "./types";

type PageKey = "config" | "analyze" | "review";

const defaultConfig: Config = {
  api_key: "",
  api_base: "",
  model: "",
};

export default function App() {
  const [page, setPage] = useState<PageKey>("config");
  const [config, setConfig] = useState<Config>(defaultConfig);
  const [bootMessage, setBootMessage] = useState("正在连接桌面运行时...");

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

  return (
    <div className="min-h-screen px-6 py-8 text-slate-100 lg:px-10">
      <div className="mx-auto max-w-[1480px]">
        <header className="mb-8 flex flex-col gap-5 rounded-[32px] border border-slate-800 bg-slate-900/60 px-6 py-6 shadow-2xl shadow-slate-950/30 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p className="text-xs uppercase tracking-[0.35em] text-sky-300">DeltaScope Desktop</p>
            <h1 className="mt-3 text-4xl font-semibold tracking-tight text-white">桌面分析工作台</h1>
            <p className="mt-3 max-w-3xl text-sm leading-6 text-slate-400">
              基于 Wails + React 构建的图形界面，用于执行 Git 缺陷分析和 AI 变更评审。
            </p>
          </div>
          <div className="rounded-2xl border border-slate-800 bg-slate-950/60 px-4 py-3 text-sm text-slate-400">
            {bootMessage}
          </div>
        </header>

        <div className="mb-6 flex flex-wrap gap-3">
          <NavButton label="配置" active={page === "config"} onClick={() => setPage("config")} />
          <NavButton label="Analyze" active={page === "analyze"} onClick={() => setPage("analyze")} />
          <NavButton label="Review" active={page === "review"} onClick={() => setPage("review")} />
        </div>

        {page === "config" ? <ConfigPage config={config} onChange={setConfig} /> : null}
        {page === "analyze" ? <AnalyzePage /> : null}
        {page === "review" ? <ReviewPage config={config} /> : null}
      </div>
    </div>
  );
}

interface NavButtonProps {
  label: string;
  active: boolean;
  onClick: () => void;
}

function NavButton({ label, active, onClick }: NavButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={[
        "rounded-2xl px-5 py-3 text-sm font-medium transition",
        active
          ? "bg-white text-slate-950 shadow-lg shadow-white/10"
          : "border border-slate-800 bg-slate-900/70 text-slate-300 hover:border-slate-600 hover:text-white",
      ].join(" ")}
    >
      {label}
    </button>
  );
}
