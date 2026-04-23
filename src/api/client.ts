import axios, { AxiosInstance, AxiosError } from 'axios';
import type { ApiResponse, FlowCIError } from '../types/api';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:3847/api/v1';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        if (error.response) {
          const data = error.response.data as ApiResponse<unknown>;
          throw new FlowCIError(
            data?.code || error.response.status,
            data?.message || error.message,
            data?.data
          );
        }
        throw new FlowCIError(4003, error.message || 'Network error');
      }
    );
  }

  async get<T>(path: string): Promise<T> {
    const response = await this.client.get<ApiResponse<T>>(path);
    if (response.data.code !== 0) {
      throw new FlowCIError(response.data.code, response.data.message, response.data.data);
    }
    return response.data.data as T;
  }

  async post<T, D = unknown>(path: string, data?: D): Promise<T> {
    const response = await this.client.post<ApiResponse<T>>(path, data);
    if (response.data.code !== 0) {
      throw new FlowCIError(response.data.code, response.data.message, response.data.data);
    }
    return response.data.data as T;
  }

  async put<T, D = unknown>(path: string, data?: D): Promise<T> {
    const response = await this.client.put<ApiResponse<T>>(path, data);
    if (response.data.code !== 0) {
      throw new FlowCIError(response.data.code, response.data.message, response.data.data);
    }
    return response.data.data as T;
  }

  async delete<T>(path: string): Promise<T> {
    const response = await this.client.delete<ApiResponse<T>>(path);
    if (response.data.code !== 0) {
      throw new FlowCIError(response.data.code, response.data.message, response.data.data);
    }
    return response.data.data as T;
  }
}

export const apiClient = new ApiClient();
export default apiClient;
