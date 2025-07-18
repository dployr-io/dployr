export namespace models {
	
	export class Console {
	    terminal: any;
	    websocket: any;
	    fitAddon: any;
	    terminalElement: any;
	    sessionId: any;
	    status: string;
	    statusMessage: string;
	    errorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new Console(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.terminal = source["terminal"];
	        this.websocket = source["websocket"];
	        this.fitAddon = source["fitAddon"];
	        this.terminalElement = source["terminalElement"];
	        this.sessionId = source["sessionId"];
	        this.status = source["status"];
	        this.statusMessage = source["statusMessage"];
	        this.errorMessage = source["errorMessage"];
	    }
	}
	export class Deployment {
	    commitHash: string;
	    branch: string;
	    duration: number;
	    message: string;
	    // Go type: time
	    createdAt: any;
	    status?: string;
	
	    static createFrom(source: any = {}) {
	        return new Deployment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.commitHash = source["commitHash"];
	        this.branch = source["branch"];
	        this.duration = source["duration"];
	        this.message = source["message"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class Domain {
	    id: string;
	    subdomain: string;
	    provider: string;
	    auto_setup_available: boolean;
	    manual_records?: string;
	    verified: boolean;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Domain(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.subdomain = source["subdomain"];
	        this.provider = source["provider"];
	        this.auto_setup_available = source["auto_setup_available"];
	        this.manual_records = source["manual_records"];
	        this.verified = source["verified"];
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class JSON___dployr_io_pkg_models_Domain_ {
	    Data: Domain[];
	
	    static createFrom(source: any = {}) {
	        return new JSON___dployr_io_pkg_models_Domain_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Data = this.convertValues(source["Data"], Domain);
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
	export class JSON_interface____ {
	    Data: any;
	
	    static createFrom(source: any = {}) {
	        return new JSON_interface____(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Data = source["Data"];
	    }
	}
	export class LogEntry {
	    id: string;
	    host: string;
	    message: string;
	    status: string;
	    level: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.host = source["host"];
	        this.message = source["message"];
	        this.status = source["status"];
	        this.level = source["level"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	    logo?: string;
	    description?: string;
	    git_repo: string;
	    // Go type: JSON___dployr_io_pkg_models_Domain_
	    domains?: any;
	    environment?: JSON_interface____;
	    deployment_url?: string;
	    // Go type: time
	    last_deployed?: any;
	    status?: string;
	    host_configs?: JSON_interface____;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.logo = source["logo"];
	        this.description = source["description"];
	        this.git_repo = source["git_repo"];
	        this.domains = this.convertValues(source["domains"], null);
	        this.environment = this.convertValues(source["environment"], JSON_interface____);
	        this.deployment_url = source["deployment_url"];
	        this.last_deployed = this.convertValues(source["last_deployed"], null);
	        this.status = source["status"];
	        this.host_configs = this.convertValues(source["host_configs"], JSON_interface____);
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
	export class SshConnectResponse {
	    sessionId: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new SshConnectResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.status = source["status"];
	    }
	}
	export class User {
	    id: string;
	    name: string;
	    email: string;
	    avatar?: string;
	    role: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.avatar = source["avatar"];
	        this.role = source["role"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class WsMessage {
	    type: string;
	    data?: string;
	    cols?: number;
	    rows?: number;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new WsMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.data = source["data"];
	        this.cols = source["cols"];
	        this.rows = source["rows"];
	        this.message = source["message"];
	    }
	}

}

