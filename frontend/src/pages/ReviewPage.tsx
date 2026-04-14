import { useState } from "react";
import ReactMarkdown from "react-markdown";
import { runReview, selectDirectory } from "../lib/wails";
import type { Config, ReviewParams, ReviewResult } from "../types";

interface ReviewPageProps {
  config: Config;
}

const initialParams: ReviewParams = {
  repo: ".",
  base: "origin/develop",
  head: "HEAD",
  out_dir: "./deltascope-reports",
  api_key: "",
  api_base: "",
  model: "",
};

export default function ReviewPage({ config }: ReviewPageProps) {
  const [params, setParams] = useState<ReviewParams>(initialParams);
  const [running, setRunning] = useState(false);
  const [message, setMessage] = useState("等待开始评审");
  const [result, setResult] = useState<ReviewResult | null>(null);

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
    setMessage("正在执行 Review，请稍候...");
    try {
      const nextResult = await runReview({
        ...params,
        api_key: params.api_key || config.api_key,
        api_base: params.api_base || config.api_base,
        model: params.model || config.model,
      });
      setResult(nextResult);
      setMessage("Review 执行完成");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Review 执行失败");
    } finally {
      setRunning(false);
    }
  }

  return (
    <section className="grid gap-6 xl:grid-cols-[420px_minmax(0,1fr)]">
      <div className="space-y-4 rounded-3xl border border-slate-800 bg-slate-900/70 p-6">
        <div>
          <h2 className="text-2xl font-semibold text-slate-50">Review</h2>
          <p className="mt-2 text-sm text-slate-400">
            对比两个分支或提交的差异，调用已保存的模型配置生成影响分析和测试建议。
          </p>
        </div>

        <Field label="仓库路径" value={params.repo} onChange={(value) => setParams({ ...params, repo: value })} actionLabel="选择目录" onAction={chooseRepo} />
        <Field label="Base" value={params.base} onChange={(value) => setParams({ ...params, base: value })} />
        <Field label="Head" value={params.head} onChange={(value) => setParams({ ...params, head: value })} />
        <Field label="输出目录" value={params.out_dir} onChange={(value) => setParams({ ...params, out_dir: value })} actionLabel="选择目录" onAction={chooseOutDir} />

        <details className="rounded-2xl border border-slate-800 bg-slate-950/70 p-4 text-sm text-slate-300">
          <summary className="cursor-pointer font-medium text-slate-200">覆盖默认模型配置</summary>
          <div className="mt-4 space-y-4">
            <Field label="API Key" value={params.api_key} onChange={(value) => setParams({ ...params, api_key: value })} />
            <Field label="API Base URL" value={params.api_base} onChange={(value) => setParams({ ...params, api_base: value })} />
            <Field label="Model" value={params.model} onChange={(value) => setParams({ ...params, model: value })} />
          </div>
        </details>

        <button
          type="button"
          onClick={handleRun}
          disabled={running}
          className="w-full rounded-2xl bg-amber-300 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-amber-200 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {running ? "评审中..." : "运行评审"}
        </button>

        <div className="rounded-2xl border border-slate-800 bg-slate-950/80 px-4 py-3 text-sm text-slate-300">
          {message}
        </div>
      </div>

      <div className="rounded-3xl border border-slate-800 bg-slate-900/70 p-6">
        <div className="mb-4 flex flex-wrap items-center gap-3 text-xs text-slate-400">
          <span>输出目录：{result?.output_dir ?? "-"}</span>
          <span>文件：{result?.review_path ?? "-"}</span>
        </div>

        {result?.review_markdown ? (
          <article className="prose prose-invert max-w-none prose-headings:text-slate-100 prose-p:text-slate-300 prose-strong:text-slate-100 prose-code:text-sky-300">
            <ReactMarkdown>{result.review_markdown}</ReactMarkdown>
          </article>
        ) : (
          <div className="flex min-h-[720px] items-center justify-center rounded-2xl border border-dashed border-slate-700 bg-slate-950/70 text-sm text-slate-500">
            运行后将在这里展示 review.md
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
