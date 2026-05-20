package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
)

const topicOrderCreated = "order.created"

var kafkaProducer sarama.SyncProducer

type OrderCreatedEvent struct {
	OrderID uint        `json:"order_id"`
	Items   []OrderItem `json:"items"`
}

func initKafkaProducer() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Retry.Backoff = 500 * time.Millisecond

	var err error
	for i := 0; i < 10; i++ {
		kafkaProducer, err = sarama.NewSyncProducer([]string{brokers}, cfg)
		if err == nil {
			log.Println("Kafka producer connected")
			return
		}
		log.Printf("Kafka not ready, retrying (%d/10): %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	log.Fatalf("Failed to connect Kafka producer: %v", err)
}

func publishOrderCreated(order Order, items []OrderItem) {
	event := OrderCreatedEvent{OrderID: order.ID, Items: items}
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal order.created event: %v", err)
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: topicOrderCreated,
		Value: sarama.ByteEncoder(data),
	}

	if _, _, err := kafkaProducer.SendMessage(msg); err != nil {
		log.Printf("Failed to publish order.created for order %d: %v", order.ID, err)
	} else {
		log.Printf("Published order.created for order %d", order.ID)
	}
}
