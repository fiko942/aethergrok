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
	
	export class GrokInstallStatus {
	    installed: boolean;
	    version: string;
	    binaryPath: string;
	    platform: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new GrokInstallStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.version = source["version"];
	        this.binaryPath = source["binaryPath"];
	        this.platform = source["platform"];
	        this.error = source["error"];
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
	    grokSessionId?: string;
	
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
	        this.grokSessionId = source["grokSessionId"];
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

export namespace logger {
	
	export class LogEntry {
	    id: string;
	    timestamp: number;
	    level: string;
	    category: string;
	    message: string;
	    details?: any;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.level = source["level"];
	        this.category = source["category"];
	        this.message = source["message"];
	        this.details = source["details"];
	    }
	}

}

export namespace main {
	
	export class VolumeMuteResult {
	    originalVolume: number;
	    wasMuted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VolumeMuteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.originalVolume = source["originalVolume"];
	        this.wasMuted = source["wasMuted"];
	    }
	}

}

export namespace permissions {
	
	export class SystemPermissionItem {
	    id: string;
	    title: string;
	    description: string;
	    granted: boolean;
	    message: string;
	    required: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SystemPermissionItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.granted = source["granted"];
	        this.message = source["message"];
	        this.required = source["required"];
	    }
	}
	export class AllPermissionsStatus {
	    platform: string;
	    allGranted: boolean;
	    items: SystemPermissionItem[];
	
	    static createFrom(source: any = {}) {
	        return new AllPermissionsStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.allGranted = source["allGranted"];
	        this.items = this.convertValues(source["items"], SystemPermissionItem);
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
	export class AudioDeviceInfo {
	    name: string;
	    isDefault: boolean;
	    transport: string;
	    manufacturer: string;
	
	    static createFrom(source: any = {}) {
	        return new AudioDeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.isDefault = source["isDefault"];
	        this.transport = source["transport"];
	        this.manufacturer = source["manufacturer"];
	    }
	}
	export class Status {
	    granted: boolean;
	    message: string;
	    platform: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.granted = source["granted"];
	        this.message = source["message"];
	        this.platform = source["platform"];
	    }
	}

}

export namespace screen {
	
	export class CacheStats {
	    totalBytes: number;
	    fileCount: number;
	    formattedSize: string;
	
	    static createFrom(source: any = {}) {
	        return new CacheStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalBytes = source["totalBytes"];
	        this.fileCount = source["fileCount"];
	        this.formattedSize = source["formattedSize"];
	    }
	}
	export class ClearCacheResult {
	    freedBytes: number;
	    deletedCount: number;
	    formattedSize: string;
	
	    static createFrom(source: any = {}) {
	        return new ClearCacheResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.freedBytes = source["freedBytes"];
	        this.deletedCount = source["deletedCount"];
	        this.formattedSize = source["formattedSize"];
	    }
	}
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
	
	export class DiscoveredSkill {
	    name: string;
	    description: string;
	    category: string;
	    tags: string[];
	    relativePath: string;
	    skillFile: string;
	    prereqs: string[];
	    commands: string[];
	
	    static createFrom(source: any = {}) {
	        return new DiscoveredSkill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.tags = source["tags"];
	        this.relativePath = source["relativePath"];
	        this.skillFile = source["skillFile"];
	        this.prereqs = source["prereqs"];
	        this.commands = source["commands"];
	    }
	}
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
	export class SkillAnalysisResult {
	    repoUrl: string;
	    repoName: string;
	    tempPath: string;
	    skills: DiscoveredSkill[];
	    globalPrereqs: string[];
	    suggestedScripts: string[];
	
	    static createFrom(source: any = {}) {
	        return new SkillAnalysisResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repoUrl = source["repoUrl"];
	        this.repoName = source["repoName"];
	        this.tempPath = source["tempPath"];
	        this.skills = this.convertValues(source["skills"], DiscoveredSkill);
	        this.globalPrereqs = source["globalPrereqs"];
	        this.suggestedScripts = source["suggestedScripts"];
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
	export class SkillInstallPayload {
	    tempPath: string;
	    skillPaths: string[];
	    targetScope: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillInstallPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tempPath = source["tempPath"];
	        this.skillPaths = source["skillPaths"];
	        this.targetScope = source["targetScope"];
	    }
	}
	export class SkillInstallResult {
	    success: boolean;
	    installedCount: number;
	    installedPaths: string[];
	    errors?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SkillInstallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.installedCount = source["installedCount"];
	        this.installedPaths = source["installedPaths"];
	        this.errors = source["errors"];
	    }
	}

}

export namespace storage {
	
