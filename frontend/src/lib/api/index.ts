import { apiClient } from './axios';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken?: string;
  user: {
    id: string;
    email: string;
    name?: string;
  };
}

export async function login(payload: LoginRequest): Promise<LoginResponse> {
  const response = await apiClient.post<LoginResponse>('/auth/login', payload);
  return response.data;
}

export async function getProfile() {
  const response = await apiClient.get('/auth/me');
  return response.data;
}
