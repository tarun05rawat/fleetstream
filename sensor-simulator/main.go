package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/joho/godotenv"
)

// SensorEvent represents a sensor reading event
type SensorEvent struct {
	Timestamp       time.Time `json:"timestamp"`
	MachineID       string    `json:"machine_id"`
	ConveyorSpeed   float64   `json:"conveyor_speed"`
	Temperature     float64   `json:"temperature"`
	RobotArmAngle   float64   `json:"robot_arm_angle"`
	Status          string    `json:"status"`
	EventType       string    `json:"event_type"`
	AdditionalData  map[string]interface{} `json:"additional_data,omitempty"`
}

// SensorSimulator handles sensor data generation and publishing
type SensorSimulator struct {
	producer        *kafka.Producer
	topic           string
	frequency       time.Duration
	machineID       string
	faultRate       float64
	conveyorSpeed   float64
	temperature     float64
	robotArmAngle   float64
}

// NewSensorSimulator creates a new sensor simulator instance
func NewSensorSimulator(brokers, topic, machineID string, frequency time.Duration) (*SensorSimulator, error) {
	config := kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"client.id":         fmt.Sprintf("sensor-simulator-%s", machineID),
		"acks":              "all",
		"retries":           3,
		"retry.backoff.ms":  100,
	}

	producer, err := kafka.NewProducer(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	return &SensorSimulator{
		producer:      producer,
		topic:         topic,
		frequency:     frequency,
		machineID:     machineID,
		faultRate:     0.02, // 2% fault probability
		conveyorSpeed: 1.5,  // Initial speed
		temperature:   72.0, // Initial temperature
		robotArmAngle: 90.0, // Initial angle
	}, nil
}

// generateSensorEvent creates a realistic sensor event with potential faults
func (s *SensorSimulator) generateSensorEvent() *SensorEvent {
	now := time.Now()

	// Add some realistic variation to sensor readings
	s.conveyorSpeed += (rand.Float64() - 0.5) * 0.2 // ±0.1 variation
	s.temperature += (rand.Float64() - 0.5) * 2.0   // ±1.0 variation
	s.robotArmAngle += (rand.Float64() - 0.5) * 10.0 // ±5.0 variation

	// Keep values within realistic bounds
	s.conveyorSpeed = clamp(s.conveyorSpeed, 0.5, 3.0)
	s.temperature = clamp(s.temperature, 20.0, 80.0)
	s.robotArmAngle = clamp(s.robotArmAngle, 0.0, 180.0)

	status := "ok"
	eventType := "normal"
	additionalData := make(map[string]interface{})

	// Simulate faults
	if rand.Float64() < s.faultRate {
		faultType := rand.Intn(4)
		switch faultType {
		case 0: // Conveyor jam
			s.conveyorSpeed = 0.0
			status = "fault"
			eventType = "conveyor_jam"
			additionalData["fault_code"] = "CONV_JAM_001"
			additionalData["description"] = "Conveyor belt jammed"
		case 1: // Overheating
			s.temperature = 95.0 + rand.Float64()*10.0
			status = "fault"
			eventType = "overheat"
			additionalData["fault_code"] = "TEMP_HIGH_001"
			additionalData["description"] = "Temperature exceeds safe operating limits"
		case 2: // Robot arm stuck
			status = "fault"
			eventType = "robot_fault"
			additionalData["fault_code"] = "ROBOT_STUCK_001"
			additionalData["description"] = "Robot arm movement restricted"
		case 3: // Warning condition
			status = "warning"
			eventType = "maintenance_due"
			additionalData["warning_code"] = "MAINT_DUE_001"
			additionalData["description"] = "Scheduled maintenance approaching"
		}
	}

	// Add operational metrics
	additionalData["vibration_level"] = rand.Float64() * 0.5
	additionalData["power_consumption"] = 15.0 + rand.Float64()*5.0
	additionalData["cycle_count"] = rand.Intn(1000) + 5000

	return &SensorEvent{
		Timestamp:      now,
		MachineID:      s.machineID,
		ConveyorSpeed:  s.conveyorSpeed,
		Temperature:    s.temperature,
		RobotArmAngle:  s.robotArmAngle,
		Status:         status,
		EventType:      eventType,
		AdditionalData: additionalData,
	}
}

// publishEvent sends sensor event to Kafka
func (s *SensorSimulator) publishEvent(event *SensorEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &s.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(event.MachineID),
		Value: eventJSON,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "machine_id", Value: []byte(event.MachineID)},
			{Key: "status", Value: []byte(event.Status)},
		},
	}

	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	err = s.producer.Produce(message, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %v", err)
	}

	// Wait for delivery confirmation
	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return fmt.Errorf("delivery failed: %v", m.TopicPartition.Error)
	}

	log.Printf("Event delivered to topic %s [%d] at offset %v: %s",
		*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset, event.EventType)

	return nil
}

// Start begins the sensor simulation loop
func (s *SensorSimulator) Start() {
	log.Printf("Starting sensor simulator for machine %s, frequency: %v", s.machineID, s.frequency)

	ticker := time.NewTicker(s.frequency)
	defer ticker.Stop()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			event := s.generateSensorEvent()
			if err := s.publishEvent(event); err != nil {
				log.Printf("Error publishing event: %v", err)
			}
		case sig := <-sigChan:
			log.Printf("Received signal %v, shutting down...", sig)
			s.Close()
			return
		}
	}
}

// Close gracefully shuts down the simulator
func (s *SensorSimulator) Close() {
	log.Println("Closing sensor simulator...")
	s.producer.Flush(15 * 1000) // Wait up to 15 seconds
	s.producer.Close()
}

// clamp constrains a value between min and max
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Configuration
	brokers := getEnvOrDefault("KAFKA_BROKERS", "localhost:9092")
	topic := getEnvOrDefault("KAFKA_TOPIC", "line1.sensor")
	machineID := getEnvOrDefault("MACHINE_ID", "sensor_hub_001")

	frequencyMs := getEnvOrDefault("SENSOR_FREQUENCY", "100")
	frequency, err := strconv.Atoi(frequencyMs)
	if err != nil {
		log.Fatalf("Invalid sensor frequency: %v", err)
	}

	log.Printf("Configuration: brokers=%s, topic=%s, machine=%s, frequency=%dms",
		brokers, topic, machineID, frequency)

	// Create and start simulator
	simulator, err := NewSensorSimulator(brokers, topic, machineID, time.Duration(frequency)*time.Millisecond)
	if err != nil {
		log.Fatalf("Failed to create sensor simulator: %v", err)
	}

	// Start simulation
	simulator.Start()
}