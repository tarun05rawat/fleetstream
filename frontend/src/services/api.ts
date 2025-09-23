import axios from 'axios';
import { SensorEvent, Alert, ProcessParameter, Machine, EventStats, SystemHealth, AnomalyThresholds } from '../types';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: `${API_BASE_URL}/api`,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    console.log(`API Request: ${config.method?.toUpperCase()} ${config.url}`);
    return config;
  },
  (error) => {
    console.error('API Request Error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => {
    console.log(`API Response: ${response.status} ${response.config.url}`);
    return response;
  },
  (error) => {
    console.error('API Response Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface EventsResponse {
  events: SensorEvent[];
  pagination: {
    limit: number;
    offset: number;
    count: number;
  };
}

export interface StatsResponse {
  stats: EventStats;
  period: {
    since: string;
    duration: string;
  };
}

// Events API
export const eventsApi = {
  getEvents: async (params?: {
    limit?: number;
    offset?: number;
    machine_id?: string;
  }): Promise<EventsResponse> => {
    const response = await api.get('/events', { params });
    return response.data;
  },

  getEventStats: async (params?: {
    machine_id?: string;
    since?: string;
  }): Promise<StatsResponse> => {
    const response = await api.get('/events/stats', { params });
    return response.data;
  },
};

// Alerts API
export const alertsApi = {
  getAlerts: async (): Promise<{ alerts: Alert[]; count: number }> => {
    const response = await api.get('/alerts');
    return response.data;
  },

  acknowledgeAlert: async (alertId: number): Promise<{ message: string; alert_id: number }> => {
    const response = await api.put(`/alerts/${alertId}/acknowledge`);
    return response.data;
  },
};

// Process Parameters API
export const parametersApi = {
  getParameters: async (): Promise<{ parameters: ProcessParameter[]; count: number }> => {
    const response = await api.get('/parameters');
    return response.data;
  },

  updateParameter: async (
    parameterName: string,
    parameterValue: string
  ): Promise<{ message: string; parameter_name: string; parameter_value: string }> => {
    const response = await api.put('/parameters', {
      parameter_name: parameterName,
      parameter_value: parameterValue,
    });
    return response.data;
  },
};

// Machines API
export const machinesApi = {
  getMachines: async (): Promise<{ machines: Machine[]; count: number }> => {
    const response = await api.get('/machines');
    return response.data;
  },
};

// System Health API
export const systemApi = {
  getHealth: async (): Promise<SystemHealth> => {
    const response = await api.get('/system/health');
    return response.data;
  },
};

// Anomaly Detection API
export const anomalyApi = {
  getThresholds: async (): Promise<{ thresholds: AnomalyThresholds }> => {
    const response = await api.get('/anomaly/thresholds');
    return response.data;
  },

  updateThresholds: async (thresholds: AnomalyThresholds): Promise<{
    message: string;
    thresholds: AnomalyThresholds;
  }> => {
    const response = await api.put('/anomaly/thresholds', thresholds);
    return response.data;
  },
};

// Generic error handler
export const handleApiError = (error: any): string => {
  if (error.response?.data?.error) {
    return error.response.data.error;
  }
  if (error.response?.data?.details) {
    return error.response.data.details;
  }
  if (error.message) {
    return error.message;
  }
  return 'An unexpected error occurred';
};

export default api;