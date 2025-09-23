package kafka

import (
	"backend/models"
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	consumer     *kafka.Consumer
	eventChannel chan *models.SensorEvent
	errorChannel chan error
	stopChannel  chan bool
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers, groupID string, topics []string) (*Consumer, error) {
	config := kafka.ConfigMap{
		"bootstrap.servers":        brokers,
		"group.id":                 groupID,
		"auto.offset.reset":        "latest",
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  1000,
		"session.timeout.ms":       30000,
		"heartbeat.interval.ms":    3000,
		"max.poll.interval.ms":     300000,
		"fetch.min.bytes":          1,
		"fetch.wait.max.ms":        500,
	}

	consumer, err := kafka.NewConsumer(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %v", err)
	}

	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to subscribe to topics: %v", err)
	}

	return &Consumer{
		consumer:     consumer,
		eventChannel: make(chan *models.SensorEvent, 100),
		errorChannel: make(chan error, 10),
		stopChannel:  make(chan bool, 1),
	}, nil
}

// EventChannel returns the channel for receiving sensor events
func (c *Consumer) EventChannel() <-chan *models.SensorEvent {
	return c.eventChannel
}

// ErrorChannel returns the channel for receiving errors
func (c *Consumer) ErrorChannel() <-chan error {
	return c.errorChannel
}

// Start begins consuming messages
func (c *Consumer) Start() {
	log.Println("Starting Kafka consumer...")

	go func() {
		defer close(c.eventChannel)
		defer close(c.errorChannel)

		for {
			select {
			case <-c.stopChannel:
				log.Println("Stopping Kafka consumer...")
				return
			default:
				msg, err := c.consumer.ReadMessage(-1) // Block indefinitely
				if err != nil {
					kafkaErr, ok := err.(kafka.Error)
					if ok && kafkaErr.Code() == kafka.ErrTimedOut {
						continue // Timeout is normal, continue polling
					}
					select {
					case c.errorChannel <- fmt.Errorf("consumer error: %v", err):
					default:
						log.Printf("Error channel full, dropping error: %v", err)
					}
					continue
				}

				c.processMessage(msg)
			}
		}
	}()
}

// processMessage processes an incoming Kafka message
func (c *Consumer) processMessage(msg *kafka.Message) {
	log.Printf("Received message from topic %s [%d] at offset %v",
		*msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset)

	// Parse the sensor event
	var event models.SensorEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		select {
		case c.errorChannel <- fmt.Errorf("failed to unmarshal message: %v", err):
		default:
			log.Printf("Error channel full, dropping unmarshal error: %v", err)
		}
		return
	}

	// Validate the event
	if err := c.validateEvent(&event); err != nil {
		select {
		case c.errorChannel <- fmt.Errorf("invalid event: %v", err):
		default:
			log.Printf("Error channel full, dropping validation error: %v", err)
		}
		return
	}

	// Send event to processing channel
	select {
	case c.eventChannel <- &event:
		log.Printf("Event processed successfully: machine=%s, status=%s, type=%s",
			event.MachineID, event.Status, event.EventType)
	default:
		log.Printf("Event channel full, dropping event from machine %s", event.MachineID)
	}
}

// validateEvent validates a sensor event
func (c *Consumer) validateEvent(event *models.SensorEvent) error {
	if event.MachineID == "" {
		return fmt.Errorf("machine_id is required")
	}

	if event.Status == "" {
		return fmt.Errorf("status is required")
	}

	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	// Validate status values
	validStatuses := map[string]bool{
		"ok":      true,
		"warning": true,
		"fault":   true,
	}

	if !validStatuses[event.Status] {
		return fmt.Errorf("invalid status: %s", event.Status)
	}

	// Validate sensor values are within reasonable bounds
	if event.ConveyorSpeed < 0 || event.ConveyorSpeed > 10 {
		return fmt.Errorf("conveyor speed out of range: %f", event.ConveyorSpeed)
	}

	if event.Temperature < -50 || event.Temperature > 200 {
		return fmt.Errorf("temperature out of range: %f", event.Temperature)
	}

	if event.RobotArmAngle < 0 || event.RobotArmAngle > 360 {
		return fmt.Errorf("robot arm angle out of range: %f", event.RobotArmAngle)
	}

	return nil
}

// Stop gracefully stops the consumer
func (c *Consumer) Stop() error {
	log.Println("Stopping Kafka consumer...")

	select {
	case c.stopChannel <- true:
	default:
	}

	return c.consumer.Close()
}

// GetMetadata returns consumer metadata
func (c *Consumer) GetMetadata() (*kafka.Metadata, error) {
	return c.consumer.GetMetadata(nil, false, 5000)
}