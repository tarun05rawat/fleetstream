package models

import (
	"time"
)

// Event represents a sensor event from the database
type Event struct {
	ID              int                    `json:"id" db:"id"`
	Timestamp       time.Time              `json:"timestamp" db:"timestamp"`
	MachineID       string                 `json:"machine_id" db:"machine_id"`
	SensorType      string                 `json:"sensor_type" db:"sensor_type"`
	ConveyorSpeed   *float64               `json:"conveyor_speed" db:"conveyor_speed"`
	Temperature     *float64               `json:"temperature" db:"temperature"`
	RobotArmAngle   *float64               `json:"robot_arm_angle" db:"robot_arm_angle"`
	Status          string                 `json:"status" db:"status"`
	RawData         map[string]interface{} `json:"raw_data" db:"raw_data"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// Alert represents an alert in the system
type Alert struct {
	ID              int       `json:"id" db:"id"`
	EventID         *int      `json:"event_id" db:"event_id"`
	AlertType       string    `json:"alert_type" db:"alert_type"`
	Severity        string    `json:"severity" db:"severity"`
	Message         string    `json:"message" db:"message"`
	Acknowledged    bool      `json:"acknowledged" db:"acknowledged"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at" db:"acknowledged_at"`
}

// ProcessParameter represents a configurable process parameter
type ProcessParameter struct {
	ID             int       `json:"id" db:"id"`
	ParameterName  string    `json:"parameter_name" db:"parameter_name"`
	ParameterValue string    `json:"parameter_value" db:"parameter_value"`
	ParameterType  string    `json:"parameter_type" db:"parameter_type"`
	Description    string    `json:"description" db:"description"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Machine represents a machine in the factory
type Machine struct {
	ID          int                    `json:"id" db:"id"`
	MachineID   string                 `json:"machine_id" db:"machine_id"`
	MachineType string                 `json:"machine_type" db:"machine_type"`
	Location    string                 `json:"location" db:"location"`
	Status      string                 `json:"status" db:"status"`
	Config      map[string]interface{} `json:"config" db:"config"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// SensorEvent represents incoming sensor data from Kafka
type SensorEvent struct {
	Timestamp      time.Time              `json:"timestamp"`
	MachineID      string                 `json:"machine_id"`
	ConveyorSpeed  float64                `json:"conveyor_speed"`
	Temperature    float64                `json:"temperature"`
	RobotArmAngle  float64                `json:"robot_arm_angle"`
	Status         string                 `json:"status"`
	EventType      string                 `json:"event_type"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// WebSocketMessage represents a message sent to WebSocket clients
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// AnomalyThresholds defines thresholds for anomaly detection
type AnomalyThresholds struct {
	ConveyorSpeedMin float64 `json:"conveyor_speed_min"`
	ConveyorSpeedMax float64 `json:"conveyor_speed_max"`
	TemperatureMin   float64 `json:"temperature_min"`
	TemperatureMax   float64 `json:"temperature_max"`
	RobotAngleMin    float64 `json:"robot_angle_min"`
	RobotAngleMax    float64 `json:"robot_angle_max"`
}

// EventStats represents aggregated event statistics
type EventStats struct {
	TotalEvents    int64   `json:"total_events"`
	FaultEvents    int64   `json:"fault_events"`
	WarningEvents  int64   `json:"warning_events"`
	AvgTemperature float64 `json:"avg_temperature"`
	AvgConveyorSpeed float64 `json:"avg_conveyor_speed"`
	UptimePercent  float64 `json:"uptime_percent"`
	LastEventTime  time.Time `json:"last_event_time"`
}