export interface SensorEvent {
  id?: number;
  timestamp: string;
  machine_id: string;
  sensor_type?: string;
  conveyor_speed: number;
  temperature: number;
  robot_arm_angle: number;
  status: 'ok' | 'warning' | 'fault';
  event_type?: string;
  raw_data?: Record<string, any>;
  additional_data?: Record<string, any>;
  created_at?: string;
}

export interface Alert {
  id: number;
  event_id?: number;
  alert_type: string;
  severity: 'low' | 'medium' | 'high';
  message: string;
  acknowledged: boolean;
  created_at: string;
  acknowledged_at?: string;
}

export interface ProcessParameter {
  id: number;
  parameter_name: string;
  parameter_value: string;
  parameter_type: 'string' | 'int' | 'float' | 'boolean';
  description: string;
  updated_at: string;
}

export interface Machine {
  id: number;
  machine_id: string;
  machine_type: string;
  location: string;
  status: 'active' | 'inactive' | 'maintenance';
  config: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface EventStats {
  total_events: number;
  fault_events: number;
  warning_events: number;
  avg_temperature: number;
  avg_conveyor_speed: number;
  uptime_percent: number;
  last_event_time: string;
}

export interface AnomalyThresholds {
  conveyor_speed_min: number;
  conveyor_speed_max: number;
  temperature_min: number;
  temperature_max: number;
  robot_angle_min: number;
  robot_angle_max: number;
}

export interface SystemHealth {
  status: 'healthy' | 'degraded' | 'unhealthy';
  timestamp: string;
  websocket: {
    connected_clients: number;
  };
  database: {
    status: string;
  };
  recent_activity: EventStats;
  thresholds: AnomalyThresholds;
}

export interface WebSocketMessage {
  type: 'connection' | 'sensor_event' | 'alert' | 'stats' | 'pong';
  data: any;
  timestamp: string;
}

export interface ChartDataPoint {
  timestamp: string;
  conveyor_speed: number;
  temperature: number;
  robot_arm_angle: number;
  status: string;
}