export interface Config {
  api_key: string;
  api_base: string;
  model: string;
}

export interface AnalyzeParams {
  repo: string;
  since: string;
  from: string;
  to: string;
  out_dir: string;
  branch: string;
  prefix: string;
  generate_json: boolean;
  generate_charts: boolean;
}

export interface AnalyzeResult {
  output_dir: string;
  report_path: string;
  csv_path: string;
  json_path: string;
  dashboard_path: string;
  report_markdown: string;
  dashboard_html: string;
}

export interface ReviewParams {
  repo: string;
  base: string;
  head: string;
  out_dir: string;
  api_key: string;
  api_base: string;
  model: string;
}

export interface ReviewResult {
  output_dir: string;
  review_path: string;
  review_markdown: string;
}
