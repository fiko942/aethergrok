export namespace grokrunner {
	
	export class DiscoveredToolCall {
	    id: string;
	    tool: string;
	    params?: Record<string, any>;
	    result?: string;
	    status: string;
	    startTime?: number;
	    endTime?: number;
	
	    static createFrom(source: any = {}) {
	        return new DiscoveredToolCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.tool = source["tool"];
	        this.params = source["params"];
	        this.result = source["result"];
	        this.status = source["status"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	    }
	}
	export class DiscoveredChatMessage {
	    id: string;
	    role: string;
	    content: string;
	    timestamp: number;
	    reasoningContent?: string;
	    toolCalls?: DiscoveredToolCall[];
	    tokens?: Record<string, any>;
	    status?: string;
	
	    static createFrom(source: any = {}) {
	        return new DiscoveredChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.timestamp = source["timestamp"];
	        this.reasoningContent = source["reasoningContent"];
	        this.toolCalls = this.convertValues(source["toolCalls"], DiscoveredToolCall);
	        this.tokens = source["tokens"];
	        this.status = source["status"];
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
	
	export class GrokSessionMetadata {
	    id: string;
	    workspaceId: string;
	    title: string;
	    path: string;
	    createdAt: number;
	    updatedAt: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new GrokSessionMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.title = source["title"];
	        this.path = source["path"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.status = source["status"];
	    }
	}
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
	
	export class SessionUsageStats {
	    sessionId: string;
	    usedTokens: number;
	    maxTokens: number;
	    lastTurnInput: number;
	    lastTurnOutput: number;
	    lastTurnCacheRead: number;
	    lastTurnReasoning: number;
	    lastTurnModelCalls: number;
	    totalInput: number;
	    totalOutput: number;
	    totalCacheRead: number;
	    turnCount: number;
	    primaryModelId: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionUsageStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.usedTokens = source["usedTokens"];
	        this.maxTokens = source["maxTokens"];
	        this.lastTurnInput = source["lastTurnInput"];
	        this.lastTurnOutput = source["lastTurnOutput"];
	        this.lastTurnCacheRead = source["lastTurnCacheRead"];
	        this.lastTurnReasoning = source["lastTurnReasoning"];
	        this.lastTurnModelCalls = source["lastTurnModelCalls"];
	        this.totalInput = source["totalInput"];
	        this.totalOutput = source["totalOutput"];
	        this.totalCacheRead = source["totalCacheRead"];
	        this.turnCount = source["turnCount"];
	        this.primaryModelId = source["primaryModelId"];
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
	    sizeBytes: number;
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
	        this.sizeBytes = source["sizeBytes"];
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

