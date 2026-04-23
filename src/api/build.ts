import { apiClient } from './client';
import type {
  BuildRequest,
  BuildResponse,
  BuildListResponse,
  BuildLogsResponse,
} from '../types/build';

export const buildApi = {
  async createBuild(request: BuildRequest): Promise<BuildResponse> {
    return apiClient.post<BuildResponse, BuildRequest>('/builds', request);
  },

  async getBuild(buildId: string): Promise<BuildResponse> {
    return apiClient.get<BuildResponse>(`/builds/${buildId}`);
  },

  async getBuildLogs(buildId: string): Promise<BuildLogsResponse> {
    return apiClient.get<BuildLogsResponse>(`/builds/${buildId}/logs`);
  },

  async listBuilds(): Promise<BuildListResponse> {
    return apiClient.get<BuildListResponse>('/builds');
  },
};

export default buildApi;
