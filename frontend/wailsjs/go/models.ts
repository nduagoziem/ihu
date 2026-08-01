export namespace config {
	
	export class WSLConfig {
	    defaultLinuxDistro: string;
	    pinnedFolders: string[];
	    backgroundImage: string;
	    backgroundMode: string;
	
	    static createFrom(source: any = {}) {
	        return new WSLConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultLinuxDistro = source["defaultLinuxDistro"];
	        this.pinnedFolders = source["pinnedFolders"];
	        this.backgroundImage = source["backgroundImage"];
	        this.backgroundMode = source["backgroundMode"];
	    }
	}

}

export namespace wsl {
	
	export class DefaultHome {
	    user: string;
	    home: string;
	
	    static createFrom(source: any = {}) {
	        return new DefaultHome(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user = source["user"];
	        this.home = source["home"];
	    }
	}
	export class Entry {
	    name: string;
	    path: string;
	    isDir: boolean;
	    isSymlink: boolean;
	    isHidden: boolean;
	    size: number;
	    modified: string;
	    perm: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.isSymlink = source["isSymlink"];
	        this.isHidden = source["isHidden"];
	        this.size = source["size"];
	        this.modified = source["modified"];
	        this.perm = source["perm"];
	    }
	}
	export class SystemStats {
	    arch: string;
	    distro: string;
	    kernel: string;
	    cpuUsage: number;
	    memoryUsage: number;
	    diskUsage: number;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.arch = source["arch"];
	        this.distro = source["distro"];
	        this.kernel = source["kernel"];
	        this.cpuUsage = source["cpuUsage"];
	        this.memoryUsage = source["memoryUsage"];
	        this.diskUsage = source["diskUsage"];
	        this.timestamp = source["timestamp"];
	    }
	}

}

