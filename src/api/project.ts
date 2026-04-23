import { apiClient } from './client';
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectListResponse,
} from '../types/project';

export const projectApi = {
  async listProjects(): Promise<ProjectListResponse> {
    return apiClient.get<ProjectListResponse>('/projects');
  },

  async createProject(request: CreateProjectRequest): Promise<Project> {
    return apiClient.post<Project, CreateProjectRequest>('/projects', request);
  },

  async getProject(projectId: string): Promise<Project> {
    return apiClient.get<Project>(`/projects/${projectId}`);
  },

  async updateProject(projectId: string, request: UpdateProjectRequest): Promise<Project> {
    return apiClient.put<Project, UpdateProjectRequest>(`/projects/${projectId}`, request);
  },

  async deleteProject(projectId: string): Promise<void> {
    return apiClient.delete<void>(`/projects/${projectId}`);
  },
};

export default projectApi;
