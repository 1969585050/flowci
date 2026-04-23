import { apiClient } from './client';
import type {
  DeployRequest,
  DeployResponse,
  RollbackResponse,
  DeployListResponse,
} from '../types/deploy';

export const deployApi = {
  async createDeploy(request: DeployRequest): Promise<DeployResponse> {
    return apiClient.post<DeployResponse, DeployRequest>('/deploys', request);
  },

  async getDeployStatus(deployId: string): Promise<DeployResponse> {
    return apiClient.get<DeployResponse>(`/deploys/${deployId}`);
  },

  async rollbackDeploy(deployId: string): Promise<RollbackResponse> {
    return apiClient.post<RollbackResponse>(`/deploys/${deployId}/rollback`);
  },

  async listDeploys(): Promise<DeployListResponse> {
    return apiClient.get<DeployListResponse>('/deploys');
  },
};

export default deployApi;
