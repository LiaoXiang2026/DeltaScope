import { selectDirectory } from "../lib/wails";
import type { AnalyzeParams, AnalyzeResult } from "../types";
import { StatusPill, WorkbenchField, WorkbenchSection, WorkbenchToggle } from "../components/workbench";

interface AnalyzePageProps {
  params: AnalyzeParams;
  onParamsChange: (params: AnalyzeParams) => void;
  running: boolean;
  message: string;
  result: AnalyzeResult | null;
  onRun: () => void | Promise<void>;
}

export default function AnalyzePage({
  params,
  onParamsChange,
  running,
  message,
  result,
  onRun,
}: AnalyzePageProps) {

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
        <WorkbenchSection title="分析设置" description="选择仓库和时间范围，然后开始扫描。">
          <WorkbenchField
            label="仓库路径"
            value={params.repo}
            onChange={(value) => onParamsChange({ ...params, repo: value })}
            actionLabel="选择目录"
            onAction={chooseRepo}
          />
          <WorkbenchField
            label="相对时间"
            value={params.since}
            onChange={(value) => onParamsChange({ ...params, since: value })}
            placeholder="例如 3m / 30d"
          />

          <div className="field-grid field-grid--two">
            <WorkbenchField
              label="开始日期"
              value={params.from}
              onChange={(value) => onParamsChange({ ...params, from: value })}
              placeholder="YYYY-MM-DD"
            />
            <WorkbenchField
              label="结束日期"
              value={params.to}
              onChange={(value) => onParamsChange({ ...params, to: value })}
              placeholder="YYYY-MM-DD"
            />
          </div>

          <WorkbenchField
            label="输出目录"
            value={params.out_dir}
            onChange={(value) => onParamsChange({ ...params, out_dir: value })}
            actionLabel="选择目录"
            onAction={chooseOutDir}
          />

          <div className="field-grid field-grid--two">
            <WorkbenchField label="Hotfix 分支模式" value={params.branch} onChange={(value) => onParamsChange({ ...params, branch: value })} />
            <WorkbenchField label="缺陷前缀" value={params.prefix} onChange={(value) => onParamsChange({ ...params, prefix: value })} />
          </div>

          <div className="field-grid field-grid--two">
            <WorkbenchToggle
              label="生成 JSON"
              description="保留结构化结果。"
              checked={params.generate_json}
              onChange={(checked) => onParamsChange({ ...params, generate_json: checked })}
            />
            <WorkbenchToggle
              label="生成图表"
              description="输出 dashboard.html。"
              checked={params.generate_charts}
              onChange={(checked) => onParamsChange({ ...params, generate_charts: checked })}
            />
          </div>

          <div className="action-row">
            <button type="button" onClick={onRun} disabled={running} className="action-button action-button--primary">
              {running ? "分析中..." : "开始 Analyze"}
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
            <h2>结果</h2>
            <StatusPill tone={result?.dashboard_html ? "ready" : "neutral"}>
              {result?.dashboard_html ? "可预览" : "暂无结果"}
            </StatusPill>
          </div>

          <dl className="summary-rows summary-rows--compact">
            <div className="summary-row">
              <dt>输出目录</dt>
              <dd>{result?.output_dir ?? params.out_dir}</dd>
            </div>
            <div className="summary-row">
              <dt>Markdown</dt>
              <dd>{result?.report_path ?? "-"}</dd>
            </div>
            <div className="summary-row">
              <dt>Dashboard</dt>
              <dd>{result?.dashboard_path ?? "-"}</dd>
            </div>
          </dl>

          {result?.dashboard_html ? (
            <iframe title="Analyze Dashboard" srcDoc={result.dashboard_html} className="iframe-stage" />
          ) : (
            <div className="empty-state">
              <h3>还没有结果</h3>
              <p>执行一次 Analyze 后，这里会直接显示 dashboard。</p>
            </div>
          )}
        </section>
      </div>
    </section>
  );
}
