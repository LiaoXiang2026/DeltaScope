import type { AnalyzeParams, AnalyzeResult, Config, ReviewParams, ReviewResult } from "../types";

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          LoadConfig: () => Promise<Config>;
          SaveConfig: (cfg: Config) => Promise<void>;
          RunAnalyze: (params: AnalyzeParams) => Promise<AnalyzeResult>;
          RunReview: (params: ReviewParams) => Promise<ReviewResult>;
          SelectDirectory: () => Promise<string>;
        };
      };
    };
  }
}

async function waitForBridge() {
  for (let attempt = 0; attempt < 50; attempt += 1) {
    const app = window.go?.main?.App;
    if (app) {
      return app;
    }
    await new Promise((resolve) => window.setTimeout(resolve, 100));
  }
  throw new Error("Wails bridge is unavailable. Please start the app with Wails.");
}

export async function loadConfig() {
  const app = await waitForBridge();
  return app.LoadConfig();
}

export async function saveConfig(config: Config) {
  const app = await waitForBridge();
  return app.SaveConfig(config);
}

export async function runAnalyze(params: AnalyzeParams) {
  const app = await waitForBridge();
  return app.RunAnalyze(params);
}

export async function runReview(params: ReviewParams) {
  const app = await waitForBridge();
  return app.RunReview(params);
}

export async function selectDirectory() {
  const app = await waitForBridge();
  return app.SelectDirectory();
}
