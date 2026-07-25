export namespace boot {
	
	export class SystemStats {
	    arch: string;
	    distro: string;
	    kernel: string;
	    cpuUsage: number;
	    memoryUsage: number;
	    diskUsage: number;
	    temperature: number;
	    networkStatus: string;
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
	        this.temperature = source["temperature"];
	        this.networkStatus = source["networkStatus"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class BootData {
	    systemStats?: SystemStats;
	    bootedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new BootData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.systemStats = this.convertValues(source["systemStats"], SystemStats);
	        this.bootedAt = source["bootedAt"];
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

