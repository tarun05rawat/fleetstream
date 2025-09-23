package database

import (
	"backend/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps the database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

// InsertEvent inserts a new event into the database
func (db *DB) InsertEvent(event *models.SensorEvent) (*models.Event, error) {
	rawDataJSON, err := json.Marshal(event.AdditionalData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw data: %v", err)
	}

	query := `
		INSERT INTO events (timestamp, machine_id, sensor_type, conveyor_speed, temperature, robot_arm_angle, status, raw_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, timestamp, machine_id, sensor_type, conveyor_speed, temperature, robot_arm_angle, status, raw_data, created_at
	`

	var dbEvent models.Event
	var rawDataBytes []byte

	err = db.QueryRow(query, event.Timestamp, event.MachineID, event.EventType,
		event.ConveyorSpeed, event.Temperature, event.RobotArmAngle, event.Status, rawDataJSON).Scan(
		&dbEvent.ID, &dbEvent.Timestamp, &dbEvent.MachineID, &dbEvent.SensorType,
		&dbEvent.ConveyorSpeed, &dbEvent.Temperature, &dbEvent.RobotArmAngle,
		&dbEvent.Status, &rawDataBytes, &dbEvent.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert event: %v", err)
	}

	if err := json.Unmarshal(rawDataBytes, &dbEvent.RawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw data: %v", err)
	}

	return &dbEvent, nil
}

// GetRecentEvents retrieves recent events with pagination
func (db *DB) GetRecentEvents(limit, offset int, machineID string) ([]models.Event, error) {
	query := `
		SELECT id, timestamp, machine_id, sensor_type, conveyor_speed, temperature, robot_arm_angle, status, raw_data, created_at
		FROM events
		WHERE ($3 = '' OR machine_id = $3)
		ORDER BY timestamp DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := db.Query(query, limit, offset, machineID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %v", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		var rawDataBytes []byte

		err := rows.Scan(&event.ID, &event.Timestamp, &event.MachineID, &event.SensorType,
			&event.ConveyorSpeed, &event.Temperature, &event.RobotArmAngle,
			&event.Status, &rawDataBytes, &event.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %v", err)
		}

		if err := json.Unmarshal(rawDataBytes, &event.RawData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw data: %v", err)
		}

		events = append(events, event)
	}

	return events, nil
}

// GetEventStats retrieves aggregated event statistics
func (db *DB) GetEventStats(machineID string, since time.Time) (*models.EventStats, error) {
	query := `
		SELECT
			COUNT(*) as total_events,
			COUNT(*) FILTER (WHERE status = 'fault') as fault_events,
			COUNT(*) FILTER (WHERE status = 'warning') as warning_events,
			COALESCE(AVG(temperature), 0) as avg_temperature,
			COALESCE(AVG(conveyor_speed), 0) as avg_conveyor_speed,
			MAX(timestamp) as last_event_time
		FROM events
		WHERE ($1 = '' OR machine_id = $1) AND timestamp >= $2
	`

	var stats models.EventStats
	var lastEventTime sql.NullTime

	err := db.QueryRow(query, machineID, since).Scan(
		&stats.TotalEvents, &stats.FaultEvents, &stats.WarningEvents,
		&stats.AvgTemperature, &stats.AvgConveyorSpeed, &lastEventTime)

	if err != nil {
		return nil, fmt.Errorf("failed to get event stats: %v", err)
	}

	if lastEventTime.Valid {
		stats.LastEventTime = lastEventTime.Time
	}

	// Calculate uptime percentage
	if stats.TotalEvents > 0 {
		uptimeEvents := stats.TotalEvents - stats.FaultEvents
		stats.UptimePercent = float64(uptimeEvents) / float64(stats.TotalEvents) * 100
	}

	return &stats, nil
}

// InsertAlert inserts a new alert
func (db *DB) InsertAlert(alert *models.Alert) error {
	query := `
		INSERT INTO alerts (event_id, alert_type, severity, message)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(query, alert.EventID, alert.AlertType, alert.Severity, alert.Message)
	if err != nil {
		return fmt.Errorf("failed to insert alert: %v", err)
	}

	return nil
}

// GetUnacknowledgedAlerts retrieves unacknowledged alerts
func (db *DB) GetUnacknowledgedAlerts() ([]models.Alert, error) {
	query := `
		SELECT id, event_id, alert_type, severity, message, acknowledged, created_at, acknowledged_at
		FROM alerts
		WHERE acknowledged = false
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %v", err)
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		err := rows.Scan(&alert.ID, &alert.EventID, &alert.AlertType, &alert.Severity,
			&alert.Message, &alert.Acknowledged, &alert.CreatedAt, &alert.AcknowledgedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %v", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (db *DB) AcknowledgeAlert(alertID int) error {
	query := `
		UPDATE alerts
		SET acknowledged = true, acknowledged_at = NOW()
		WHERE id = $1
	`

	_, err := db.Exec(query, alertID)
	if err != nil {
		return fmt.Errorf("failed to acknowledge alert: %v", err)
	}

	return nil
}

// GetProcessParameters retrieves all process parameters
func (db *DB) GetProcessParameters() ([]models.ProcessParameter, error) {
	query := `
		SELECT id, parameter_name, parameter_value, parameter_type, description, updated_at
		FROM process_parameters
		ORDER BY parameter_name
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query process parameters: %v", err)
	}
	defer rows.Close()

	var params []models.ProcessParameter
	for rows.Next() {
		var param models.ProcessParameter
		err := rows.Scan(&param.ID, &param.ParameterName, &param.ParameterValue,
			&param.ParameterType, &param.Description, &param.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan process parameter: %v", err)
		}
		params = append(params, param)
	}

	return params, nil
}

// UpdateProcessParameter updates a process parameter
func (db *DB) UpdateProcessParameter(name, value string) error {
	query := `
		UPDATE process_parameters
		SET parameter_value = $2, updated_at = NOW()
		WHERE parameter_name = $1
	`

	_, err := db.Exec(query, name, value)
	if err != nil {
		return fmt.Errorf("failed to update process parameter: %v", err)
	}

	return nil
}

// GetMachines retrieves all machines
func (db *DB) GetMachines() ([]models.Machine, error) {
	query := `
		SELECT id, machine_id, machine_type, location, status, config, created_at, updated_at
		FROM machines
		ORDER BY machine_id
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query machines: %v", err)
	}
	defer rows.Close()

	var machines []models.Machine
	for rows.Next() {
		var machine models.Machine
		var configBytes []byte

		err := rows.Scan(&machine.ID, &machine.MachineID, &machine.MachineType,
			&machine.Location, &machine.Status, &configBytes,
			&machine.CreatedAt, &machine.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan machine: %v", err)
		}

		if err := json.Unmarshal(configBytes, &machine.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal machine config: %v", err)
		}

		machines = append(machines, machine)
	}

	return machines, nil
}