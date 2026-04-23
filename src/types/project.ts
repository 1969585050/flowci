export interface Project {
  id: string;
  name: string;
  path: string;
  language: string;
  build_config?: BuildConfig;
  deploy_config?: DeployConfig;
  created_at: string;
  updated_at: string;
}

export interface BuildConfig {
  dockerfile_path?: string;
  context_path?: string;
}

export interface DeployConfig {
  deployment_type?: string;
  container_name?: string;
}

export interface CreateProjectRequest {
  name: string;
  path: string;
  language: string;
  build_config?: BuildConfig;
  deploy_config?: DeployConfig;
}

export interface UpdateProjectRequest {
  name?: string;
  language?: string;
  build_config?: BuildConfig;
  deploy_config?: DeployConfig;
}

export interface ProjectListResponse {
  projects: Project[];
  total: number;
}

export function validateProjectRequest(req: CreateProjectRequest): string[] {
  const errors: string[] = [];

  if (!req.name || req.name.trim() === '') {
    errors.push('name is required');
  }

  if (!req.path || req.path.trim() === '') {
    errors.push('path is required');
  }

  if (!req.language || req.language.trim() === '') {
    errors.push('language is required');
  }

  return errors;
}
