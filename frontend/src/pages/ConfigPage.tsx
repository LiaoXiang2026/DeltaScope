import { useState } from "react";
import { saveConfig } from "../lib/wails";
import type { Config } from "../types";

interface ConfigPageProps {
  config: Config;
  onChange: (config: Config) => void;
}

export default function ConfigPage({ config, onChange }: ConfigPageProps) {
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  async function handleSave() {
    setSaving(true);
    setMessage("");
    try {
      await saveConfig(config);
      setMessage("配置已保存到本地。");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "保存配置失败");
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold text-slate-50">配置中心</h2>
        <p className="mt-2 text-sm text-slate-400">
          Review 页面默认会读取这里保存的 API 配置。配置保存在用户目录下的
          <span className="mx-1 rounded bg-slate-800 px-2 py-1 text-slate-200">~/.deltascope/config.json</span>
        </p>
      </div>

      <div className="grid gap-4 rounded-3xl border border-slate-800 bg-slate-900/70 p-6 shadow-2xl shadow-slate-950/40">
        <label className="space-y-2">
          <span className="text-sm font-medium text-slate-300">API Key</span>
          <input
            type="password"
            value={config.api_key}
            onChange={(event) => onChange({ ...config, api_key: event.target.value })}
            className="w-full rounded-2xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100 outline-none transition focus:border-sky-400"
          />
        </label>

        <label className="space-y-2">
          <span className="text-sm font-medium text-slate-300">API Base URL</span>
          <input
            type="text"
            value={config.api_base}
            onChange={(event) => onChange({ ...config, api_base: event.target.value })}
            className="w-full rounded-2xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100 outline-none transition focus:border-sky-400"
          />
        </label>

        <label className="space-y-2">
          <span className="text-sm font-medium text-slate-300">Model</span>
          <input
            type="text"
            value={config.model}
            onChange={(event) => onChange({ ...config, model: event.target.value })}
            className="w-full rounded-2xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100 outline-none transition focus:border-sky-400"
          />
        </label>

        <div className="flex items-center gap-3 pt-2">
          <button
            type="button"
            onClick={handleSave}
            disabled={saving}
            className="rounded-2xl bg-sky-500 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {saving ? "保存中..." : "保存配置"}
          </button>
          {message ? <span className="text-sm text-slate-300">{message}</span> : null}
        </div>
      </div>
    </section>
  );
}
