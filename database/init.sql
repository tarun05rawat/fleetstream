-- FactoryFlow Database Schema

-- Events table to store all sensor events
CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    machine_id VARCHAR(50) NOT NULL,
    sensor_type VARCHAR(50) NOT NULL,
    conveyor_speed DECIMAL(5,2),
    temperature DECIMAL(5,2),
    robot_arm_angle DECIMAL(5,2),
    status VARCHAR(20) NOT NULL DEFAULT 'ok',
    raw_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Alerts table for fault detection
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    event_id INTEGER REFERENCES events(id),
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    message TEXT NOT NULL,
    acknowledged BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acknowledged_at TIMESTAMPTZ
);

-- Process parameters table for dynamic control
CREATE TABLE IF NOT EXISTS process_parameters (
    id SERIAL PRIMARY KEY,
    parameter_name VARCHAR(100) NOT NULL UNIQUE,
    parameter_value TEXT NOT NULL,
    parameter_type VARCHAR(20) NOT NULL DEFAULT 'string',
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Machine configurations
CREATE TABLE IF NOT EXISTS machines (
    id SERIAL PRIMARY KEY,
    machine_id VARCHAR(50) NOT NULL UNIQUE,
    machine_type VARCHAR(50) NOT NULL,
    location VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    config JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_machine_id ON events(machine_id);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_alerts_acknowledged ON alerts(acknowledged);

-- Insert default process parameters
INSERT INTO process_parameters (parameter_name, parameter_value, parameter_type, description) VALUES
('conveyor_speed_min', '0.5', 'float', 'Minimum conveyor speed in m/s'),
('conveyor_speed_max', '3.0', 'float', 'Maximum conveyor speed in m/s'),
('temperature_min', '20.0', 'float', 'Minimum operating temperature in Celsius'),
('temperature_max', '80.0', 'float', 'Maximum operating temperature in Celsius'),
('sensor_frequency', '100', 'int', 'Sensor reading frequency in milliseconds'),
('fault_probability', '0.02', 'float', 'Probability of fault events (0-1)')
ON CONFLICT (parameter_name) DO NOTHING;

-- Insert default machines
INSERT INTO machines (machine_id, machine_type, location, config) VALUES
('conveyor_001', 'conveyor', 'Line 1 - Station A', '{"max_speed": 3.0, "length": 10.0}'),
('robot_arm_001', 'robot_arm', 'Line 1 - Station B', '{"max_angle": 180, "payload": 50}'),
('sensor_hub_001', 'sensor_hub', 'Line 1 - Central', '{"sensors": ["temperature", "speed", "position"]}')
ON CONFLICT (machine_id) DO NOTHING;