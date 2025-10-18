package main

import (
	"log"
	"os"
	"strings"

	"github.com/IBM/sarama"
)

func main() {
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Net.SASL.Enable = true
	config.Net.SASL.User = os.Getenv("KAFKA_USERNAME")
	config.Net.SASL.Password = os.Getenv("KAFKA_PASSWORD")
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.TLS.Enable = true
	config.Version = sarama.V2_8_0_0

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("❌ failed to create producer: %v", err)
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: os.Getenv("KAFKA_TOPIC"),
		Value: sarama.StringEncoder(`{"machine_id":"test_machine_local","temperature":80.5,"status":"ok"}`),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Fatalf("❌ failed to send message: %v", err)
	}
	log.Printf("✅ Test message sent to partition %d at offset %d", partition, offset)
}
