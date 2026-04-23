import { apiClient } from './client';
import type {
  DockerCheckResponse,
  ImageListResponse,
  ContainerListResponse,
  HealthResponse,
} from '../types/docker';

export const dockerApi = {
  async checkConnection(): Promise<DockerCheckResponse> {
    return apiClient.get<DockerCheckResponse>('/docker/check');
  },

  async listImages(): Promise<ImageListResponse> {
    return apiClient.get<ImageListResponse>('/docker/images');
  },

  async listContainers(): Promise<ContainerListResponse> {
    return apiClient.get<ContainerListResponse>('/docker/containers');
  },

  async healthCheck(): Promise<HealthResponse> {
    return apiClient.get<HealthResponse>('/health');
  },
};

export default dockerApi;
