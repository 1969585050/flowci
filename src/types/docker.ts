export interface DockerCheckResponse {
  connected: boolean;
  version: string;
  api_version: string;
  os: string;
  arch: string;
}

export interface ImageInfo {
  id: string;
  tags: string[];
  size: number;
  created: string;
}

export interface ContainerInfo {
  id: string;
  names: string[];
  image: string;
  state: ContainerState;
  status: string;
  ports: PortInfo[];
  created: string;
}

export type ContainerState = 'running' | 'exited' | 'paused' | 'restarting' | 'removing' | 'dead';

export interface PortInfo {
  host_port: number;
  container_port: number;
  protocol: 'tcp' | 'udp';
}

export interface ImageListResponse {
  images: ImageInfo[];
  total: number;
}

export interface ContainerListResponse {
  containers: ContainerInfo[];
  total: number;
}

export interface HealthResponse {
  status: 'healthy' | 'degraded' | 'unhealthy';
  version: string;
  uptime_seconds: number;
  docker_connected: boolean;
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';

  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function formatContainerState(state: ContainerState): string {
  const stateMap: Record<ContainerState, string> = {
    running: '运行中',
    exited: '已停止',
    paused: '已暂停',
    restarting: '重启中',
    removing: '删除中',
    dead: '已终止',
  };

  return stateMap[state] || state;
}
