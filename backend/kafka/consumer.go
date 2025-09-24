package kafka

import (
	"backend/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	eventChannel  chan *models.SensorEvent
	errorChannel  chan error
	stopChannel   chan bool
	ctx           context.Context
	cancel        context.CancelFunc
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	eventChannel chan *models.SensorEvent
	errorChannel chan error
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers, groupID string, topics []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 20 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Version = sarama.V2_6_0_0

	brokerList := strings.Split(brokers, ",")
	consumerGroup, err := sarama.NewConsumerGroup(brokerList, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		consumerGroup: consumerGroup,
		eventChannel:  make(chan *models.SensorEvent, 100),
		errorChannel:  make(chan error, 10),
		stopChannel:   make(chan bool, 1),
		ctx:           ctx,
		cancel:        cancel,
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
func (c *Consumer) Start(topics []string) {
	log.Println("Starting Kafka consumer...")

	handler := &ConsumerGroupHandler{
		eventChannel: c.eventChannel,
		errorChannel: c.errorChannel,
	}

	go func() {
		defer close(c.eventChannel)
		defer close(c.errorChannel)

		for {
			select {
			case <-c.ctx.Done():
				log.Println("Consumer context cancelled")
				return
			case <-c.stopChannel:
				log.Println("Stopping Kafka consumer...")
				return
			default:
				err := c.consumerGroup.Consume(c.ctx, topics, handler)
				if err != nil {
					select {
					case c.errorChannel <- fmt.Errorf("consumer group error: %v", err):
					default:
						log.Printf("Error channel full, dropping error: %v", err)
					}
					continue
				}
			}
		}
	}()

	// Handle consumer group errors in a separate goroutine
	go func() {
		for err := range c.consumerGroup.Errors() {
			select {
			case c.errorChannel <- fmt.Errorf("consumer group error: %v", err):
			default:
				log.Printf("Error channel full, dropping error: %v", err)
			}
		}
	}()
}

// Stop gracefully stops the consumer
func (c *Consumer) Stop() error {
	log.Println("Stopping Kafka consumer...")

	select {
	case c.stopChannel <- true:
	default:
	}

	c.cancel()
	return c.consumerGroup.Close()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim starts a consumer loop of ConsumerGroupClaim's Messages()
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}
			h.processMessage(message)
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// processMessage processes an incoming Kafka message
func (h *ConsumerGroupHandler) processMessage(msg *sarama.ConsumerMessage) {
	log.Printf("Received message from topic %s [%d] at offset %v",
		msg.Topic, msg.Partition, msg.Offset)

	// Parse the sensor event
	var event models.SensorEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		select {
		case h.errorChannel <- fmt.Errorf("failed to unmarshal message: %v", err):
		default:
			log.Printf("Error channel full, dropping unmarshal error: %v", err)
		}
		return
	}

	// Validate the event
	if err := validateEvent(&event); err != nil {
		select {
		case h.errorChannel <- fmt.Errorf("invalid event: %v", err):
		default:
			log.Printf("Error channel full, dropping validation error: %v", err)
		}
		return
	}

	// Send event to processing channel
	select {
	case h.eventChannel <- &event:
		log.Printf("Event processed successfully: machine=%s, status=%s, type=%s",
			event.MachineID, event.Status, event.EventType)
	default:
		log.Printf("Event channel full, dropping event from machine %s", event.MachineID)
	}
}

// validateEvent validates a sensor event
func validateEvent(event *models.SensorEvent) error {
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