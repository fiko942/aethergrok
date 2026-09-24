export namespace grokrunner {
	
	export class ModelInfo {
	    id: string;
	    name: string;
	    description: string;
	    isDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.isDefault = source["isDefault"];
	    }
	}
	export class PermissionResponse {
	    sessionId: string;
	    requestId: string;
	    decision: string;
	
	    static createFrom(source: any = {}) {
	        return new PermissionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.requestId = source["requestId"];
	        this.decision = source["decision"];
	    }
	}
	export class SessionOptions {
	    model?: string;
	    reasoningEffort?: string;
	    workingDir?: string;
	    skillDirs?: string[];
	    temperature?: number;
	    disableTools?: boolean;
	    systemPrompt?: string;
	    customFlags?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SessionOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.reasoningEffort = source["reasoningEffort"];
	        this.workingDir = source["workingDir"];
	        this.skillDirs = source["skillDirs"];
	        this.temperature = source["temperature"];
	        this.disableTools = source["disableTools"];
	        this.systemPrompt = source["systemPrompt"];
	        this.customFlags = source["customFlags"];
	    }
	}
	export class PromptRequest {
	    sessionId: string;
	    prompt: string;
	    images?: string[];
	    options?: SessionOptions;
	
	    static createFrom(source: any = {}) {
	        return new PromptRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.prompt = source["prompt"];
	        this.images = source["images"];
	        this.options = this.convertValues(source["options"], SessionOptions);
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

}

export namespace screen {
	
	export class SnapshotResult {
	    filePath: string;
	    dataUrl: string;
	    base64: string;
	    width: number;
	    height: number;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new SnapshotResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.dataUrl = source["dataUrl"];
	        this.base64 = source["base64"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.timestamp = source["timestamp"];
	    }
	}

}

export namespace skills {
	
	export class Skill {
	    id: string;
	    name: string;
	    description: string;
	    category: string;
	    tags: string[];
	    actions: string[];
	    path: string;
	    directory: string;
	    scope: string;
	    prompt?: string;
	
	    static createFrom(source: any = {}) {
	        return new Skill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.tags = source["tags"];
	        this.actions = source["actions"];
	        this.path = source["path"];
	        this.directory = source["directory"];
	        this.scope = source["scope"];
	        this.prompt = source["prompt"];
	    }
	}

}

