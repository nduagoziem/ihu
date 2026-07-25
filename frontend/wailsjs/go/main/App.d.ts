// Wails bindings for the App struct. Generated bindings extended for ihu.

export interface SystemStats {
  arch: string;
  distro: string;
  kernel: string;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  temperature: number;
  networkStatus: string;
  timestamp: string;
}

export interface BootData {
  systemStats: SystemStats;
  bootedAt: string;
}

export interface WSLConfig {
  welcomeDisabled: boolean;
  defaultLinuxPath: string;
  defaultLinuxUser: string;
  defaultLinuxDistro: string;
  pinnedFolders: string[];
  backgroundImage: string;
  backgroundMode: string;
}

export interface Entry {
  name: string;
  path: string;
  isDir: boolean;
  isSymlink: boolean;
  isHidden: boolean;
  size: number;
  modified: string;
  perm: string;
}

export function Greet(arg1: string): Promise<string>;
export function GetBootData(): Promise<BootData>;
export function GetConfig(): Promise<WSLConfig>;
export function SetWelcomeDisabled(arg1: boolean): Promise<WSLConfig>;
export function SetDefaultLinuxPath(arg1: string): Promise<WSLConfig>;
export function SetDefaultLinuxUser(arg1: string): Promise<WSLConfig>;
export function SetDefaultLinuxDistro(arg1: string): Promise<WSLConfig>;
export function TogglePinnedFolder(arg1: string): Promise<WSLConfig>;
export function SetBackground(arg1: string, arg2: string): Promise<WSLConfig>;
export function ListDir(arg1: string): Promise<Entry[]>;
export function HomePath(arg1: string): Promise<string>;
export function ListDistros(): Promise<string[]>;
export function ListUsers(): Promise<string[]>;
export function ReadFile(arg1: string): Promise<string>;
export function RunWSLCommand(arg1: string): Promise<string>;
export function ReadFileBase64(arg1: string): Promise<string>;
