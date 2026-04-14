import { useState } from "react";
import { runAnalyze, selectDirectory } from "../lib/wails";
import type { AnalyzeParams, AnalyzeResult } from "../types";

const initialParams: AnalyzeParams = {
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

export default function AnalyzePage() {
  const [params, setParams] = useState<AnalyzeParams>(initialParams);
  const [running, setRunning] = useState(false);
  const [message, setMessage] = useState("等待开始分析");
  const [result, setResult] = useState<AnalyzeResult | null>(null);

  async function chooseRepo() {
    const dir = await selectDirectory();
    if (dir) {
      setParams((current) => ({ ...current, repo: dir }));
    }
  }

  async function chooseOutDir() {
    const dir = await selectDirectory();
    if (dir) {
      setParams((current) => ({ ...current, out_dir: dir }));
    }
  }

  async function handleRun() {
    setRunning(true);
    setMessage("正在执行 Analyze，请稍候...");
    try {
      const nextResult = await runAnalyze(params);
      setResult(nextResult);
      setMessage("Analyze 执行完成");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Analyze 执行失败");
    } finally {
      setRunning(false);
    }
  }

  return (
    <section className="grid gap-6 xl:grid-cols-[420px_minmax(0,1fr)]">
      <div className="space-y-4 rounded-3xl border border-slate-800 bg-slate-900/70 p-6">
        <div>
          <h2 className="text-2xl font-semibold text-slate-50">Analyze</h2>
          <p className="mt-2 text-sm text-slate-400">
            运行 Git 缺陷分析，生成 Markdown、CSV、JSON 和可视化看板。
          </p>
        </div>

        <Field label="仓库路径" value={params.repo} onChange={(value) => setParams({ ...params, repo: value })} actionLabel="选择目录" onAction={chooseRepo} />
        <Field label="相对时间" value={params.since} onChange={(value) => setParams({ ...params, since: value })} />
        <div className="grid gap-4 md:grid-cols-2">
          <Field label="开始日期" value={params.from} onChange={(value) => setParams({ ...params, from: value })} />
          <Field label="结束日期" value={params.to} onChange={(value) => setParams({ ...params, to: value })} />
        </div>
        <Field label="输出目录" value={params.out_dir} onChange={(value) => setParams({ ...params, out_dir: value })} actionLabel="选择目录" onAction={chooseOutDir} />
        <Field label="Hotfix 分支模式" value={params.branch} onChange={(value) => setParams({ ...params, branch: value })} />
        <Field label="缺陷前缀" value={params.prefix} onChange={(value) => setParams({ ...params, prefix: value })} />

        <div className="grid gap-3 sm:grid-cols-2">
          <Toggle
            label="生成 JSON"
            checked={params.generate_json}
            onChange={(checked) => setParams({ ...params, generate_json: checked })}
          />
          <Toggle
            label="生成图表"
            checked={params.generate_charts}
            onChange={(checked) => setParams({ ...params, generate_charts: checked })}
          />
        </div>

        <button
          type="button"
          onClick={handleRun}
          disabled={running}
          className="w-full rounded-2xl bg-emerald-400 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-emerald-300 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {running ? "分析中..." : "运行分析"}
        </button>

        <div className="rounded-2xl border border-slate-800 bg-slate-950/80 px-4 py-3 text-sm text-slate-300">
          {message}
        </div>
      </div>

      <div className="space-y-4 rounded-3xl border border-slate-800 bg-slate-900/70 p-4">
        <div className="flex flex-wrap items-center gap-3 px-2 pt-2 text-xs text-slate-400">
          <span>输出目录：{result?.output_dir ?? "-"}</span>
          <span>Markdown：{result?.report_path ?? "-"}</span>
          <span>Dashboard：{result?.dashboard_path ?? "-"}</span>
        </div>

        {result?.dashboard_html ? (
          <iframe
            title="Analyze Dashboard"
            srcDoc={result.dashboard_html}
            className="min-h-[720px] w-full rounded-2xl border border-slate-800 bg-white"
          />
        ) : (
          <div className="flex min-h-[720px] items-center justify-center rounded-2xl border border-dashed border-slate-700 bg-slate-950/70 text-sm text-slate-500">
            运行后将在这里展示 dashboard.html
          </div>
        )}
      </div>
    </section>
  );
}

interface FieldProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  actionLabel?: string;
  onAction?: () => void;
}

function Field({ label, value, onChange, actionLabel, onAction }: FieldProps) {
  return (
    <label className="space-y-2">
      <span className="text-sm font-medium text-slate-300">{label}</span>
      <div className="flex gap-2">
        <input
          type="text"
          value={value}
          onChange={(event) => onChange(event.target.value)}
          className="w-full rounded-2xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100 outline-none transition focus:border-sky-400"
        />
        {actionLabel && onAction ? (
          <button
            type="button"
            onClick={onAction}
            className="shrink-0 rounded-2xl border border-slate-700 px-4 py-3 text-sm text-slate-200 transition hover:border-sky-400 hover:text-white"
          >
            {actionLabel}
          </button>
        ) : null}
      </div>
    </label>
  );
}

interface ToggleProps {
  label: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}

function Toggle({ label, checked, onChange }: ToggleProps) {
  return (
    <label className="flex items-center justify-between rounded-2xl border border-slate-800 bg-slate-950 px-4 py-3 text-sm text-slate-200">
      <span>{label}</span>
      <input type="checkbox" checked={checked} onChange={(event) => onChange(event.target.checked)} />
    </label>
  );
}
