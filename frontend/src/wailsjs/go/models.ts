export namespace docker {
	
	export class BuildResult {
	    imageName: string;
	    imageTag: string;
	    log: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imageName = source["imageName"];
	        this.imageTag = source["imageTag"];
	        this.log = source["log"];
	    }
	}
	export class ComposeDeployResult {
	    output: string;
	
	    static createFrom(source: any = {}) {
	        return new ComposeDeployResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.output = source["output"];
	    }
	}
	export class Container {
	    id: string;
	    names: string[];
	    image: string;
	    state: string;
	    status: string;
	    ports: string;
	
	    static createFrom(source: any = {}) {
	        return new Container(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.names = source["names"];
	        this.image = source["image"];
	        this.state = source["state"];
	        this.status = source["status"];
	        this.ports = source["ports"];
	    }
	}
	export class DeployResult {
	    id: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DeployResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.message = source["message"];
	    }
	}
	export class EnvReport {
	    host: string;
	    connected: boolean;
	    clientVersion: string;
	    serverVersion: string;
	    serverOS: string;
	    serverArch: string;
	    hasBuildx: boolean;
	    hasCompose: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new EnvReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.connected = source["connected"];
	        this.clientVersion = source["clientVersion"];
	        this.serverVersion = source["serverVersion"];
	        this.serverOS = source["serverOS"];
	        this.serverArch = source["serverArch"];
	        this.hasBuildx = source["hasBuildx"];
	        this.hasCompose = source["hasCompose"];
	        this.message = source["message"];
	    }
	}
	export class Image {
	    id: string;
	    repository: string;
	    tag: string;
	    size: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Image(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.repository = source["repository"];
	        this.tag = source["tag"];
	        this.size = source["size"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class PushResult {
	    log: string;
	
	    static createFrom(source: any = {}) {
	        return new PushResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.log = source["log"];
	    }
	}
	export class Status {
	    connected: boolean;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.version = source["version"];
	    }
	}

}

export namespace git {
	
	export class InstallHint {
	    method: string;
	    label: string;
	    command: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallHint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.label = source["label"];
	        this.command = source["command"];
	        this.url = source["url"];
	    }
	}
	export class EnvReport {
	    installed: boolean;
	    version: string;
	    path: string;
	    message: string;
	    installHints: InstallHint[];
	
	    static createFrom(source: any = {}) {
	        return new EnvReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.version = source["version"];
	        this.path = source["path"];
	        this.message = source["message"];
	        this.installHints = this.convertValues(source["installHints"], InstallHint);
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

export namespace gitprovider {
	
	export class Repo {
	    name: string;
	    fullName: string;
	    cloneUrl: string;
	    htmlUrl: string;
	    defaultBranch: string;
	    description: string;
	    private: boolean;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Repo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.fullName = source["fullName"];
	        this.cloneUrl = source["cloneUrl"];
	        this.htmlUrl = source["htmlUrl"];
	        this.defaultBranch = source["defaultBranch"];
	        this.description = source["description"];
	        this.private = source["private"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class UserInfo {
	    username: string;
	    email: string;
	    avatarUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new UserInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.email = source["email"];
	        this.avatarUrl = source["avatarUrl"];
	    }
	}

}

export namespace handler {
	
	export class AIKeyStatus {
	    configured: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AIKeyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.configured = source["configured"];
	    }
	}
	export class BuildImageRequest {
	    projectId: string;
	    tag: string;
	    contextPath: string;
	    noCache: boolean;
	    pullLatest: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BuildImageRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.tag = source["tag"];
	        this.contextPath = source["contextPath"];
	        this.noCache = source["noCache"];
	        this.pullLatest = source["pullLatest"];
	    }
	}
	export class CreatePipelineRequest {
	    projectId: string;
	    name: string;
	    steps: store.PipelineStep[];
	    config: store.PipelineConfig;
	
	    static createFrom(source: any = {}) {
	        return new CreatePipelineRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.name = source["name"];
	        this.steps = this.convertValues(source["steps"], store.PipelineStep);
	        this.config = this.convertValues(source["config"], store.PipelineConfig);
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
	export class CreateProjectRequest {
	    name: string;
	    path: string;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateProjectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.language = source["language"];
	    }
	}
	export class DeployContainerRequest {
	    image: string;
	    name: string;
	    hostPort: string;
	    containerPort: string;
	    restartPolicy: string;
	    env: string;
	
	    static createFrom(source: any = {}) {
	        return new DeployContainerRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.image = source["image"];
	        this.name = source["name"];
	        this.hostPort = source["hostPort"];
	        this.containerPort = source["containerPort"];
	        this.restartPolicy = source["restartPolicy"];
	        this.env = source["env"];
	    }
	}
	export class DeployWithComposeRequest {
	    compose: string;
	    workdir: string;
	
	    static createFrom(source: any = {}) {
	        return new DeployWithComposeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.compose = source["compose"];
	        this.workdir = source["workdir"];
	    }
	}
	export class DetectDockerEnvRequest {
	    host: string;
	
	    static createFrom(source: any = {}) {
	        return new DetectDockerEnvRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	    }
	}
	export class DiagnoseBuildRequest {
	    buildId: string;
	
	    static createFrom(source: any = {}) {
	        return new DiagnoseBuildRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.buildId = source["buildId"];
	    }
	}
	export class DiagnoseBuildResponse {
	    markdown: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new DiagnoseBuildResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.markdown = source["markdown"];
	        this.model = source["model"];
	    }
	}
	export class ExecutePipelineRequest {
	    pipelineId: string;
	    projectId: string;
	
	    static createFrom(source: any = {}) {
	        return new ExecutePipelineRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pipelineId = source["pipelineId"];
	        this.projectId = source["projectId"];
	    }
	}
	export class GenerateComposeRequest {
	    image: string;
	    name: string;
	    hostPort: string;
	    containerPort: string;
	    restartPolicy: string;
	    env: string;
	
	    static createFrom(source: any = {}) {
	        return new GenerateComposeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.image = source["image"];
	        this.name = source["name"];
	        this.hostPort = source["hostPort"];
	        this.containerPort = source["containerPort"];
	        this.restartPolicy = source["restartPolicy"];
	        this.env = source["env"];
	    }
	}
	export class GiteaStatusResponse {
	    baseUrl: string;
	    hasToken: boolean;
	    tokenSettingsUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new GiteaStatusResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.baseUrl = source["baseUrl"];
	        this.hasToken = source["hasToken"];
	        this.tokenSettingsUrl = source["tokenSettingsUrl"];
	    }
	}
	export class ImportError {
	    fullName: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fullName = source["fullName"];
	        this.error = source["error"];
	    }
	}
	export class ImportGiteaRepo {
	    fullName: string;
	    cloneUrl: string;
	    branch?: string;
	    name?: string;
	    language?: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportGiteaRepo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fullName = source["fullName"];
	        this.cloneUrl = source["cloneUrl"];
	        this.branch = source["branch"];
	        this.name = source["name"];
	        this.language = source["language"];
	    }
	}
	export class ImportGiteaReposRequest {
	    repos: ImportGiteaRepo[];
	
	    static createFrom(source: any = {}) {
	        return new ImportGiteaReposRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repos = this.convertValues(source["repos"], ImportGiteaRepo);
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
	export class ImportGiteaReposResponse {
	    imported: store.Project[];
	    errors: ImportError[];
	
	    static createFrom(source: any = {}) {
	        return new ImportGiteaReposResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imported = this.convertValues(source["imported"], store.Project);
	        this.errors = this.convertValues(source["errors"], ImportError);
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
	export class ImportPipelineYamlRequest {
	    projectId: string;
	    yaml: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportPipelineYamlRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.yaml = source["yaml"];
	    }
	}
	export class Language {
	    language: string;
	    displayName: string;
	
	    static createFrom(source: any = {}) {
	        return new Language(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.displayName = source["displayName"];
	    }
	}
	export class PushImageRequest {
	    image: string;
	    registry: string;
	    username: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new PushImageRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.image = source["image"];
	        this.registry = source["registry"];
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}
	export class SaveAIKeyRequest {
	    apiKey: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveAIKeyRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiKey = source["apiKey"];
	    }
	}
	export class SaveGiteaConfigRequest {
	    baseUrl: string;
	    token: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveGiteaConfigRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.baseUrl = source["baseUrl"];
	        this.token = source["token"];
	    }
	}
	export class SaveSettingsRequest {
	    settings: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new SaveSettingsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.settings = source["settings"];
	    }
	}
	export class UpdatePipelineRequest {
	    id: string;
	    name: string;
	    steps: store.PipelineStep[];
	    config: store.PipelineConfig;
	
	    static createFrom(source: any = {}) {
	        return new UpdatePipelineRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.steps = this.convertValues(source["steps"], store.PipelineStep);
	        this.config = this.convertValues(source["config"], store.PipelineConfig);
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
	export class UpdateProjectRequest {
	    id: string;
	    name: string;
	    path: string;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateProjectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.language = source["language"];
	    }
	}

}

export namespace pipeline {
	
	export class StepLog {
	    step: string;
	    type: string;
	    status: string;
	    message?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new StepLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.step = source["step"];
	        this.type = source["type"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class ExecuteResult {
	    success: boolean;
	    logs: StepLog[];
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ExecuteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.logs = this.convertValues(source["logs"], StepLog);
	        this.message = source["message"];
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

export namespace store {
	
	export class BuildRecord {
	    id: string;
	    projectId: string;
	    imageName: string;
	    imageTag: string;
	    status: string;
	    log?: string;
	    logSize: number;
	    // Go type: time
	    startedAt: any;
	    // Go type: time
	    finishedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new BuildRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.imageName = source["imageName"];
	        this.imageTag = source["imageTag"];
	        this.status = source["status"];
	        this.log = source["log"];
	        this.logSize = source["logSize"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.finishedAt = this.convertValues(source["finishedAt"], null);
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
	export class PipelineConfig {
	    parallel: boolean;
	    stopOnFail: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PipelineConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parallel = source["parallel"];
	        this.stopOnFail = source["stopOnFail"];
	    }
	}
	export class PipelineStep {
	    type: string;
	    name: string;
	    config: Record<string, any>;
	    retry: number;
	    onFail: string;
	
	    static createFrom(source: any = {}) {
	        return new PipelineStep(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	        this.config = source["config"];
	        this.retry = source["retry"];
	        this.onFail = source["onFail"];
	    }
	}
	export class Pipeline {
	    id: string;
	    projectId: string;
	    name: string;
	    steps: PipelineStep[];
	    config: PipelineConfig;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Pipeline(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.name = source["name"];
	        this.steps = this.convertValues(source["steps"], PipelineStep);
	        this.config = this.convertValues(source["config"], PipelineConfig);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	
	
	export class Project {
	    id: string;
	    name: string;
	    path: string;
	    language: string;
	    repoUrl: string;
	    repoBranch: string;
	    // Go type: time
	    lastPullAt?: any;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.language = source["language"];
	        this.repoUrl = source["repoUrl"];
	        this.repoBranch = source["repoBranch"];
	        this.lastPullAt = this.convertValues(source["lastPullAt"], null);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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

