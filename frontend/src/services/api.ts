import axios from "axios";
import {
  SensorEvent,
  Alert,
  ProcessParameter,
  Machine,
  EventStats,
  SystemHealth,
  AnomalyThresholds,
} from "../types";

const API_BASE_URL = process.env.REACT_APP_API_URL || "http://localhost:8080";

const api = axios.create({
  baseURL: `${API_BASE_URL}/api`,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    console.log(`API Request: ${config.method?.toUpperCase()} ${config.url}`);
    return config;
  },
  (error) => {
    console.error("API Request Error:", error);
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
    console.error("API Response Error:", error.response?.data || error.message);
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
    const response = await api.get("/events", { params });
    return {
      events: Array.isArray(response.data?.events) ? response.data.events : [],
      pagination: response.data?.pagination || {
        limit: 0,
        offset: 0,
        count: 0,
      },
    };
  },

  getEventStats: async (params?: {
    machine_id?: string;
    since?: string;
  }): Promise<StatsResponse> => {
    const response = await api.get("/events/stats", { params });
    return {
      stats: response.data?.stats || {
        total_events: 0,
        fault_events: 0,
        warning_events: 0,
        avg_temperature: 0,
        avg_conveyor_speed: 0,
        uptime_percent: 0,
        last_event_time: new Date().toISOString(),
      },
      period: response.data?.period || { since: "", duration: "" },
    };
  },
};

// Alerts API
export const alertsApi = {
  getAlerts: async (): Promise<{ alerts: Alert[]; count: number }> => {
    const response = await api.get("/alerts");
    return {
      alerts: Array.isArray(response.data?.alerts) ? response.data.alerts : [],
      count: response.data?.count || 0,
    };
  },

  acknowledgeAlert: async (
    alertId: number
  ): Promise<{ message: string; alert_id: number }> => {
    const response = await api.put(`/alerts/${alertId}/acknowledge`);
    return (
      response.data || { message: "Alert acknowledged", alert_id: alertId }
    );
  },
};

// Process Parameters API
export const parametersApi = {
  getParameters: async (): Promise<{
    parameters: ProcessParameter[];
    count: number;
  }> => {
    const response = await api.get("/parameters");
    return {
      parameters: Array.isArray(response.data?.parameters)
        ? response.data.parameters
        : [],
      count: response.data?.count || 0,
    };
  },

  updateParameter: async (
    parameterName: string,
    parameterValue: string
  ): Promise<{
    message: string;
    parameter_name: string;
    parameter_value: string;
  }> => {
    const response = await api.put("/parameters", {
      parameter_name: parameterName,
      parameter_value: parameterValue,
    });
    return (
      response.data || {
        message: "Parameter updated",
        parameter_name: parameterName,
        parameter_value: parameterValue,
      }
    );
  },
};

// Machines API
export const machinesApi = {
  getMachines: async (): Promise<{ machines: Machine[]; count: number }> => {
    const response = await api.get("/machines");
    return {
      machines: Array.isArray(response.data?.machines)
        ? response.data.machines
        : [],
      count: response.data?.count || 0,
    };
  },
};

// System Health API
export const systemApi = {
  getHealth: async (): Promise<SystemHealth> => {
    const response = await api.get("/system/health");
    return (
      response.data || {
        status: "unhealthy",
        timestamp: new Date().toISOString(),
        websocket: { connected_clients: 0 },
        database: { status: "unknown" },
        recent_activity: {
          total_events: 0,
          fault_events: 0,
          warning_events: 0,
          avg_temperature: 0,
          avg_conveyor_speed: 0,
          uptime_percent: 0,
          last_event_time: new Date().toISOString(),
        },
        thresholds: {
          conveyor_speed_min: 0,
          conveyor_speed_max: 100,
          temperature_min: 0,
          temperature_max: 100,
          robot_angle_min: 0,
          robot_angle_max: 360,
        },
      }
    );
  },
};

// Anomaly Detection API
export const anomalyApi = {
  getThresholds: async (): Promise<{ thresholds: AnomalyThresholds }> => {
    const response = await api.get("/anomaly/thresholds");
    return {
      thresholds: response.data?.thresholds || {
        conveyor_speed_min: 0,
        conveyor_speed_max: 100,
        temperature_min: 0,
        temperature_max: 100,
        robot_angle_min: 0,
        robot_angle_max: 360,
      },
    };
  },

  updateThresholds: async (
    thresholds: AnomalyThresholds
  ): Promise<{
    message: string;
    thresholds: AnomalyThresholds;
  }> => {
    const response = await api.put("/anomaly/thresholds", thresholds);
    return response.data || { message: "Thresholds updated", thresholds };
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
  return "An unexpected error occurred";
};

export default api;
