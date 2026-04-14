import { useState } from "react";
import { saveConfig } from "../lib/wails";
import type { Config } from "../types";
import { StatusPill, WorkbenchField, WorkbenchSection } from "../components/workbench";

interface ConfigPageProps {
  config: Config;
  onChange: (config: Config) => void;
}

export default function ConfigPage({ config, onChange }: ConfigPageProps) {
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const ready = Boolean(config.api_key && config.api_base && config.model);

  async function handleSave() {
    setSaving(true);
    setMessage("");
    try {
      await saveConfig(config);
      setMessage("已保存");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "保存失败");
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className="workspace-layout">
      <div className="workspace-main">
        <WorkbenchSection title="默认模型配置" description="保存后，Review 会自动使用这套配置。">
          <WorkbenchField
            label="API Key"
            type="password"
            value={config.api_key}
            onChange={(value) => onChange({ ...config, api_key: value })}
            placeholder="输入令牌"
          />
          <WorkbenchField
            label="API Base URL"
            value={config.api_base}
            onChange={(value) => onChange({ ...config, api_base: value })}
            placeholder="https://api.example.com/v1"
          />
          <WorkbenchField
            label="Model"
            value={config.model}
            onChange={(value) => onChange({ ...config, model: value })}
            placeholder="输入默认模型名称"
          />

          <div className="action-row">
            <button type="button" onClick={handleSave} disabled={saving} className="action-button action-button--primary">
              {saving ? "保存中..." : "保存配置"}
            </button>
            <div className="feedback-line">
              <StatusPill tone={message ? "ready" : "neutral"}>{message || "尚未保存"}</StatusPill>
              <span>配置文件会写入 `~/.deltascope/config.json`。</span>
            </div>
          </div>
        </WorkbenchSection>
      </div>

      <aside className="workspace-side">
        <section className="surface-card side-panel">
          <div className="side-panel-header">
            <h2>摘要</h2>
            <StatusPill tone={ready ? "ready" : "warning"}>{ready ? "已就绪" : "待完善"}</StatusPill>
          </div>

          <dl className="summary-rows">
            <div className="summary-row">
              <dt>保存位置</dt>
              <dd>~/.deltascope/config.json</dd>
            </div>
            <div className="summary-row">
              <dt>当前模型</dt>
              <dd>{config.model || "未设置"}</dd>
            </div>
            <div className="summary-row">
              <dt>API Base</dt>
              <dd>{config.api_base || "未设置"}</dd>
            </div>
            <div className="summary-row">
              <dt>配置状态</dt>
              <dd>{ready ? "可以直接执行 Review" : "建议先补全三项配置"}</dd>
            </div>
          </dl>
        </section>
      </aside>
    </section>
  );
}
