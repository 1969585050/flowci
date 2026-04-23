export interface BuildRequest {
  project_id: string;
  language: string;
  context_path: string;
  dockerfile_path?: string;
  image_tags: string[];
  build_args?: Record<string, string>;
  no_cache?: boolean;
  pull_base_image?: boolean;
}

export interface BuildResponse {
  id: string;
  image_id: string;
  tags: string[];
  size: number;
  duration_ms: number;
  status: BuildStatus;
  logs?: string[];
  started_at: string;
  finished_at: string;
}

export type BuildStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled';

export interface BuildListResponse {
  builds: BuildResponse[];
  total: number;
}

export interface BuildLogsResponse {
  build_id: string;
  logs: string[];
}

export type ProgrammingLanguage =
  | 'java-maven'
  | 'java-gradle'
  | 'nodejs'
  | 'python'
  | 'go'
  | 'php'
  | 'ruby'
  | 'dotnet'
  | 'custom';

export const SUPPORTED_LANGUAGES: ProgrammingLanguage[] = [
  'java-maven',
  'java-gradle',
  'nodejs',
  'python',
  'go',
  'php',
  'ruby',
  'dotnet',
  'custom',
];

export function isValidLanguage(lang: string): lang is ProgrammingLanguage {
  return SUPPORTED_LANGUAGES.includes(lang as ProgrammingLanguage);
}
