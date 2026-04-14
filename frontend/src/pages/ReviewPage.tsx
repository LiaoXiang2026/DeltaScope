import ReactMarkdown from "react-markdown";
import { selectDirectory } from "../lib/wails";
import type { Config, ReviewParams, ReviewResult } from "../types";
import { StatusPill, WorkbenchField, WorkbenchSection } from "../components/workbench";

interface ReviewPageProps {
  config: Config;
  params: ReviewParams;
  onParamsChange: (params: ReviewParams) => void;
  running: boolean;
  message: string;
  result: ReviewResult | null;
  onRun: () => void | Promise<void>;
}

export default function ReviewPage({ config, params, onParamsChange, running, message, result, onRun }: ReviewPageProps) {

  async function chooseRepo() {
    const dir = await selectDirectory();
    if (dir) {
      onParamsChange({ ...params, repo: dir });
    }
  }

  async function chooseOutDir() {
    const dir = await selectDirectory();
    if (dir) {
      onParamsChange({ ...params, out_dir: dir });
    }
  }

  return (
    <section className="workspace-layout workspace-layout--viewer">
      <aside className="workspace-side workspace-side--form">
        <WorkbenchSection title="评审设置" description="指定仓库、对比范围和输出目录。">
          <WorkbenchField
            label="仓库路径"
            value={params.repo}
            onChange={(value) => onParamsChange({ ...params, repo: value })}
            actionLabel="选择目录"
            onAction={chooseRepo}
          />

          <div className="field-grid field-grid--two">
            <WorkbenchField label="Base" value={params.base} onChange={(value) => onParamsChange({ ...params, base: value })} />
            <WorkbenchField label="Head" value={params.head} onChange={(value) => onParamsChange({ ...params, head: value })} />
          </div>

          <WorkbenchField
            label="输出目录"
            value={params.out_dir}
            onChange={(value) => onParamsChange({ ...params, out_dir: value })}
            actionLabel="选择目录"
            onAction={chooseOutDir}
          />

          <details className="config-disclosure">
            <summary>覆盖默认模型配置</summary>
            <div className="disclosure-body">
              <WorkbenchField
                label="API Key"
                type="password"
                value={params.api_key}
                onChange={(value) => onParamsChange({ ...params, api_key: value })}
              />
              <WorkbenchField
                label="API Base URL"
                value={params.api_base}
                onChange={(value) => onParamsChange({ ...params, api_base: value })}
              />
              <WorkbenchField label="Model" value={params.model} onChange={(value) => onParamsChange({ ...params, model: value })} />
            </div>
          </details>

          <div className="action-row">
            <button type="button" onClick={onRun} disabled={running} className="action-button action-button--primary">
              {running ? "评审中..." : "开始 Review"}
            </button>
            <div className="feedback-line">
              <StatusPill tone={running ? "running" : result ? "ready" : "neutral"}>
                {running ? "运行中" : result ? "已完成" : "待执行"}
              </StatusPill>
              <span>{message}</span>
            </div>
          </div>
        </WorkbenchSection>
      </aside>

      <div className="workspace-main workspace-main--viewer">
        <section className="surface-card viewer-panel viewer-panel--wide">
          <div className="side-panel-header">
            <h2>报告</h2>
            <StatusPill tone={result?.review_markdown ? "ready" : "neutral"}>
              {result?.review_markdown ? "可阅读" : "暂无结果"}
            </StatusPill>
          </div>

          <dl className="summary-rows summary-rows--compact">
            <div className="summary-row">
              <dt>输出目录</dt>
              <dd>{result?.output_dir ?? params.out_dir}</dd>
            </div>
            <div className="summary-row">
              <dt>文件</dt>
              <dd>{result?.review_path ?? "-"}</dd>
            </div>
            <div className="summary-row">
              <dt>默认模型</dt>
              <dd>{config.model || "未设置"}</dd>
            </div>
          </dl>

          {result?.review_markdown ? (
            <article className="report-markdown">
              <ReactMarkdown>{result.review_markdown}</ReactMarkdown>
            </article>
          ) : (
            <div className="empty-state">
              <h3>还没有报告</h3>
              <p>执行一次 Review 后，这里会直接显示 `review.md`。</p>
            </div>
          )}
        </section>
      </div>
    </section>
  );
}
