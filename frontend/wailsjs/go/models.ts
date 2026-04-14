export namespace backend {
	
	export class Config {
	    api_key: string;
	    api_base: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.api_key = source["api_key"];
	        this.api_base = source["api_base"];
	        this.model = source["model"];
	    }
	}

}

export namespace main {
	
	export class AnalyzeParams {
	    repo: string;
	    since: string;
	    from: string;
	    to: string;
	    out_dir: string;
	    branch: string;
	    prefix: string;
	    generate_json: boolean;
	    generate_charts: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AnalyzeParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repo = source["repo"];
	        this.since = source["since"];
	        this.from = source["from"];
	        this.to = source["to"];
	        this.out_dir = source["out_dir"];
	        this.branch = source["branch"];
	        this.prefix = source["prefix"];
	        this.generate_json = source["generate_json"];
	        this.generate_charts = source["generate_charts"];
	    }
	}
	export class AnalyzeResult {
	    output_dir: string;
	    report_path: string;
	    csv_path: string;
	    json_path: string;
	    dashboard_path: string;
	    report_markdown: string;
	    dashboard_html: string;
	
	    static createFrom(source: any = {}) {
	        return new AnalyzeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.output_dir = source["output_dir"];
	        this.report_path = source["report_path"];
	        this.csv_path = source["csv_path"];
	        this.json_path = source["json_path"];
	        this.dashboard_path = source["dashboard_path"];
	        this.report_markdown = source["report_markdown"];
	        this.dashboard_html = source["dashboard_html"];
	    }
	}
	export class ReviewParams {
	    repo: string;
	    base: string;
	    head: string;
	    out_dir: string;
	    api_key: string;
	    api_base: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new ReviewParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repo = source["repo"];
	        this.base = source["base"];
	        this.head = source["head"];
	        this.out_dir = source["out_dir"];
	        this.api_key = source["api_key"];
	        this.api_base = source["api_base"];
	        this.model = source["model"];
	    }
	}
	export class ReviewResult {
	    output_dir: string;
	    review_path: string;
	    review_markdown: string;
	
	    static createFrom(source: any = {}) {
	        return new ReviewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.output_dir = source["output_dir"];
	        this.review_path = source["review_path"];
	        this.review_markdown = source["review_markdown"];
	    }
	}

}

