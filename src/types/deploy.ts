export interface DeployRequest {
  project_id: string;
  deployment_type: DeploymentType;
  image_tag: string;
  container_name: string;
  ports: PortMapping[];
  env_vars?: Record<string, string>;
  volumes?: string[];
  restart_policy?: RestartPolicy;
  replicas?: number;
}

export type DeploymentType = 'compose' | 'single' | 'kubernetes';

export type RestartPolicy = 'no' | 'always' | 'unless-stopped' | 'on-failure';

export interface PortMapping {
  host_port: number;
  container_port: number;
  protocol: 'tcp' | 'udp';
}

export interface DeployResponse {
  id: string;
  container_id: string;
  name: string;
  status: ContainerStatus;
  ports: PortMapping[];
  created_at: string;
}

export type ContainerStatus = 'pending' | 'running' | 'stopped' | 'failed' | 'rolling_update';

export interface RollbackResponse {
  id: string;
  status: 'rolled_back';
  previous_image_tag: string;
  created_at: string;
}

export interface DeployListResponse {
  containers: DeployResponse[];
  total: number;
}

export const DEFAULT_RESTART_POLICY: RestartPolicy = 'unless-stopped';

export function validateDeployRequest(req: DeployRequest): string[] {
  const errors: string[] = [];

  if (!req.project_id) {
    errors.push('project_id is required');
  }

  if (!req.image_tag) {
    errors.push('image_tag is required');
  }

  if (!req.container_name) {
    errors.push('container_name is required');
  }

  return errors;
}
