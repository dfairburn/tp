export namespace main {
	
	export class ConfigInfo {
	    configPath: string;
	    templatesDir: string;
	    environmentFile: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.configPath = source["configPath"];
	        this.templatesDir = source["templatesDir"];
	        this.environmentFile = source["environmentFile"];
	    }
	}
	export class HTTPResponse {
	    statusCode: number;
	    status: string;
	    headers: Record<string, string>;
	    body: string;
	    contentType: string;
	    duration: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new HTTPResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.status = source["status"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.contentType = source["contentType"];
	        this.duration = source["duration"];
	        this.error = source["error"];
	    }
	}
	export class TemplateItem {
	    name: string;
	    absolutePath: string;
	    relativePath: string;
	    method: string;
	    url: string;
	    headers: Record<string, string>;
	    body: string;
	    descriptions: Record<string, string>;
	    isDir: boolean;
	    depth: number;
	    children: TemplateItem[];
	
	    static createFrom(source: any = {}) {
	        return new TemplateItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.absolutePath = source["absolutePath"];
	        this.relativePath = source["relativePath"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.descriptions = source["descriptions"];
	        this.isDir = source["isDir"];
	        this.depth = source["depth"];
	        this.children = this.convertValues(source["children"], TemplateItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TemplateResponse {
	    name: string;
	    method: string;
	    url: string;
	    headers: Record<string, string>;
	    body: string;
	    descriptions: Record<string, string>;
	    rawContent: string;
	
	    static createFrom(source: any = {}) {
	        return new TemplateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.descriptions = source["descriptions"];
	        this.rawContent = source["rawContent"];
	    }
	}

}