	export class AppSettings {
	    theme: string;
	    defaultModel: string;
	    defaultReasoningEffort: string;
	    permissionMode: string;
	    planGateMode: string;
	    animationsEnabled: boolean;
	    grokBinaryPath: string;
	    snapshotShortcut: string;
	    snapshotDelayMs: number;
	    snapshotAutoHideWindow: boolean;
	    snapshotSoundEnabled: boolean;
	    snapshotFlashEnabled: boolean;
	    snapshotAutoAttach: boolean;
	    activeWindowTurnCount: number;
	    maxContextTokens: number;
	    sidebarWidth: number;
	    sidebarCollapsed: boolean;
	    selectedMicrophoneDeviceId: string;
	    dictationShortcut?: string;
	    dictationMuteSystemAudio?: boolean;
	    dictationHoldThresholdMs?: number;
	    updatedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.defaultModel = source["defaultModel"];
	        this.defaultReasoningEffort = source["defaultReasoningEffort"];
	        this.permissionMode = source["permissionMode"];
	        this.planGateMode = source["planGateMode"];
	        this.animationsEnabled = source["animationsEnabled"];
	        this.grokBinaryPath = source["grokBinaryPath"];
	        this.snapshotShortcut = source["snapshotShortcut"];
	        this.snapshotDelayMs = source["snapshotDelayMs"];
	        this.snapshotAutoHideWindow = source["snapshotAutoHideWindow"];
	        this.snapshotSoundEnabled = source["snapshotSoundEnabled"];
	        this.snapshotFlashEnabled = source["snapshotFlashEnabled"];
	        this.snapshotAutoAttach = source["snapshotAutoAttach"];
	        this.activeWindowTurnCount = source["activeWindowTurnCount"];
	        this.maxContextTokens = source["maxContextTokens"];
	        this.sidebarWidth = source["sidebarWidth"];
	        this.sidebarCollapsed = source["sidebarCollapsed"];
	        this.selectedMicrophoneDeviceId = source["selectedMicrophoneDeviceId"];
	        this.dictationShortcut = source["dictationShortcut"];
	        this.dictationMuteSystemAudio = source["dictationMuteSystemAudio"];
	        this.dictationHoldThresholdMs = source["dictationHoldThresholdMs"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class UIState {
	    activeWorkspaceId: string;
	    activeSessionId: string;
	    openTabSessionIds: string[];
	    updatedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new UIState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activeWorkspaceId = source["activeWorkspaceId"];
	        this.activeSessionId = source["activeSessionId"];
	        this.openTabSessionIds = source["openTabSessionIds"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class Workspace {
	    id: string;
	    name: string;
	    path: string;
	    createdAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Workspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.createdAt = source["createdAt"];
	    }
	}

}

export namespace updater {
	
	export class ReleaseAsset {
	    name: string;
	    size: number;
	    downloadUrl: string;
	    contentType: string;
	
	    static createFrom(source: any = {}) {
	        return new ReleaseAsset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.downloadUrl = source["downloadUrl"];
	        this.contentType = source["contentType"];
	    }
	}
	export class ReleaseInfo {
	    version: string;
	    tagName: string;
	    title: string;
	    publishedAt: string;
	    body: string;
	    highlights: string[];
	    assets: ReleaseAsset[];
	    isLatest: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ReleaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.tagName = source["tagName"];
	        this.title = source["title"];
	        this.publishedAt = source["publishedAt"];
	        this.body = source["body"];
	        this.highlights = source["highlights"];
	        this.assets = this.convertValues(source["assets"], ReleaseAsset);
	        this.isLatest = source["isLatest"];
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
	export class UpdateCheckResult {
	    updateAvailable: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    latestRelease?: ReleaseInfo;
	    allReleases: ReleaseInfo[];
	    platformAsset?: ReleaseAsset;
	    checkedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.updateAvailable = source["updateAvailable"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.latestRelease = this.convertValues(source["latestRelease"], ReleaseInfo);
	        this.allReleases = this.convertValues(source["allReleases"], ReleaseInfo);
	        this.platformAsset = this.convertValues(source["platformAsset"], ReleaseAsset);
	        this.checkedAt = source["checkedAt"];
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

export namespace workspace {
	
	export class FileCheckResult {
	    exists: boolean;
	    fullPath: string;
	    relPath: string;
	    isDir: boolean;
	    sizeBytes: number;
	
	    static createFrom(source: any = {}) {
	        return new FileCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exists = source["exists"];
	        this.fullPath = source["fullPath"];
	        this.relPath = source["relPath"];
	        this.isDir = source["isDir"];
	        this.sizeBytes = source["sizeBytes"];
	    }
	}
	export class FileItem {
	    name: string;
	    path: string;
	    fullPath: string;
	    isDir: boolean;
	    sizeBytes: number;
	    ext: string;
	
	    static createFrom(source: any = {}) {
	        return new FileItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.fullPath = source["fullPath"];
	        this.isDir = source["isDir"];
	        this.sizeBytes = source["sizeBytes"];
	        this.ext = source["ext"];
	    }
	}
	export class GitFileChange {
	    path: string;
	    status: string;
	    addedLines: number;
	    removedLines: number;
	
	    static createFrom(source: any = {}) {
	        return new GitFileChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.status = source["status"];
	        this.addedLines = source["addedLines"];
	        this.removedLines = source["removedLines"];
	    }
	}
	export class GitStatusResult {
	    branch: string;
	    isClean: boolean;
	    aheadCount: number;
	    behindCount: number;
	    changedFiles: GitFileChange[];
	    totalAdditions: number;
	    totalDeletions: number;
	
	    static createFrom(source: any = {}) {
	        return new GitStatusResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.branch = source["branch"];
	        this.isClean = source["isClean"];
	        this.aheadCount = source["aheadCount"];
	        this.behindCount = source["behindCount"];
	        this.changedFiles = this.convertValues(source["changedFiles"], GitFileChange);
	        this.totalAdditions = source["totalAdditions"];
	        this.totalDeletions = source["totalDeletions"];
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

